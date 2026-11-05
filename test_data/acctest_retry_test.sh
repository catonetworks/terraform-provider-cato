#!/usr/bin/env bash
set -eu
# shellcheck source=test_data/acctest_retry.sh
source "$(dirname "$0")/acctest_retry.sh"
retry_fixture=$(mktemp)
trap 'rm -f "$retry_fixture"' EXIT
check_retry() {
 local expected=$1 output=$2 actual=1
 printf '%s\n' "$output" > "$retry_fixture"
 should_retry "$retry_fixture" && actual=0
 if [ "$actual" != "$expected" ]; then
  printf 'Unexpected retry result for %s\n' "$output" >&2
  exit 1
 fi
}
check_retry 0 '{"Output":"DOWNSTREAM_SERVICE_ERROR: Internal server error"}'
check_retry 0 '{"Output":"connection refused"}'
check_retry 0 '{"Output":"reorderPolicyBlockedByActiveSessions"}'
check_retry 1 '{"Output":"DOWNSTREAM_SERVICE_ERROR: Access denied: Insufficient privileges"}'
check_retry 1 '{"Output":"DOWNSTREAM_SERVICE_ERROR: classification=ValidationError"}'
check_retry 1 '{"Output":"skipping test: Internal server error"}'
check_retry 1 '{"Output":"ordinary assertion failure"}'
