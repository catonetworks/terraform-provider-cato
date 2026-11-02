# Enable or disable always-on policy
resource "cato_always_on_policy" "this" {
  enabled = "true"
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
