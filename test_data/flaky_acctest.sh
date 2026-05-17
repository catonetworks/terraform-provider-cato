#!/usr/bin/env bash

export OUT=tmp_recorded/output
export TF_ACC=1
export TF_ACC_MOCK=''
retry_count=3
result=ok

cleanup() {
	ACCTEST_CLEANUP=true go test -tags acctest -count=1 --timeout=3m -run TestCleanupAccTestResources ./internal/acctests/acc > /dev/null
}

run_test() {
	go test -timeout 5m -tags acctest -count=1 -parallel=1 -p=1 "$1" -json
}

should_retry() {
	grep '\(internal server error\|connection refused\|message\\":\\"Internal server\\n"\)' "$1" > /dev/null && return 0
	return 1
}
retry_test() {
	log="$1"; i=$2; count=$3; tdir=$4
	sleeptime=5

	for r in `seq $retry_count`; do
		if should_retry "$log"; then
			sleep $sleeptime; sleeptime=$((sleeptime+5))
			printf "  Retrying test %d/%d:\t%-30s" "$i" "$count" "$(basename "$tdir")"

			cleanup
			run_test "$tdir" &> "$log" && { echo OK; return 0; }
			echo ERROR
		else
			result=error
			return
		fi
	done
	result=error
}

rm -rf "$OUT"
mkdir -p "$OUT"

test_dirs=$(find ./internal/acctests/ -type f -name '*_test.go' | sed 's|/[^/]*$||' | sort | uniq | grep -v '^./internal/acctests/acc$')
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
	run_test "$tdir" &> "$OUT/$log_file" && { echo OK; continue; }
	echo ERROR

	retry_test "$OUT/$log_file" "$i" "$count" "$tdir"
done

cat "$OUT/"* | go tool tparse -trimpath github.com/catonetworks/terraform-provider-cato/ --all --follow
if [ "$result" = ok ]; then exit 0; fi
exit 1
