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
  baseurl                = var.baseurl
  token                  = var.token
  account_id             = var.account_id
  #retry_max              = 3
  #retry_wait_min_seconds = 10
  #retry_wait_max_seconds = 30
}

resource "cato_socket_site" "example" {
  name            = "tf-doc-example-socket-site"
  description     = "Socket site managed by Terraform example"
  site_type       = "BRANCH"
  connection_type = "SOCKET_X1500"

  native_range = {
    native_network_range = "192.168.20.0/24"
    local_ip             = "192.168.20.1"
    interface_dest_type  = "LAN"
    # vlan               = 20
    # mdns_reflector     = true
    # translated_subnet  = "10.20.20.0/24"
    # lag_min_links      = 2

    dhcp_settings = {
      dhcp_type = "DHCP_RANGE"
      ip_range  = "192.168.20.10-192.168.20.50"
      # dhcp_microsegmentation = true
    }
  }

  site_location = {
    country_code = "FR"
    timezone     = "Europe/Paris"
    # city       = "Paris"
    # state_code = "IDF"
    # address    = "1 Example Street"
  }
}
