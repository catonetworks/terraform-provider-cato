---
name: rebalance-acctest-groups
description: Rebalance this provider's acceptance-test groups from CI timings or authorized local measurements, including new suites and per-test splits. Ask for local account configuration and API-key sources before live execution.
---

# Rebalance acceptance groups

Refresh `test_data/acctest_timings.json` so the existing planner distributes the
acceptance workload evenly. Group membership is computed, not maintained as ten
handwritten suite lists. Preserve the current group count unless the user asks
to expand the account pool.

## Choose the timing source

Read `test_data/acctest_groups.py`, the timing inventory, the acceptance runner,
and the acceptance section of `README.md` before changing the plan. Inspect the
working tree and preserve unrelated changes.

Use a user-supplied successful CI run when available. Otherwise ask whether to
use an existing CI run or measure locally. Independent inventory inspection and
offline planning can proceed while waiting. For Cato GitHub operations, apply
the repository's GitHub identity skill before using `gh`.

## Ask for configuration before local execution

Before running real acceptance tests locally, ask for these together unless
the user has already provided them for this execution:

- The dedicated test account ID and HTTPS GraphQL endpoint (`CATO_ACCOUNT_ID`,
  `CATO_BASEURL`; use the same endpoint for `CATO_ENDPOINT` when needed).
- The API key for that account (`CATO_TOKEN`), supplied through an existing
  environment variable, a protected local file, or an authorized secret-manager
  item. Ask for the source or reference, not for a key pasted into chat.
- The matching account's fixture references (`TFACC_TEST_VARS`), optional
  `TFACC_TEST_SKIP`, and feature switches such as
  `TFACC_ENABLE_RULES_INDEX_CRUD`.
- Which suites or the full suite to measure, and authorization to run their
  real resource creation and cleanup on that account.

Suggested prompt: “For local measurements, which dedicated account and HTTPS
endpoint should I use, where should I load its API key and fixture configuration,
and may I run acceptance tests and cleanup there?”

If the user points to 1Password, resolve the configuration and credential items
they authorize; do not assume a vault, account, or credentials from another task.
Never put keys in skill files, timing data, commands, logs, commits, or replies.
Load credentials through environment variables or the approved secret manager.
Keep secret-bearing files and raw logs outside Git, with owner-only access.
Check required values without printing their contents. Do not dump the process
environment or enable provider debug logging.

If inputs or execution authorization are missing, continue offline work and
leave live measurement pending. Creating or invoking this skill alone does not
authorize live tests. CI-log-only rebalancing needs no local API credentials.

## Collect comparable measurements

For local measurement, use the repository's acceptance commands with the supplied
environment, `DISABLE_POLICY_RULE_CLEANUP=true`, and serial execution on one
dedicated account:

```sh
make acctest-flaky t=<suite>
# Full measurement, only when requested:
make acctest-flaky
```

The runner performs cleanup and may delete disposable acceptance resources.
Check [fixture prerequisites](../../../test_data/acctest_fixtures/README.md)
before running. Missing fixtures or rollout prerequisites are failures to fix
or report, not a reason to silently skip coverage. Stop on persistent failures
after the runner's bounded retries; do not loop indefinitely or alter account
features, grants, credentials, or CI secrets just to obtain timings.

Use successful package execution durations from Go JSON events or the runner's
package summary. Exclude runner setup, compilation, cleanup, backoff, failed
attempts, and workflow wall time. Prefer one comparable successful attempt per
suite; do not add retry attempts together. Record skips separately from measured
passes so skipped coverage does not appear artificially fast.

For CI matrix logs, combine all jobs from the same run. Packages split across
groups need their selected top-level test durations combined; never overwrite
the full package estimate with just one group's partial duration. Keep private
logs outside the repository and extract only names, durations, and provenance.

## Update and assess the plan

- Update package `seconds` in `test_data/acctest_timings.json`. Add new package
  names exactly as their directory names under `internal/acctests/`.
- Preserve provenance. Update the source run URL/commit for a complete baseline;
  identify local or partial measurements honestly rather than attributing them
  to the old source job. Mark estimates when no successful measurement exists.
- Preserve existing individual `tests` mappings. For a dominant package, consider
  splitting independent top-level tests: include every declared top-level test,
  including skipped ones, with anchored selection handled by the existing planner.
  Do not split tests that depend on shared package state or execution order.
- Keep the planner's invocation allowance separate from measured test durations.
  Change the balancing algorithm or overhead only with evidence that timing
  updates alone cannot produce useful balance.

```sh
python3 test_data/acctest_groups.py validate
python3 test_data/acctest_groups.py summary
python3 -m unittest discover -s test_data -p '*_test.py'
git diff --check
```

Compare before/after minimum and maximum group estimates and verify every suite
and split test appears exactly once. Local `make acctest-flaky` must still run
the full suite on one account; group selection must remain optional. Preserve
the random one-to-one group/account assignment and workflow-wide account-pool
lock. Any new fixture must exist in every pool account before randomized CI runs.

Expanding the number of groups is a separate scope: update `GROUP_COUNT`, pool
validation, tests, workflow parallelism, help text, and documentation together,
and arrange the additional dedicated accounts and fixtures. Do not assume an
extra account or create credentials without authorization.

Report changed timing sources, the before/after estimated spread, checks run,
and whether live tests ran or remain pending. Distinguish estimated balance
from observed CI wall time. Commit, push, or trigger CI only within the user's
requested scope.
