# Example usage:
#   export TF_VAR_token="your-api-token"
#   export TF_VAR_account_id="123456"
#   terraform init
#   terraform plan

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
  #retry_max              = 3
  #retry_wait_min_seconds = 10
  #retry_wait_max_seconds = 30
}

resource "cato_role" "example" {
  name = "tf-doc-example-role"
  # account_id  = var.account_id
  # description = "Read-only site access"
  permissions = [
    { resource = "Sites", action = "VIEW" },
  ]
}

output "role_id" {
  value = cato_role.example.role_id
}

# Import a custom role using ACCOUNT_ID:ROLE_ID:
# terraform import cato_role.example 123456:98765
