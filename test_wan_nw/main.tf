terraform {
  required_providers {
    cato = {
      source  = "terraform-providers/cato"
      version = "0.0.51"
    }
  }
}

provider "cato" {
  # Configure your provider here
}

resource "cato_wan_nw_rule" "kitchen_sink" {
  # This will be populated by import
}
