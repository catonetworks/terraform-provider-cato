terraform {
  required_providers {
    cato = { source = "catonetworks/cato" }
  }
}

variable "token" {
  type        = string
  sensitive   = true
  description = "API token supplied through TF_VAR_token."
}

variable "account_id" {
  type        = string
  description = "Cato account ID supplied through TF_VAR_account_id."
}

provider "cato" {
  token      = var.token
  account_id = var.account_id
}

# Changes are automatically published. Use a test account when evaluating beta APIs.
resource "cato_split_tunnel_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    action      = "ROUTE_ONLY_SELECTED"
    country     = [{ id = "replace-with-object-id" }]
    coverage    = "ALL"
    description = "Managed by Terraform"
    destination = {
      application     = [{ id = "replace-with-object-id" }]
      domain          = ["example.com"]
      fqdn            = ["example.com"]
      global_ip_range = [{ id = "replace-with-object-id" }]
    }
    dns_exclusion = {
      domain = ["example.com"]
    }
    enabled          = true
    name             = "terraform-split-tunnel-rule"
    platform         = ["WINDOWS"]
    routing_priority = "LAN"
    source = {
      user        = [{ id = "replace-with-object-id" }]
      users_group = [{ id = "replace-with-object-id" }]
    }
    source_network = {
      source_network_type = "ANY"
    }
  }
}
