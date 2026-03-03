Terraform `CATO` Provider
=========================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/0/04/Terraform_Logo.svg/768px-Terraform_Logo.svg.png" width="600px">

Maintainers
-----------

This provider plugin is maintained by the team at [Cato](https://www.catonetworks.com/).

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.19 (to build the provider plugin)

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

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
make testacc
```

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
