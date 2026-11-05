# Acceptance test prerequisites

Live suites require `TF_ACC=1`, `DISABLE_POLICY_RULE_CLEANUP=true`,
`CATO_BASEURL`, `CATO_TOKEN`, and `CATO_ACCOUNT_ID`. Run the full suite with
`make acctest-flaky`, or select a package with `make acctest-flaky t=mailing_list`.

Site Web Proxy, Socket Bypass, and Terminal Server suites check published-policy
access before creating fixtures. An explicit API permission denial for the selected
account and policy is reported as a skip. Those suites still need an authorized
account for live coverage. Network errors, incomplete responses, and other API
errors fail the prerequisite check.

Mailing-list membership clearing needs an existing `LOGIN` administrator in the
selected account. Fixture lookup pages through administrators and excludes service
principals, which cannot receive notification emails. A missing fixture fails with
an explicit prerequisite message.

Policy cleanup treats `PolicyRevisionNotFound` as an already-completed discard.
Other cleanup failures retain their API error codes and messages. Cleanup sends
matching revision IDs in both discard and policy-mutation inputs.

The retry runner does not retry permission or validation failures. It still
retries transient downstream/connection failures and concurrent policy-reorder
blocks. `make test` includes regression checks for this classification.

CI also runs acceptance helper tests against local HTTP fixtures, without making
Cato API calls, before starting live acceptance tests. To run those helper tests:

```sh
go test -tags=acctest ./internal/acctests/acc ./internal/acctests/mailing_list \
  -run '^Test(WithStableResourceID|CheckPolicyPlacement|RequireLiveSkipsBeforeFixtures|PolicyAccessDenied|RequirePolicyAccess|CleanupMutationErrors|DiscardWanFirewallPolicyRevision|EligibleMailingListAdmin|FindMailingListAdmin)$'
```
