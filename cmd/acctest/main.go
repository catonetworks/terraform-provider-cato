// cmd/acctest runs Cato provider acceptance tests with concurrency control and
// automatic retry on transient API failures.
//
// Usage:
//
//	go run ./cmd/acctest [flags] [test-dir...]
//
// Flags:
//
//	--coverage          Enable coverage profiling
//	--nocolor           Disable color in tparse output
//	--suite <file>      File listing test directory names to run (one per line)
//	--max-parallel <n>  Max concurrent independent packages (default 6; ACCTEST_MAX_PARALLEL env sets default)
//	--retry-count <n>   Retries on transient API failures (default 3)
//
// Packages are split into two groups:
//
//	serial   — share the IF/WF/WAN-Network policy and must run sequentially.
//	           cleanup() is called before each retry in this group.
//	parallel — independent packages, up to --max-parallel run concurrently.
//
// Both groups run simultaneously; the serial stream is one goroutine, the
// parallel packages are bounded by a semaphore channel.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	outDir      = "tmp_recorded/output"
	coverageDir = "tmp_recorded/coverage"
)

// retryRe matches transient API errors that warrant a retry.
var retryRe = regexp.MustCompile(
	`internal server error|connection refused|rate limit|reorderPolicyBlockedByActiveSessions`,
)

// sharedPolicyPkgs share the IF/WF/WAN-Network policy revision and must run
// sequentially to avoid draft-revision conflicts. cleanup() is called before
// each retry in this stream.
var sharedPolicyPkgs = map[string]bool{
	"if_rule": true, "if_rules_index": true, "if_section": true,
	"wf_rule": true, "wf_rules_index": true, "wf_rules_index_with_rule_data": true, "wf_section": true,
	"wnw_rule": true, "wnw_rules_index": true, "wan_network_section": true,
}

// longTimeoutPkgs need more than the default 5-minute test timeout.
var longTimeoutPkgs = map[string]bool{
	"wf_rules_index": true, "wf_rules_index_with_rule_data": true,
}

// testEnv is injected into every `go test` subprocess.
var testEnv = map[string]string{
	"TF_ACC":                      "1",
	"DISABLE_POLICY_RULE_CLEANUP": "true",
	"TF_ACC_MOCK":                 "",
}

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

var (
	printMu     sync.Mutex
	colorOutput = true // disabled by --nocolor
)

func logf(format string, args ...any) {
	printMu.Lock()
	fmt.Printf(format+"\n", args...)
	printMu.Unlock()
}

func green(s string) string {
	if !colorOutput {
		return s
	}
	return colorGreen + s + colorReset
}

func red(s string) string {
	if !colorOutput {
		return s
	}
	return colorRed + s + colorReset
}

// buildEnv merges os.Environ() with overrides, deduplicating by key so that
// overridden keys appear exactly once at the end.
func buildEnv(overrides map[string]string) []string {
	base := os.Environ()
	result := make([]string, 0, len(base)+len(overrides))
	for _, e := range base {
		key, _, _ := strings.Cut(e, "=")
		if _, skip := overrides[key]; !skip {
			result = append(result, e)
		}
	}
	for k, v := range overrides {
		result = append(result, k+"="+v)
	}
	return result
}

func pkgTimeout(pkg string) string {
	if longTimeoutPkgs[pkg] {
		return "12m"
	}
	return "5m"
}

func runCleanup() {
	cmd := exec.Command("go", "test",
		"-tags", "acctest",
		"-count=1",
		"--timeout=3m",
		"-run", "TestCleanupAccTestResources",
		"./internal/acctests/acc",
	)
	cmd.Env = buildEnv(map[string]string{"ACCTEST_CLEANUP": "true"})
	if err := cmd.Run(); err != nil {
		logf("cleanup warning: %v", err)
	}
}

func runTest(tdir, logFile, coverFile string, coverage bool) error {
	pkg := filepath.Base(tdir)
	args := []string{
		"test",
		"-timeout", pkgTimeout(pkg),
		"-tags", "acctest",
		"-count=1",
		"-parallel=1",
		"-p=1",
		tdir,
		"-json",
	}
	if coverage {
		args = append(args,
			"-coverprofile="+coverFile,
			"-covermode=atomic",
			"-coverpkg=./...",
		)
	}

	cmd := exec.Command("go", args...)
	cmd.Env = buildEnv(testEnv)

	f, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("create log %s: %w", logFile, err)
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = f
	return cmd.Run()
}

func shouldRetry(logFile string) bool {
	data, err := os.ReadFile(logFile)
	if err != nil {
		return false
	}
	return retryRe.Match(data)
}

// runPkg runs one package with up to retries retries on transient errors.
// Returns true on success.
func runPkg(tdir string, needsCleanup, coverage bool, retries int) bool {
	pkg := filepath.Base(tdir)
	logFile := filepath.Join(outDir, pkg)
	coverFile := filepath.Join(coverageDir, pkg)

	logf("Starting:\t%s", pkg)
	if runTest(tdir, logFile, coverFile, coverage) == nil {
		logf("%s\t\t%s", green("OK:"), pkg)
		return true
	}
	logf("%s\t\t%s", red("ERROR:"), pkg)

	for r := 1; r <= retries; r++ {
		if !shouldRetry(logFile) {
			return false
		}
		time.Sleep(5 * time.Second)
		if needsCleanup {
			runCleanup()
		}
		logf("Retrying (%d/%d):\t%s", r, retries, pkg)
		if runTest(tdir, logFile, coverFile, coverage) == nil {
			logf("%s (retry %d):\t%s", green("OK"), r, pkg)
			return true
		}
		logf("%s (retry %d):\t%s", red("ERROR"), r, pkg)
	}
	return false
}

// discoverTestDirs finds all test packages under ./internal/acctests/, excluding
// the shared acc/ helper package. If filter is non-empty, only packages whose
// base name matches a filter entry are returned.
func discoverTestDirs(filter []string) ([]string, error) {
	seen := map[string]bool{}
	var all []string
	err := filepath.Walk("./internal/acctests/", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.HasSuffix(path, "_test.go") {
			dir := filepath.Dir(path)
			if !seen[dir] && filepath.Base(dir) != "acc" {
				seen[dir] = true
				all = append(all, "./"+dir)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(all)

	if len(filter) == 0 {
		return all, nil
	}
	filterSet := map[string]bool{}
	for _, f := range filter {
		filterSet[filepath.Base(strings.TrimSpace(f))] = true
	}
	var out []string
	for _, d := range all {
		if filterSet[filepath.Base(d)] {
			out = append(out, d)
		}
	}
	return out, nil
}

// mergeCoverage concatenates per-package coverage profiles into a single
// cover.out file in coverageDir.
func mergeCoverage() error {
	entries, err := os.ReadDir(coverageDir)
	if err != nil {
		return err
	}
	outFile, err := os.Create(filepath.Join(coverageDir, "cover.out"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, "mode: atomic")
	skip := map[string]bool{"cover.out": true, "cover.html": true, "cover-stats.txt": true}
	for _, e := range entries {
		if skip[e.Name()] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(coverageDir, e.Name()))
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(strings.NewReader(string(data)))
		for sc.Scan() {
			line := sc.Text()
			if line == "mode: atomic" || line == "" {
				continue
			}
			fmt.Fprintln(outFile, line)
		}
	}
	return nil
}

func computeCoverage() {
	if err := mergeCoverage(); err != nil {
		logf("coverage merge: %v", err)
		return
	}
	fmt.Println("\nTest Coverage\n~~~~~~~~~~~~~")
	cmd := exec.Command("go", "tool", "cover", "-func="+filepath.Join(coverageDir, "cover.out"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logf("coverage report: %v", err)
	}
}

func runTparse(nocolor bool) {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		logf("tparse: %v", err)
		return
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for _, e := range entries {
			data, err := os.ReadFile(filepath.Join(outDir, e.Name()))
			if err != nil {
				continue
			}
			pw.Write(data) //nolint:errcheck
		}
	}()

	args := []string{
		"tool", "tparse",
		"-trimpath", "github.com/catonetworks/terraform-provider-cato/",
		"--all",
	}
	if nocolor {
		args = append(args, "-nocolor")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdin = pr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() //nolint:errcheck
}

func main() {
	// Resolve ACCTEST_MAX_PARALLEL env before flag.Parse so an explicit
	// --max-parallel flag still wins.
	defaultMaxParallel := 6
	if v := os.Getenv("ACCTEST_MAX_PARALLEL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			defaultMaxParallel = n
		}
	}

	coverage := flag.Bool("coverage", false, "enable coverage profiling")
	nocolor := flag.Bool("nocolor", false, "disable color output")
	suiteFile := flag.String("suite", "", "file containing test directory names to run (one per line)")
	maxParallel := flag.Int("max-parallel", defaultMaxParallel,
		"max concurrent independent packages (ACCTEST_MAX_PARALLEL env sets default)")
	retries := flag.Int("retry-count", 3, "retries on transient API failures")
	flag.Parse()

	if *nocolor {
		colorOutput = false
	}

	// Collect filter list from --suite file and/or positional args.
	var filter []string
	if *suiteFile != "" {
		f, err := os.Open(*suiteFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "suite file: %v\n", err)
			os.Exit(1)
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if t := strings.TrimSpace(sc.Text()); t != "" {
				filter = append(filter, t)
			}
		}
		f.Close()
	}
	filter = append(filter, flag.Args()...)

	testDirs, err := discoverTestDirs(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover: %v\n", err)
		os.Exit(1)
	}
	if len(testDirs) == 0 {
		fmt.Fprintln(os.Stderr, "no tests selected")
		os.Exit(1)
	}

	os.RemoveAll(outDir)   //nolint:errcheck
	os.RemoveAll(coverageDir) //nolint:errcheck
	os.MkdirAll(outDir, 0o755)    //nolint:errcheck
	os.MkdirAll(coverageDir, 0o755) //nolint:errcheck

	fmt.Println("Running cleanup...")
	runCleanup()

	var serialDirs, parallelDirs []string
	for _, d := range testDirs {
		if sharedPolicyPkgs[filepath.Base(d)] {
			serialDirs = append(serialDirs, d)
		} else {
			parallelDirs = append(parallelDirs, d)
		}
	}

	fmt.Printf("Streams: %d independent (max %d concurrent) + 1 shared-policy (%d packages)\n",
		len(parallelDirs), *maxParallel, len(serialDirs))

	var failed atomic.Bool
	var wg sync.WaitGroup

	// Serial stream: shared-policy packages run one-at-a-time with cleanup on retry.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, d := range serialDirs {
			if !runPkg(d, true, *coverage, *retries) {
				failed.Store(true)
			}
		}
	}()

	// Parallel stream: independent packages bounded by a semaphore channel.
	// Acquire before spawning so at most maxParallel goroutines are active.
	sem := make(chan struct{}, *maxParallel)
	for _, d := range parallelDirs {
		d := d
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if !runPkg(d, false, *coverage, *retries) {
				failed.Store(true)
			}
		}()
	}

	wg.Wait()

	if *coverage {
		computeCoverage()
	}
	runTparse(*nocolor)

	if failed.Load() {
		os.Exit(1)
	}
}
