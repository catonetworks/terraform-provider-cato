Terraform `CATO` Provider
=========================

> Versioning: This provider follows the [Terraform Provider Versioning Policy](terraform-provider-versioning.md).

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://raw.githubusercontent.com/hashicorp/terraform-website/master/public/img/logo-hashicorp.svg " width="300px">

Maintainers
-----------

This provider plugin is maintained by the team at [Cato](https://www.catonetworks.com/).

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.25.0 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-cato`

```sh
git clone git@github.com:catonetworks/terraform-provider-cato.git $GOPATH/src/github.com/terraform-providers/terraform-provider-cato
```

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/terraform-providers/terraform-provider-cato
make build
```

Using the Provider
---------------------------
Set the following 2 environment variables: 
```
export TF_VAR_token="abcde12345abcde12345"
export TF_VAR_account_id="12345"
```
Optional retry configuration:
```
export CATO_RETRY_MAX="5"
export CATO_RETRY_WAIT_MIN_SECONDS="1"
export CATO_RETRY_WAIT_MAX_SECONDS="30"
```

### Authentication

Choose one authentication method:

1. **Explicit credentials:** set `token` and `account_id` in the provider block, or use `CATO_TOKEN` and `CATO_ACCOUNT_ID`. Configure `baseurl` in the provider block or through `CATO_BASEURL`. Provider attributes override their environment values.
2. **catocli profile:** set `profile` in the provider block or `CATO_PROFILE` in the environment. The provider reads `cato_token`, `account_id`, and `endpoint` from that profile in `~/.cato/credentials`. Override the file path with `CATO_CREDENTIALS_FILE`. A selected profile overrides ambient credential environment variables; an explicit `baseurl` attribute overrides its endpoint.

A selected profile cannot be combined with `token` or `account_id` attributes in the provider block. This produces a diagnostic instead of silently switching accounts. For CI, use explicit credentials through environment variables and leave `CATO_PROFILE` unset.

```hcl
# explicit (unchanged)
provider "cato" {
  baseurl    = "..."
  token      = "..."
  account_id = "..."
}

# catocli profile
provider "cato" {
  profile = "refarch-dev"
}
```

```sh
# or via env, with no provider block attribute needed at all
export CATO_PROFILE=refarch-dev
terraform plan
```

Why use a profile: it's the same `~/.cato/credentials` store `catocli` itself reads and writes, so there's one shared credential store instead of two — no API tokens need to live in `*.tfvars` files or shell `export`s, and switching between tenants/accounts is just a matter of pointing `profile` (or `CATO_PROFILE`) at a different profile name.

To point at a credentials file somewhere other than `~/.cato/credentials` (useful in CI, or for tests that shouldn't depend on the invoking user's real credentials), set:
```
export CATO_CREDENTIALS_FILE="/path/to/alternate/credentials"
```

### Debugging GraphQL API Traffic

When Terraform debug logging is enabled, the provider logs each GraphQL API
request and response from the Cato SDK under the `API Call` debug message. The
log entry includes:

- `request`: the GraphQL query, operation name, and variables sent to Cato
- `response`: the raw GraphQL response body returned by Cato
- `trace_id`: the Cato server trace ID when returned by the API

Example:

```sh
export TF_LOG=DEBUG
export TF_LOG_PATH=~/Downloads/cato_terraform_debug.log
terraform apply
```

For easier support handoff, set `TF_API_DUMP_DIR` to write each GraphQL call to
its own file:

```sh
export TF_API_DUMP_DIR=~/Downloads/cato_api_calls
terraform apply
```

Debug logs and dump files can contain sensitive configuration values. Remove API
tokens and other secrets before sharing them outside your trusted support path.

Sample terraform files can be found in the examples folder in this repository.  You can initialize and run these terraform files with the following commands:
```
terraform init
terraform plan
terraform apply --auto-approve
```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
make install
...
$GOPATH/bin/terraform-provider-cato
...
```

In order to test the provider, you can simply run `make test`.

```sh
make test
```

To run all acceptance tests locally against one account, export `CATO_ACCOUNT_ID`,
`CATO_TOKEN`, `CATO_BASEURL`, and that account's `TFACC_TEST_VARS`, then run:

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
make acctest-flaky
# Run just one suite, or one of the CI groups, using the same local account:
make acctest-flaky t=if_rule
make acctest-flaky group=0
```
`make acctest` also runs the complete suite on one account, without retries.
Both commands run packages serially to avoid concurrent policy revisions.

GitHub Actions runs ten balanced acceptance groups concurrently. Each run shuffles
the ten account slots once, assigns a distinct account to each group, and holds a
workflow-wide lock until every group finishes. The existing required `test` check
passes only when all groups pass. Local runs do not need the CI account pool.

The `CATO_ACCTEST_ACCOUNTS` Actions secret is a JSON array of exactly ten records,
each with `id` (numeric string), `api_key`, and `test_vars` (that account's fixture
references). IDs and keys must be unique. Each job selects one record and uses its
fixture references; the old shared `TFACC_TEST_VARS` variable is not used in CI.
Keep this file and Terraform fixture state outside the checkout.

Timings were refreshed from a local `make acctest-flaky` run on 6 November 2026.
The timing inventory records the measured commit, platform, failed packages, and
skipped tests. Failed unsplit packages retain the previous CI estimate. Skipped-only
packages and skipped or failed split tests use explicit conservative estimates
rather than short skip or failure durations.

The planner uses longest-duration-first assignment with a three-second allowance
per package invocation. It splits `socket_site` and the independent
`bulk_policy_families` tests by top-level test so each can run across multiple
accounts. Group estimates exclude cleanup, compilation, setup, and retries;
account features and previously skipped coverage can change actual durations.
Current estimates are about 131–137 seconds per group (a 4.1% spread).

```sh
python3 test_data/acctest_groups.py summary
python3 test_data/acctest_groups.py validate
```

When adding a package, add its estimated seconds to
`test_data/acctest_timings.json`. For a package with individual test timings,
include every top-level test. Validation fails if packages or split tests are
missing, preventing silent omissions. Update durations from a new successful run
when the baseline becomes stale. See [fixture setup](test_data/acctest_fixtures/README.md)
for persistent account prerequisites.
Please refer to [Running TF Provider Tests](https://catonetworks.atlassian.net/wiki/spaces/RnD/pages/7257526403/Running+TF+Provider+Tests) for more details

## Helpful Terraform Aliases (Unix & Windows)

This guide explains how to create persistent Terraform helper aliases
on:

-   macOS / Linux (bash or zsh)
-   Windows PowerShell
-   Windows Git Bash

------------------------------------------------------------------------

## macOS / Linux (bash or zsh)

### 1. Open your shell config file

#### Bash

``` bash
nano ~/.bashrc
```

#### Zsh (default on macOS)

``` bash
nano ~/.zshrc
```

------------------------------------------------------------------------

### 2. Add the aliases

``` bash
alias tf='terraform'
alias tfap='terraform apply --auto-approve'
alias tfapp='terraform apply --auto-approve -parallelism=1'
alias tfdap='terraform destroy --auto-approve'
alias tfdapp='terraform destroy --auto-approve -parallelism=1'
alias tfclear='rm -rf .terraform* && rm terraform.tfstate*'
alias tftestapp='terraform test -parallelism=1'
alias tfdocs="terraform-docs markdown . --output-file README.md"
alias tfdebugon="export TF_LOG=DEBUG && export TF_LOG_PATH=~/Downloads/error_log.txt"
alias tfdebugoff="unset TF_LOG && unset TF_LOG_PATH"
alias tffmt="terraform fmt -recursive"
```

------------------------------------------------------------------------

### 3. Reload your shell

``` bash
source ~/.bashrc
```

or

``` bash
source ~/.zshrc
```

Your aliases are now persistent.

------------------------------------------------------------------------

## Windows -- PowerShell (Recommended)

PowerShell requires functions instead of bash-style aliases.

### 1. Open your PowerShell profile

``` powershell
notepad $PROFILE
```

If it does not exist:

``` powershell
New-Item -ItemType File -Path $PROFILE -Force
notepad $PROFILE
```

------------------------------------------------------------------------

### 2. Add the functions

``` powershell
function tf { terraform @args }

function tfap { terraform apply --auto-approve @args }

function tfapp { terraform apply --auto-approve -parallelism=1 @args }

function tfdap { terraform destroy --auto-approve @args }

function tfdapp { terraform destroy --auto-approve -parallelism=1 @args }

function tfclear {
    Remove-Item -Recurse -Force .terraform* -ErrorAction SilentlyContinue
    Remove-Item -Force terraform.tfstate* -ErrorAction SilentlyContinue
}

function tftestapp { terraform test -parallelism=1 @args }

function tfdocs {
    terraform-docs markdown . --output-file README.md
}

function tfdebugon {
    $env:TF_LOG="DEBUG"
    $env:TF_LOG_PATH="$HOME\tf_error_log.txt"
}

function tfdebugoff {
    Remove-Item Env:TF_LOG -ErrorAction SilentlyContinue
    Remove-Item Env:TF_LOG_PATH -ErrorAction SilentlyContinue
}

function tffmt {
    terraform fmt -recursive
}
```

------------------------------------------------------------------------

## 3. Restart PowerShell

Close and reopen your terminal to apply changes.

Your aliases are now persistent.

------------------------------------------------------------------------

# Windows -- Git Bash

If using Git Bash on Windows, follow the macOS/Linux instructions and
edit:

``` bash
nano ~/.bashrc
```

Then reload:

``` bash
source ~/.bashrc
```

------------------------------------------------------------------------

# Verification

Test with:

``` bash
tf --version
tfap
```

If Terraform runs, your aliases are working.
