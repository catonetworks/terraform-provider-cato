#!/usr/bin/env bash

export OUT=tmp_recorded/output
export COVERAGE=tmp_recorded/coverage
export TF_ACC=1
export TF_ACC_MOCK=''
enable_coverage=''
retry_count=3
result=ok

cleanup() {
	ACCTEST_CLEANUP=true go test -tags acctest -count=1 --timeout=3m -run TestCleanupAccTestResources ./internal/acctests/acc > /dev/null
}

run_test() {
	tdir=$1; cover_file=$2
	if [ "$enable_coverage" = y ]; then
		go test -timeout 5m -tags acctest -count=1 -parallel=1 -p=1 "$tdir" -json -coverprofile="$cover_file" -covermode=atomic -coverpkg=./...
	else
		go test -timeout 5m -tags acctest -count=1 -parallel=1 -p=1 "$1" -json
	fi
}

should_retry() {
	grep '\(internal server error\|connection refused\|message\\":\\"Internal server\\n"\)' "$1" > /dev/null && return 0
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

rm -rf "$OUT" "$COVERAGE"
mkdir -p "$OUT" "$COVERAGE"

test_dirs=$(find ./internal/acctests/ -type f -name '*_test.go' | sed 's|/[^/]*$||' | sort | uniq | grep -v '^./internal/acctests/acc$')
if [ "$1" = "--coverage" ]; then
	enable_coverage=y
	shift
fi
if [ -n "$1" ]; then
	one_test=$(basename "$1")
	test_dirs=$(echo "$test_dirs" | grep "./internal/acctests/$one_test\$")
fi
[ -n "$test_dirs" ] || { echo "No tests selected"; exit; }
count="$(echo "$test_dirs" | wc -l)"

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
cat "$OUT/"* | go tool tparse -trimpath github.com/catonetworks/terraform-provider-cato/ --all
if [ "$result" = ok ]; then exit 0; fi
exit 1
