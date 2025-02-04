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