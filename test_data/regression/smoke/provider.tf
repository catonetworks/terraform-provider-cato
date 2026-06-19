terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
  required_version = ">= 1.5"
}

provider "cato" {
  baseurl    = var.baseurl
  token      = var.cato_token
  account_id = var.account_id
}
