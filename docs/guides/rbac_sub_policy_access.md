---
subcategory: ""
page_title: "Connect RBAC Roles to Firewall Sub-Policies"
description: |-
  Assign user or group roles scoped to WAN and Internet Firewall sub-policies.
---

# Connect RBAC Roles to Firewall Sub-Policies

Use `cato_rbac_assignment` to connect a principal, a role, and the sub-policies
that the principal may access:

- `cato_role.permissions` defines capabilities, using resource names such as
  `WanFirewall` and `InternetFirewall` and actions such as `VIEW` or `EDIT`.
- `cato_rbac_assignment.role_ids` references the role's `role_id`.
- `cato_rbac_assignment.allowed_entities` references sub-policy IDs, with types
  `WAN_FIREWALL_POLICY` or `INTERNET_FIREWALL_POLICY` and the intended action.
- A firewall rule's `sub_policy_id` attaches the rule to its sub-policy.

The role contains the permission definitions. The assignment carries the
connection to the sub-policies. Terraform references establish creation and
destruction dependencies automatically.

## Prerequisites

Use an existing user in the same account, or replace `user_id` with `group_id`
to assign a users group. A user ID identifies a user, rather than a legacy
admin principal. The provider does not create the user in this example.

Set `CATO_TOKEN`, `TF_VAR_account_id`, and `TF_VAR_user_id` in your environment.
The API token must be authorized to manage roles, assignments, and both firewall
policy families. Discover available permission resource names and actions with
`cato_permission_catalog` when adapting the role.

## Complete configuration

The example creates two disabled sub-policies and disabled rules scoped to a
documentation IP address. It grants the selected user `VIEW` access to both
sub-policies. Adapt the scope and enable the policies and rules when ready to
use them.

```terraform
# Connect a user to custom role permissions on specific firewall sub-policies.
# Set CATO_TOKEN, TF_VAR_account_id, and TF_VAR_user_id before running Terraform.
terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
}

variable "account_id" {
  description = "Cato account that owns the user, roles, and sub-policies."
  type        = string
}

variable "user_id" {
  description = "Existing user ID in this account; do not use a legacy admin ID."
  type        = string
}

provider "cato" {
  account_id = var.account_id
  # baseurl = "https://api.catonetworks.com/api/v1/graphql2"
  # Authentication is read from CATO_TOKEN.
}

# 1. Create the policies that will define the assignment's scope.
resource "cato_wf_sub_policy" "wan" {
  name        = "Terraform RBAC WAN example"
  description = "WAN rules accessible to the assigned user"
  at = {
    position = "LAST_IN_POLICY"
  }
  scope = {
    enabled     = false # Enable when ready to enforce this sub-policy.
    application = {}
    direction   = "TO"
    source      = { ip = ["192.0.2.123"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}

resource "cato_if_sub_policy" "internet" {
  name        = "Terraform RBAC Internet example"
  description = "Internet rules accessible to the assigned user"
  at = {
    position = "LAST_IN_POLICY"
  }
  scope = {
    enabled           = false
    connection_origin = "REMOTE"
    source            = { ip = ["192.0.2.123"] }
    destination       = {}
    tracking          = { event = { enabled = false } }
  }
}

# 2. Place rules inside the sub-policies using sub_policy_id.
resource "cato_wf_rule" "wan" {
  sub_policy_id = cato_wf_sub_policy.wan.id
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name = "Terraform RBAC WAN rule"
    # description = "Rule owned by the WAN sub-policy"
    enabled     = false
    action      = "BLOCK"
    direction   = "TO"
    application = {}
    source      = { ip = ["192.0.2.123"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}

resource "cato_if_rule" "internet" {
  sub_policy_id = cato_if_sub_policy.internet.id
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name = "Terraform RBAC Internet rule"
    # description = "Rule owned by the Internet sub-policy"
    enabled     = false
    action      = "BLOCK"
    source      = { ip = ["192.0.2.123"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}

# 3. Define what the user may do. The role does not contain policy IDs.
resource "cato_role" "viewer" {
  name = "Terraform firewall policy viewer"
  # description = "View the assigned firewall sub-policies"
  # account_id = var.account_id # Defaults to the provider account.
  permissions = [
    { resource = "WanFirewall", action = "VIEW" },
    { resource = "InternetFirewall", action = "VIEW" },
  ]
}

# 4. Connect the principal, role, and policy scope through one assignment.
resource "cato_rbac_assignment" "viewer" {
  user_id = var.user_id
  # group_id = "67890" # Use instead of user_id for a users-group principal.
  # account_id = var.account_id # Defaults to the provider account.
  # description = "View only the two example firewall sub-policies"
  # allow_viewing_unobfuscated_data = false

  role_ids = [cato_role.viewer.role_id]
  allowed_entities = [
    {
      id     = cato_wf_sub_policy.wan.id
      type   = "WAN_FIREWALL_POLICY"
      action = "VIEW"
    },
    {
      id     = cato_if_sub_policy.internet.id
      type   = "INTERNET_FIREWALL_POLICY"
      action = "VIEW"
    },
  ]
}
```

## Apply and check drift

```shell
terraform init
terraform plan -out=rbac.tfplan
terraform apply rbac.tfplan
terraform plan -detailed-exitcode
```

After apply, a clean plan exits with `0`. A plan that detects changes exits
with `2`; errors exit with `1`. If a role's permissions or an assignment's roles
or policy grants are changed outside Terraform, review the next plan and apply
it to restore the configuration.

Permissions, role IDs, and entity scopes are unordered sets. Reordering entries
should produce no changes. Updates replace the configured collections, so
removing a grant removes it on apply. Removing a role or assignment description
clears it.

For edit access, configure `EDIT` in both the role permission and the corresponding
policy entity grant. These examples describe provisioning and scoping; validate
the selected user's effective access separately before relying on it.

## Cleanup

```shell
terraform destroy
```

This removes the example's assignments, role, rules, and sub-policies. The
existing user or group remains.
