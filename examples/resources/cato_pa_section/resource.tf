// NOTE: `cato_pa_section` is the abbreviated alias of `cato_private_access_section`.
// Prefer `cato_private_access_section` for new configs; this example mirrors it.

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
resource "cato_pa_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
    ref      = ""
  }
  section = {
    name = "terraform-private-access-section"
  }
}
