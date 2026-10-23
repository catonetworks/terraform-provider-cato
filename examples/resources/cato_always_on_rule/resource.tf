# Create support is implemented. Read, update, and delete are still under development.
# The rule is saved in the draft Always On policy; this resource does not publish it.

terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
}

variable "token" {
  description = "Cato API token (set with TF_VAR_token)."
  type        = string
  sensitive   = true
}

variable "account_id" {
  description = "Cato account ID (set with TF_VAR_account_id)."
  type        = string
}

variable "baseurl" {
  description = "Cato GraphQL API endpoint."
  type        = string
  default     = "https://api.catonetworks.com/api/v1/graphql2"
}

provider "cato" {
  baseurl    = var.baseurl
  token      = var.token
  account_id = var.account_id
}

resource "cato_always_on_rule" "example" {
  # at = {
  #   position = "LAST_IN_POLICY"
  #   ref      = "section-or-rule-id" # Required for positions relative to a section or rule.
  # }

  rule = {
    name              = "Example Always On rule"
    description       = "Enforce Always On for Windows and macOS devices"
    enabled           = true
    action            = "ENFORCE"
    allow_fail_open   = false
    allow_user_bypass = true
    anti_tamper_mode  = "OFF"

    bypass_duration = {
      time = 30
      unit = "MINUTES"
    }

    platform = ["WINDOWS", "MACOS"]

    # Supply either id or name for each reference. Empty sets match all entities.
    device_posture_profile = [
      # { name = "Managed devices" }, # Alternatively: { id = "device-profile-id" }
    ]
    source = {
      user = [
        # { name = "alice@example.com" }, # Alternatively: { id = "user-id" }
      ]
      users_group = [
        # { name = "Employees" }, # Alternatively: { id = "user-group-id" }
      ]
    }
  }
}
