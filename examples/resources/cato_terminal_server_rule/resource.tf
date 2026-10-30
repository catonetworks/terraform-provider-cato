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
resource "cato_terminal_server_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    allowed_host_ip = { id = "replace-with-object-id" }
    description     = "Managed by Terraform"
    enabled         = true
    exclude_traffic = [{ id = "replace-with-object-id" }]
    name            = "terraform-terminal-server-rule"
  }
}
