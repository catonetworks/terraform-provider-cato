#!/usr/bin/env bash
PRG=$(basename "$0")
HELP="$PRG runs terraform acceptance tests
Usage:
  $PRG [ --coverage ] [ --suite <file-name> | <test-dir> ]
"
export OUT=tmp_recorded/output
export COVERAGE=tmp_recorded/coverage
export TF_ACC=1
export DISABLE_POLICY_RULE_CLEANUP=true
export TF_ACC_MOCK=''
enable_coverage=''
nocolor=''
test_suite=''
single_tests=''
retry_count=3

parse_args() {
	while [ $# -gt 0 ]; do
		case "$1" in
			-help|--help|-h) printf "%s\n\n" "$HELP"; exit 1;;
			--coverage) enable_coverage=y;;
			--nocolor) nocolor='-nocolor';;
			--suite) [ $# -gt 1 ] || { echo "Error: test suite file name expected"; exit 1; }
				shift; test_suite=$1;;
			*) single_tests="$single_tests $1"
		esac
		shift
	done
	if [ -n "$test_suite" ]; then
		[ -f "$test_suite" ] || { echo "Error: test suite file '$test_suite' does not exist"; exit 1; }
		single_tests=$(<$test_suite)
	fi
}

cleanup() {
	ACCTEST_CLEANUP=true go test -tags acctest -count=1 --timeout=3m -run TestCleanupAccTestResources ./internal/acctests/acc > /dev/null
}

run_test() {
	tdir=$1; cover_file=$2
	timeout=5m
	case "$(basename "$tdir")" in
	wf_rules_index | wf_rules_index_with_rule_data) timeout=12m ;;
	esac
	if [ "$enable_coverage" = y ]; then
		go test -timeout "$timeout" -tags acctest -count=1 -parallel=1 -p=1 "$tdir" -json -coverprofile="$cover_file" -covermode=atomic -coverpkg=./...
	else
		go test -timeout "$timeout" -tags acctest -count=1 -parallel=1 -p=1 "$1" -json
	fi
}

should_retry() {
	grep -E '(internal server error|connection refused|message\\":\\"Internal server\\n")' "$1" > /dev/null && return 0
	return 1
}

# is_shared_policy returns 0 if the package shares the IF/WF/WAN-Network policy
# cleanup and must run sequentially relative to other shared-policy packages.
is_shared_policy() {
	case "$(basename "$1")" in
		if_rule|if_rules_index|if_section|\
		wf_rule|wf_rules_index|wf_rules_index_with_rule_data|wf_section|\
		wnw_rule|wnw_rules_index|wan_network_section)
			return 0 ;;
	esac
	return 1
}

compute_coverage() {
	[ "$enable_coverage" = y ] || return
	(
		cd "$COVERAGE"
		echo 'mode: atomic' > cover.out
		cat -- * | grep -v '^mode: atomic' >> cover.out
		go tool cover -html=cover.out -o cover.html
		grep '<option value="file' cover.html | sed -e 's|.*catonetworks/terraform-provider-cato/||' -e 's/<.*//' -e 's/\(.*\) (\([0-9][0-9.%]*\))/\2\t\1/' | sort -n  > cover-stats.txt
		printf "\nTest Coverage\n~~~~~~~~~~~~~\n"
		cat cover-stats.txt
	)
}

get_test_dirs() {
	test_dirs=$(find ./internal/acctests/ -type f -name '*_test.go' | sed 's|/[^/]*$||' | sort | uniq | grep -v '^./internal/acctests/acc$')
	count="$(echo "$test_dirs" | wc -l)"
	[ -n "$single_tests" ] || return

	new_tests=''; count=0
	for td in $test_dirs; do
		base_dir=$(basename "$td")
		for st in $single_tests; do
			if [ "$base_dir" = "$(basename "$st")" ]; then
				new_tests="$new_tests $td"; count=$((count + 1)); break
			fi
		done
	done
	test_dirs=$new_tests
}

# run_stream runs a list of package directories sequentially and exits with 0/1.
# needs_cleanup=y means call cleanup() before each retry.
run_stream() {
	local needs_cleanup="$1"; shift
	local stream_result=0
	for tdir in "$@"; do
		local pkg log_file cover_file
		pkg=$(basename "$tdir")
		log_file="$OUT/$pkg"
		cover_file="$COVERAGE/$pkg"

		printf "Starting:\t%s\n" "$pkg"
		if run_test "$tdir" "$cover_file" &> "$log_file"; then
			printf "OK:\t\t%s\n" "$pkg"
			continue
		fi
		printf "ERROR:\t\t%s\n" "$pkg"

		local retried=n
		for r in $(seq $retry_count); do
			if ! should_retry "$log_file"; then
				stream_result=1; break
			fi
			sleep 5
			[ "$needs_cleanup" = y ] && cleanup
			printf "Retrying (%d/%d):\t%s\n" "$r" "$retry_count" "$pkg"
			retried=y
			if run_test "$tdir" "$cover_file" &> "$log_file"; then
				printf "OK (retry %d):\t%s\n" "$r" "$pkg"
				break
			fi
			printf "ERROR (retry %d):\t%s\n" "$r" "$pkg"
			[ "$r" = "$retry_count" ] && stream_result=1
		done
		[ "$retried" = n ] && stream_result=1
	done
	return "$stream_result"
}

parse_args "$@"
get_test_dirs # -> $test_dirs

rm -rf "$OUT" "$COVERAGE"
mkdir -p "$OUT" "$COVERAGE"

[ -n "$test_dirs" ] || { echo "No tests selected"; exit; }

cleanup

# Split packages into two groups:
#   serial_dirs  — share the IF/WF/WAN-Network policy cleanup; must run sequentially.
#   parallel_dirs — independent; each runs as its own background job.
serial_dirs=""
parallel_dirs=""
for tdir in $test_dirs; do
	if is_shared_policy "$tdir"; then
		serial_dirs="$serial_dirs $tdir"
	else
		parallel_dirs="$parallel_dirs $tdir"
	fi
done

serial_count=$(echo $serial_dirs | wc -w | tr -d ' ')
parallel_count=$(echo $parallel_dirs | wc -w | tr -d ' ')
echo "Parallel streams: $parallel_count independent + 1 shared-policy ($serial_count packages)"

# Launch the shared-policy stream as one sequential background subshell.
# shellcheck disable=SC2086
(run_stream y $serial_dirs) &
bg_pids="$!"

# Launch each independent package as its own background job (fully parallel).
# Both the serial stream and all independent jobs run concurrently.
for tdir in $parallel_dirs; do
	# shellcheck disable=SC2086
	(run_stream n $tdir) &
	bg_pids="$bg_pids $!"
done

# Wait for all background jobs; collect failures.
overall_result=0
for pid in $bg_pids; do
	wait "$pid" || overall_result=1
done

compute_coverage
cat "$OUT/"* | go tool tparse -trimpath github.com/catonetworks/terraform-provider-cato/ --all $nocolor
if [ "$overall_result" -eq 0 ]; then exit 0; fi
exit 1
