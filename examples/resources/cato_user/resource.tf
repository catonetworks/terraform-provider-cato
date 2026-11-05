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

resource "cato_user" "example" {
 email = "terraform-user@example.invalid"
 first_name = "Terraform"
 last_name = "Example"
 # enabled = false # Defaults to true; set false to disable the local user.
 # department = "Engineering"
 # job_title = "Developer"
 # phone_number = "+15555550123"
}
