# Local DNS/DHCP provider testing against dev-us1-m1

Terraform runs the provider binary locally and calls the existing dev GraphQL API.
No local backend services are required. This fixture manages six older standalone
resources, not DNS/DHCP profiles and not `accountUpdate`.

## Prerequisites

- Go version from repository `go.mod`, Terraform CLI >= 1.9, Python 3.
- Account-scoped API key with DNS/DHCP read and write permissions, configured as
  `CATO_TOKEN`. CMA browser cookie/session token is not an API key.
- Effective `NCS_DNS_PUBLIC_API` and `NCS_DHCP_PUBLIC_API` tags on test account.
- Legacy DNS/DHCP settings editable; profile mode may reject legacy writes.
- Existing test site in selected account, and test IPs/unused DHCP option code.

Export target explicitly. Keep API key in environment; do not enable shell tracing:

```sh
export CATO_BASEURL=https://system.cc.dev-us1-m1.catonet.works/api/v1/graphql
# Set CATO_TOKEN from your private account-scoped API key source.
# Set CATO_ACCOUNT_ID and CATO_TEST_SITE_ID to your test account and existing site.
export TF_VAR_account_id="$CATO_ACCOUNT_ID"
export TF_VAR_site_id="$CATO_TEST_SITE_ID"
export TF_VAR_baseurl="$CATO_BASEURL"
export DISABLE_POLICY_RULE_CLEANUP=true
export CATO_VERSION_CHECK_DISABLED=true
export CATO_RETRY_MAX=0
```

`DISABLE_POLICY_RULE_CLEANUP=true` is required even for plan/read: provider
configuration otherwise discards private-access draft policy revisions.

## Select the local provider

From repository root:

```sh
mkdir -p bin
go build -o bin/terraform-provider-cato .
```

Create an ignored `dev.tfrc` under this fixture. Use absolute path to your checkout:

```hcl
provider_installation {
  dev_overrides {
    "catonetworks/cato" = "/absolute/path/to/terraform-provider-cato/bin"
  }
  direct {}
}
```

From fixture directory, export `TF_CLI_CONFIG_FILE` as absolute path to `dev.tfrc`.
The override selects local binary for schema, plan, apply, and import. Rebuild after
provider edits. `terraform init` still selects a published version for lock file;
it is not proof that local binary is used. This fixture has no modules/remote backend,
so use dev override directly for operational commands.

```sh
terraform providers schema -json > schema.json
```

Inspect six new `cato_dns_*`/`cato_dhcp_*` resource schemas under
`registry.terraform.io/catonetworks/cato`. See [HashiCorp development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

## Capture original configuration

```sh
python3 preflight.py capture
```

This is read-only. It verifies site ownership and DNS/DHCP API access, then records
account settings, site override/inheritance, and complete item collections in
ignored `baseline.json` with mode 0600. It refuses to overwrite an existing baseline.
A rollout/authentication error must be resolved before applying; helper never changes tags.
Terraform validation also restricts endpoint to dev-us1-m1 and requires effective account/site
variables to equal captured baseline target, including overrides from tfvars. This fixture
uses Terraform >= 1.9 for cross-object variable validation; provider itself needs no new version floor.

Create ignored `test.tfvars` supplying required fixture variables. Copy account DNS,
DHCP lease/relay values from baseline for first apply. Use original relay group ID
(or `null` if absent), original timeout (or `null`), and original enabled flag.
Provide unique `test_name`, `forwarding_domain`, valid `test_servers`, unused
`dhcp_option_tag`, and `dhcp_option_value`. Site override uses account DNS values in
fixture; adapt site inputs if you want different values. Addresses in published
examples are documentation addresses, not valid servers for your environment.

Do not select test relay group if original account group is absent: API cannot clear
selection back to null. Keep new relay group standalone unless original group can
be restored explicitly. Do not manage same singleton/item from another Terraform state.

## Apply, verify, update and import

```sh
terraform plan -parallelism=1 -var-file=test.tfvars -out=dev.tfplan
terraform apply -parallelism=1 dev.tfplan
terraform plan -parallelism=1 -var-file=test.tfvars -detailed-exitcode
```

Require exactly six creates/upserts in selected account and site. Final plan must exit
0 (2 means changes, 1 means error). Parallelism 1 serializes legacy account writes.
Confirm item IDs and settings independently through `networkConfig.dns`/`dhcp` and CMA.
Verify unrelated items remain intact. This proves API persistence, not packet behavior.

Update test option value/description, rule domain/servers, relay group name/servers,
DNS secondary/suffix, and lease time; re-apply and require stable item IDs/no-op second
plan. Removing secondary emits explicit null; suffix `[]` clears it; removing option
description clears it. Change one test item through API/CMA and verify Terraform detects
and repairs drift.

Import into a separate local state directory using same override and target. Account
singletons import provider account ID; site override imports site ID; other resources
import opaque IDs from outputs. Import/read only; remove temporary imported ownership
before cleanup. Do not destroy from a second state managing same configuration.

## Cleanup and restore

Before destroy, restore original account DHCP values in fixture and apply them, so
new relay group is not selected. Account singleton destroy retains current remote
settings, and deleting an in-use relay group fails instead of changing account config.

```sh
terraform plan -destroy -parallelism=1 -var-file=test.tfvars -out=destroy.tfplan
terraform apply destroy.tfplan
python3 preflight.py restore-account
python3 preflight.py restore-site
python3 preflight.py verify
```

Destroy deletes only three test collection items and site override. Restore commands
restore captured account settings and original site override/inheritance; verify
compares settings and all collections against baseline. Keep state and baseline until
verification passes. `restore-account` is also available before cleanup after a failed
run; Terraform config/state must be reconciled before applying again.

Never use existing global `make acctest` cleanup or regression `run-test` for this fixture.

## Automated tests

Mocked protocol test runs real Terraform against local HTTP fixtures and current provider
source, with no account writes:

```sh
TF_ACC_TERRAFORM_PATH=/absolute/path/to/terraform \
  go test ./internal/provider/network/dnsdhcptest -count=1
python3 -m unittest test_preflight
```

Opt-in real test captures/restores account/site settings, creates three unique collection
items, verifies no-op plans and imports, and updates option value. The automated fixture orders writes with `depends_on`;
concurrent legacy DNS writes caused a newly created forwarding rule to disappear in
live testing. Keep `-parallelism=1` for manual apply/destroy. Set
`TFACC_DNS_DHCP_SERVERS` to comma-separated valid test server IPs and
`TFACC_DHCP_OPTION_TAG` to an unused option code. Do not run in parallel with manual fixture:

```sh
unset TF_ACC_MOCK
TF_ACC=1 TFACC_ENABLE_DNS_DHCP_CRUD=true DISABLE_POLICY_RULE_CLEANUP=true \
  go test -tags=acctest -count=1 -parallel=1 -p=1 -timeout=20m \
  ./internal/acctests/dns_dhcp -run '^TestAccDNSDHCP$'
```

The suite skips unless `TFACC_ENABLE_DNS_DHCP_CRUD=true`; explicitly enabled runs
still require all endpoint and input checks.

Live tests change account/site configuration. They are separate from ordinary unit checks
and require an explicitly selected dev test account. Inspect retained baseline and errors
if a run fails; network failure can also prevent cleanup.
