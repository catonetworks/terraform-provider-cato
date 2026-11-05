terraform {
  required_providers {
    cato = { source = "catonetworks/cato" }
  }
}

variable "token" {
  type = string
  sensitive = true
  description = "API token supplied through TF_VAR_token."
}

variable "account_id" {
  type = string
  description = "Cato account ID supplied through TF_VAR_account_id."
}

provider "cato" {
  token = var.token
  account_id = var.account_id
}

# Changes are automatically published. Use a test account when evaluating beta APIs.
resource "cato_dynamic_ip_allocation_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    country = [{ id = "replace-with-object-id" }]
    description = "Managed by Terraform"
    device_posture_profile = [{ id = "replace-with-object-id" }]
    enabled = true
    name = "terraform-dynamic-ip-allocation-rule"
    platform = ["WINDOWS"]
    range = {
      global_ip_range = { id = "replace-with-object-id" }
    }
    source = {
      user = [{ id = "replace-with-object-id" }]
      users_group = [{ id = "replace-with-object-id" }]
    }
  }
}
