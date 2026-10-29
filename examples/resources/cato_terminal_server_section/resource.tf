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
resource "cato_terminal_server_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
    # ref = "reference-section-id"
  }
  section = {
    name = "terraform-terminal-server-section"
  }
}
