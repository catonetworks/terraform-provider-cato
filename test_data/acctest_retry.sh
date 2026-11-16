#!/usr/bin/env bash
# Classify the final package output, not transient codes echoed in skip reasons.
should_retry() {
 local retry_log
 retry_log=$(grep -Ev '"Output":".*skipping test' "$1") || return 1
 # These are persistent prerequisite/input failures even if wrapped as downstream errors.
 if grep -Eqi '(Insufficient privileges|Access denied|ValidationError)' <<< "$retry_log"; then
  return 1
 fi
 grep -Eqi '(internal server error|connection (error|refused)|DOWNSTREAM_SERVICE_ERROR|reorderPolicyBlockedByActiveSessions|message\\":\\"Internal server\\n")' <<< "$retry_log"
}
