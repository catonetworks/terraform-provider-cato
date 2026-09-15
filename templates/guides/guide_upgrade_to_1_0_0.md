---
subcategory: ""
page_title: "Upgrade to Cato Terraform Provider 1.0.0"
description: |-
  Upgrade from Cato Terraform provider 0.x to 1.0.0
---

# Upgrade to Cato Terraform Provider 1.0.0

Version 1.0.0 establishes the Cato Terraform provider's stable compatibility contract. It does not introduce breaking changes: valid 0.x configuration, Terraform state, and import identifiers continue to work without migration.

## Before you upgrade

1. Commit your current `.terraform.lock.hcl` and back up remote Terraform state according to your normal operational process.
2. Update the root module's provider constraint to allow the 1.x line:

```hcl
terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = ">= 1.0.0, < 2.0.0"
    }
  }
}
```

3. Upgrade in a controlled branch or non-production environment:

```sh
terraform init -upgrade
terraform plan
```

Review the provider lock-file update and plan before applying. No state upgrade, re-import, resource replacement, or configuration changes are expected solely from upgrading to 1.0.0.

## Version availability notification

The provider now performs a best-effort, non-blocking check for a newer stable provider version at startup. No configuration is required. In environments where outbound access to the Terraform Registry is not allowed, disable the notification:

```hcl
provider "cato" {
  version_check_disabled = true
}
```

Alternatively, set any non-empty `CATO_VERSION_CHECK_DISABLED` environment variable. The check has a two-second default timeout and never blocks `terraform plan` or `terraform apply`.

## Roll back

No state migration was introduced in 1.0.0. To return to 0.x, restore a provider constraint that selects the desired 0.x version, run `terraform init -upgrade`, and review `terraform plan` before applying. Retain your normal state backup throughout the upgrade and rollback process.

## 0.x maintenance status

The 0.x line remains supported alongside 1.x. Its end-of-support date has not yet been scheduled. A future major release will publish its support timeline and migration guidance before any required customer migration.
