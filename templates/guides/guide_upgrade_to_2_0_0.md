---
subcategory: ""
page_title: "Upgrade to Cato Terraform Provider 2.0.0"
description: |-
  Migrate Cato Terraform provider 1.x configurations to 2.0.0
---

# Upgrade to Cato Terraform Provider 2.0.0

Version 2.0.0 contains breaking changes to authentication selection, the mDNS reflector default, Socket LAN firewall sub-rule placement, and TLS inspection category validation. It also removes 14 policy resource types used in prereleases. Review each applicable change before running Terraform against an existing account. New resources remain subject to the API features enabled for that account.

## Before you upgrade

1. Commit the current configuration and `.terraform.lock.hcl`. Back up state using your normal secure state-backup process; state can contain credentials and other sensitive values.
2. Record the intended account ID for every configured Cato provider, including aliases, and check the environment used by local shells and CI for `CATO_PROFILE`.
3. Apply the configuration changes below in a controlled branch and test in a non-production environment. If a prerelease configuration uses an abbreviated Private Access or Remote Port Forwarding type, complete its import migration before changing the version constraint.
4. Update the root module's version constraint:

```hcl
terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}
```

5. Run `terraform init -upgrade`, then `terraform plan`. Check the provider lock-file update, selected account, policy matches, membership ownership, and proposed infrastructure changes before applying. A successful plan is not a substitute for confirming the target account.

## Authentication selection

Version 1.x ignored `CATO_PROFILE`. Version 2.0.0 reads the selected catocli profile from `~/.cato/credentials` (or the file selected by `CATO_CREDENTIALS_FILE`). An existing export can therefore affect an unchanged Terraform configuration:

- A selected profile together with explicit `token` or `account_id` provider attributes is an error.
- With credentials supplied only through environment variables, the selected profile overrides `CATO_TOKEN`, `CATO_ACCOUNT_ID`, and `CATO_BASEURL`, potentially selecting a different account.
- An explicit `baseurl` provider attribute overrides the profile endpoint.

To retain explicit or environment-based authentication, remove any `profile` provider attribute and unset the profile selection in the actual execution environment, including CI jobs:

```sh
unset CATO_PROFILE
terraform plan
```

To use profile authentication, configure only the profile selection rather than explicit `token` or `account_id` attributes:

```hcl
provider "cato" {
  profile = "my-profile"
}
```

Confirm the chosen profile's account and endpoint before use. The provider emits a diagnostic containing the resolved account ID, without the token. Terraform controls when that diagnostic is displayed and may render it after plan or apply output; do not rely on it as an approval checkpoint before mutations.

## mDNS reflector default

The following attributes no longer default to `false` when omitted:

- `cato_network_range.mdns_reflector`
- `cato_socket_site.native_range.mdns_reflector`

For supported range types, an omitted value allows the API's value to be reflected in Terraform state instead of always requesting `false`. If your configuration relies on disabling mDNS reflection, set `mdns_reflector = false` explicitly in the existing network range or native range block. Set it to `true` explicitly if enabling reflection is intended.

For a `cato_network_range` with `range_type = "Routed"`, omit `mdns_reflector`: the API does not support this setting for Routed ranges. Do not add an explicit `false` to those ranges. Review existing remote values and the upgrade plan before applying.

## Socket LAN firewall sub-rule placement

In the `at.position` setting of `cato_socket_lan_firewall_rule`, rename:

- `AFTER_RULE` to `AFTER_SUB_RULE`
- `BEFORE_RULE` to `BEFORE_SUB_RULE`

Keep `at.ref` set to the same sibling sub-rule ID. `FIRST_IN_RULE` and `LAST_IN_RULE` remain valid. These names apply to firewall sub-rules; do not replace placement values on top-level policy rules or other resources.

## TLS inspection category

`cato_tls_rule.rule.application.tls_inspect_category` no longer accepts `STREAMING_MEDIA`. Supported categories are `POPULAR_CLOUD_APPS` and `SAFE_TO_INSPECT_DOMAINS`.

If a configuration uses `STREAMING_MEDIA`, review the intended traffic match and replace that criterion with supported criteria that preserve the intended policy. There is no demonstrated one-to-one replacement: do not automatically substitute `SAFE_TO_INSPECT_DOMAINS` or simply remove the criterion, since either can change which traffic is inspected. Confirm the resulting match in the target account before applying. If the intended match cannot be expressed, retain the existing provider version until a suitable configuration is available.

## Group membership ownership

When `cato_group.members` is configured, refresh tracks those declared members rather than absorbing members managed elsewhere, including through `cato_group_members`. Imported groups can initially include all discovered members.

Review which resource owns each membership. If Terraform should manage an existing member through `cato_group`, declare it explicitly there. If a separate `cato_group_members` resource owns it, retain that ownership rather than duplicating it. Review the plan before updating a group whose previous state included members that were not declared in configuration.

## Removed prerelease policy resource types

If you used any of the 14 abbreviated policy resource types in a prerelease, update your configuration and any affected modules to use the [full-name replacements](https://registry.terraform.io/providers/catonetworks/cato/latest/docs/guides/policy_resource_type_migration).

- For the 11 bulk ordering types, use Terraform 1.8 or later, change the resource types and their references, and add a `moved` block for each affected address. Plan with v2.0.0 and verify Terraform reports the address moves without infrastructure changes.
- For `cato_pa_section`, `cato_remote_port_fwd_rule`, and `cato_remote_port_fwd_section`, first pin the provider to `= 2.0.0-prerelease.2`, which supports both names. Record the Cato IDs and follow the [configuration and import steps](https://registry.terraform.io/providers/catonetworks/cato/latest/docs/guides/policy_resource_type_migration#private-access-and-remote-port-forwarding) to bring the same objects under the full-name types. Complete the state migration before upgrading to v2.0.0.

## Other state considerations

Keep existing resource IDs and investigate any unexpected replacement before applying. Users of `1.5.0-prerelease.*` should also review the final resource schemas and examples. Subscription-group membership collections use lists in the final schema. Preserve the intended order and review the plan for changes to membership.

## Roll back

Before applying with v2, you can abandon the upgrade branch and restore the previous configuration and lock file. Run `terraform init -lockfile=readonly` to reinstall the locked provider and review the plan.

After applying with v2, do not assume that v1 can read all resulting state. Resources introduced in v2 have no v1.4.1 implementation, and prerelease schemas can differ. Plan rollback using the saved configuration, state backup, and an inventory of remote changes. Do not overwrite current state with an older snapshot without reconciling infrastructure changes made since that snapshot.

## 1.x maintenance status

The maintenance scope for the 1.x line has not yet been announced. Its end-of-support date has not yet been scheduled. This release does not establish a new support commitment or a retirement deadline for 1.x; confirm the published maintenance policy before relying on future 1.x fixes.
