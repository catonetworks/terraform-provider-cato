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
resource "cato_remote_port_forward_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    description = "Managed by Terraform"
    enabled = false
    external_ip = { id = "replace-with-object-id" }
    external_port_range = {
      from = "443"
      to = "443"
    }
    forward_icmp = false
    internal_ip = "192.0.2.241"
    internal_port_range = {
      from = "443"
      to = "443"
    }
    name = "terraform-remote-port-forward-rule"
    remote_ips = {
      global_ip_range = [{ id = "replace-with-object-id" }]
      ip = ["192.0.2.100"]
      ip_range = [{
        from = "192.0.2.101"
        to = "192.0.2.102"
      }]
      subnet = ["198.51.100.0/24"]
    }
    restriction_type = "ALLOW_LIST"
    tracking = {
      enabled = false
      frequency = "DAILY"
      mailing_list = [{ id = "replace-with-object-id" }]
      subscription_group = [{ id = "replace-with-object-id" }]
      webhook = [{ id = "replace-with-object-id" }]
    }
  }
}
