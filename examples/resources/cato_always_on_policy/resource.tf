# Enable or disable always-on policy
resource "cato_always_on_policy" "this" {
  enabled = true
}


# Audit and active draft information are computed by the existing policy resource.
output "always_on_policy_metadata" {
  value = {
    published_time = cato_always_on_policy.this.published_time
    published_by   = cato_always_on_policy.this.published_by
    revision_id    = cato_always_on_policy.this.revision_id
    revision_name  = cato_always_on_policy.this.revision_name
  }
}

# Terraform provider configuration
terraform {
  required_providers {
    cato = {
      source = "registry.terraform.io/catonetworks/cato"
    }
  }
}

variable "account_id" {
  description = "Cato account ID"
  type        = string
}

variable "cato_token" {
  description = "Cato API token"
  type    = string
}

variable "baseurl" {
  description = "Cato API base URL"
  type    = string
}

provider "cato" {
  baseurl    = var.baseurl
  account_id = var.account_id
  token      = var.cato_token
}
