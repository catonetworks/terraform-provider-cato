#!/usr/bin/env bash
# Classify the final package output, not transient codes echoed in skip reasons.
should_retry() {
 local retry_log
 retry_log=$(grep -Ev '"Output":".*skipping test' "$1")
 # These are persistent prerequisite/input failures even if wrapped as downstream errors.
 if printf '%s\n' "$retry_log" | grep -Eqi '(Insufficient privileges|Access denied|ValidationError)'; then
  return 1
 fi
 printf '%s\n' "$retry_log" |
  grep -Eqi '(internal server error|connection (error|refused)|DOWNSTREAM_SERVICE_ERROR|reorderPolicyBlockedByActiveSessions|message\\":\\"Internal server\\n")'
}
