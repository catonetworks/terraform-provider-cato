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
result=ok
count=0

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
	timeout=8m
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
	# Skip reasons can mention known transient errors but do not mean the
	# package's actual failure is transient.
	grep -Ev '"Output":".*skipping test' "$1" |
		grep -Ei '(internal server error|connection (error|refused)|DOWNSTREAM_SERVICE_ERROR|message\\":\\"Internal server\\n")' \
			> /dev/null && return 0
	return 1
}
retry_test() {
	log="$1"; i=$2; count=$3; tdir=$4; cover_file=$5
	sleeptime=5

	for r in `seq $retry_count`; do
		if should_retry "$log"; then
			sleep $sleeptime; sleeptime=$((sleeptime+5))
			printf "  Retrying test %d/%d:\t%-30s" "$i" "$count" "$(basename "$tdir")"

			cleanup
			run_test "$tdir" "$cover_file" &> "$log" && { echo OK; return 0; }
			echo ERROR
		else
			result=error
			return
		fi
	done
	result=error
}

compute_coverage() {
	[ "$enable_coverage" = y ] || return
	(
		cd "$COVERAGE"
		echo 'mode: atomic' > cover.out
		cat [0-9]* | grep -v '^mode: atomic' >> cover.out
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

parse_args "$@"
get_test_dirs # -> $test_dirs

rm -rf "$OUT" "$COVERAGE"
mkdir -p "$OUT" "$COVERAGE"

[ -n "$test_dirs" ] || { echo "No tests selected"; exit; }

cleanup
i=0
for tdir in $test_dirs; do
	i=$((i+1))
	log_file="$(printf '%03d-%s' "$i" "$(basename "$tdir")")"

	printf "Running test %d/%d:\t%-30s" "$i" "$count" "$(basename "$tdir")"
	run_test "$tdir" "$COVERAGE/$log_file" &> "$OUT/$log_file" && { echo OK; continue; }
	echo ERROR

	retry_test "$OUT/$log_file" "$i" "$count" "$tdir" "$COVERAGE/$log_file"
done

compute_coverage
cat "$OUT/"* | go tool tparse -trimpath github.com/catonetworks/terraform-provider-cato/ --all $nocolor
if [ "$result" = ok ]; then exit 0; fi
exit 1
