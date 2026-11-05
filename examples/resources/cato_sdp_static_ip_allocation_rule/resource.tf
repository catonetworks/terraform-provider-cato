# The IP must belong to a static IP range configured on the account.
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
resource "cato_sdp_static_ip_allocation_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    description = "Managed by Terraform"
    enabled = true
    ip = "192.0.2.241"
    name = "terraform-sdp-static-ip-allocation-rule"
    source = {
      user = { id = "replace-with-object-id" }
    }
  }
}
