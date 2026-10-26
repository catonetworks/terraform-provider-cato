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
