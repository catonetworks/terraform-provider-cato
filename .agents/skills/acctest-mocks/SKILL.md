---
name: acctest-mocks
description: Create or update mock data for Terraform provider acceptance tests using recorded GraphQL calls under tmp_recorded and replay fixtures under test_data. Use when Codex needs to generate or refresh mocked API data for acctests instead of real Cato API calls, inspect GraphQL operationName traffic, build accmock config.yaml files, or create resource mock YAML files for TF_ACC_MOCK=1 runs.
---

# Terraform Acctest Mocks

Use this skill to create or update mock data for tests under `./internal/acctests/`. Mock fixtures live under `./test_data/<TestFunctionName>/` and are consumed when running tests with `TF_ACC_MOCK=1`.

## Prerequisites

1. Ensure `CATO_ACCOUNT_ID` is set.
2. Ask the user to confirm the account ID is correct. Stop if the user does not confirm.

## Identify The Test

1. Determine the exact acceptance test function. Tests are in `./internal/acctests/<test-directory>/` and functions start with `TestAcc`.
2. If the prompt does not specify the exact test, ask the user before continuing.
3. Record:
   - Test function name, such as `TestAccStaticHost`.
   - Test directory, such as `static_host`.
4. Use the test function name as the mock fixture name under `./test_data/`.

## Record Real API Traffic

1. Delete `./tmp_recorded/` before recording so only this run's traffic is present. This is an intentional cleanup step for generated recording output.
2. Run the focused real-API test and require exit code 0:

```bash
TF_ACC=1 TF_ACC_MOCK='' go test -timeout 120s -tags acctest -count=1 -parallel=1 -p 1 -run <TEST-NAME> ./internal/acctests/<TEST-DIRECTORY>/ -v
```

3. Inspect `./tmp_recorded/<TEST-NAME>/`. Recorded filenames use `<time>_<operationName>.txt`, for example `20260519_110848.065_entityLookup.txt`.
4. Extract every distinct `operationName` from those filenames.
5. For each recorded file, inspect the line that starts with `{"operationName":`; this line is the GraphQL request JSON body.
6. Use the last line of each recorded file as the real API response to adapt into the mock YAML body.

## Classify Operations

For each GraphQL request, determine:

1. Operation type: `CREATE`, `READ`, `UPDATE`, `DELETE`, or `NO_CONTENT`.
2. Target resource, such as `networkRangeList` or `staticHost`.
3. Resource name or ID variable path:
   - For `CREATE`, find the path to the name, such as `variables.addStaticHostInput.name`.
   - For `READ`, `UPDATE`, and `DELETE`, find the path to the ID, such as `variables.hostId`.
4. Whether the operation is static. Static operations query existing shared resources and do not mutate the resource under test.

To identify static resources, read `./internal/acctests/acc/common.go` and scan for `GetXxxx(t *testing.T)` helper functions. Resources returned by those helpers should be treated as static fixtures.

Special cases:

1. `entityLookup`: classify as `READ`; define subtypes from `variables.type`.
2. `policyPrivateAccessDiscardRevision`: classify as `NO_CONTENT`.

## Create Fixture Layout

1. Create `./test_data/<TEST-NAME>/`.
2. Create one resource-type directory per mocked resource, such as `./test_data/TestAccStaticHost/staticHost/`.
3. Create one resource-name directory per resource instance, such as `./test_data/TestAccStaticHost/staticHost/acctest_static_host_1/`.
4. Keep names deterministic. If the real response contains randomized suffixes, remove only the random part, for example convert `acctest_static_host_fwcowd6p4h` to `acctest_static_host`.

## Create config.yaml

1. Inspect `type config struct` in `./internal/accmock/mockserver.go`.
2. Inspect an existing fixture such as `./test_data/TestAccAppConnector/config.yaml`.
3. Create or update `./test_data/<TEST-NAME>/config.yaml` to map operations to the resource fixtures needed by this test.
4. Match existing config style and field names exactly; do not invent schema fields.

## Create Mock YAML Files

1. In each resource-name directory, create action files named `<sequence>_<action>.yaml`.
2. Use a three-digit sequence number such as `000`.
3. Use action names:
   - `create`
   - `read`
   - `update`
   - `zap` for delete cleanup
4. For create files, define `ResourceID`. Use a deterministic mocked database ID such as `"1000"` unless an existing fixture pattern requires otherwise.
5. Every action file must include `GraphQL.StatusCode`, `GraphQL.Delay`, and `GraphQL.Body`.
6. Use the recorded API response as the basis for `GraphQL.Body`, preserving response shape while normalizing randomized names.
7. Ensure IDs returned inside `GraphQL.Body` match the `ResourceID` used by the create fixture.

Example shape:

```yaml
ResourceID: "1000"
GraphQL:
  StatusCode: 200
  Delay: 0ms
  Body: { "data": { "mock": { "id": "1000", "name": "acctest_static_host_1" } } }
```

## Test Mock Replay

Run the same focused test with mock replay enabled:

```bash
TF_ACC=1 TF_ACC_MOCK='1' go test -timeout 120s -tags acctest -count=1 -parallel=1 -p 1 -run <TEST-NAME> ./internal/acctests/<TEST-DIRECTORY>/ -v
```

The test must pass. If it fails, compare the failure against `./internal/accmock/mockserver.go`, the generated `config.yaml`, and the recorded request/response files, then adjust the fixture mappings or bodies.

## Restrictions

- Do not commit to git.
- Do not remove unrelated files under `./test_data/` or local scratch output.
- Do not edit provider source or acceptance test source unless the user explicitly asks for that separate change.
