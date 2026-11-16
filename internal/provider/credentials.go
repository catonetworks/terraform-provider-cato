package provider

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// credentialsFileEnvVar overrides the catocli credentials file path. Primarily useful
// for CI and tests that shouldn't depend on the invoking user's real ~/.cato/credentials.
const credentialsFileEnvVar = "CATO_CREDENTIALS_FILE"

// defaultCredentialsFilePath returns the default catocli credentials file path
// (~/.cato/credentials), the same AWS-CLI-style profile store catocli itself reads
// and writes via `catocli configure set`.
func defaultCredentialsFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cato", "credentials")
}

// credentialsFilePath resolves the path to the catocli credentials file, honoring the
// CATO_CREDENTIALS_FILE environment variable override.
func credentialsFilePath() string {
	if v := os.Getenv(credentialsFileEnvVar); v != "" {
		return v
	}
	return defaultCredentialsFilePath()
}

// parseINISections parses an AWS-CLI-style INI file: "[section]" headers followed by
// "key = value" (or "key: value") lines, until the next section header or EOF. Values
// are read raw -- no `%`-interpolation is performed, so values containing literal `%`
// characters (e.g. URL-encoded tokens) are preserved verbatim. Blank lines and lines
// beginning with "#" or ";" are treated as comments and skipped. Keys outside of any
// section are ignored.
func parseINISections(r io.Reader) (map[string]map[string]string, error) {
	sections := map[string]map[string]string{}
	var current string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.TrimSpace(line[1 : len(line)-1])
			if _, ok := sections[current]; !ok {
				sections[current] = map[string]string{}
			}
			continue
		}

		if current == "" {
			continue // stray key/value pair before any section header
		}

		key, value, ok := splitINIKeyValue(line)
		if !ok {
			continue
		}
		sections[current][key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return sections, nil
}

// splitINIKeyValue splits a "key = value" or "key: value" line. Only the first "="
// or ":" is treated as the delimiter, so values are free to contain either character.
func splitINIKeyValue(line string) (key, value string, ok bool) {
	idx := strings.IndexAny(line, "=:")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

// ResolveProfileCredentials loads the named profile's section from the catocli-style
// credentials file at path, returning its token, account ID, and endpoint. Only the
// cato_token, account_id, and endpoint keys are consumed -- scim_url, scim_token,
// cookie, private_endpoint, and any other keys present in the section are ignored.
//
// Errors never include credential values. A missing section's error lists the
// available profile names (never their contents) so the caller can diagnose a typo.
func ResolveProfileCredentials(profileName, path string) (token, accountID, endpoint string, err error) {
	if profileName == "" {
		return "", "", "", fmt.Errorf("no profile name provided")
	}
	if path == "" {
		return "", "", "", fmt.Errorf("could not determine the catocli credentials file path (set %s to override)", credentialsFileEnvVar)
	}

	f, err := os.Open(path) // #nosec G304 -- path is the user's own catocli credentials file
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("catocli credentials file not found at %q", path)
		}
		return "", "", "", fmt.Errorf("unable to read catocli credentials file at %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	sections, err := parseINISections(f)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to parse catocli credentials file at %q: %w", path, err)
	}

	section, ok := sections[profileName]
	if !ok {
		available := make([]string, 0, len(sections))
		for name := range sections {
			available = append(available, name)
		}
		sort.Strings(available)
		return "", "", "", fmt.Errorf(
			"profile %q not found in catocli credentials file %q (available profiles: %s)",
			profileName, path, strings.Join(available, ", "),
		)
	}

	token = section["cato_token"]
	accountID = section["account_id"]
	endpoint = section["endpoint"]

	if token == "" {
		return "", "", "", fmt.Errorf("profile %q in %q is missing required key %q", profileName, path, "cato_token")
	}
	if accountID == "" {
		return "", "", "", fmt.Errorf("profile %q in %q is missing required key %q", profileName, path, "account_id")
	}

	return token, accountID, endpoint, nil
}
