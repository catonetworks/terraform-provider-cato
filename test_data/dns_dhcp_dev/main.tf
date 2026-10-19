locals {
  baseline_target = fileexists("${path.module}/baseline.json") ? jsondecode(file("${path.module}/baseline.json")).target : {
    account  = ""
    site     = ""
    endpoint = ""
  }
}

terraform {
  required_version = ">= 1.9"
  required_providers {
    cato = { source = "catonetworks/cato" }
  }
}

provider "cato" {
  baseurl    = var.baseurl
  account_id = var.account_id
  retry_max  = 0
}

variable "baseurl" {
  type = string
  validation {
    condition     = var.baseurl == "https://system.cc.dev-us1-m1.catonet.works/api/v1/graphql"
    error_message = "Local fixture is restricted to dev-us1-m1 GraphQL."
  }
}
variable "account_id" {
  type = string
  validation {
    condition     = var.account_id == local.baseline_target.account
    error_message = "Terraform target must match captured baseline. Run preflight capture for intended account/site."
  }
}
variable "site_id" {
  type = string
  validation {
    condition     = var.site_id == local.baseline_target.site
    error_message = "Terraform target must match captured baseline. Run preflight capture for intended account/site."
  }
}
variable "primary_server" { type = string }
variable "secondary_server" { type = string }
variable "suffix" { type = list(string) }
variable "accept_dns_requests_on_lan_interface_ip" { type = bool }
variable "lease_time" { type = number }
variable "relay_enabled" { type = bool }
variable "relay_timeout" { type = number }
variable "relay_group_id" { type = string }
variable "test_name" { type = string }
variable "forwarding_domain" { type = string }
variable "test_servers" { type = list(string) }
variable "dhcp_option_tag" { type = number }
variable "dhcp_option_value" { type = string }

resource "cato_dns_settings" "account" {
  primary_server                          = var.primary_server
  secondary_server                        = var.secondary_server
  suffix                                  = var.suffix
  accept_dns_requests_on_lan_interface_ip = var.accept_dns_requests_on_lan_interface_ip
}

resource "cato_dns_site_settings" "site" {
  site_id          = var.site_id
  primary_server   = var.primary_server
  secondary_server = var.secondary_server
  suffix           = var.suffix
}

resource "cato_dns_forwarding_rule" "test" {
  domain = var.forwarding_domain
  server = var.test_servers
}

resource "cato_dhcp_settings" "account" {
  lease_time     = var.lease_time
  relay_enabled  = var.relay_enabled
  relay_timeout  = var.relay_timeout
  relay_group_id = var.relay_group_id
}

resource "cato_dhcp_option" "test" {
  tag         = var.dhcp_option_tag
  type        = "ASCII"
  value       = var.dhcp_option_value
  description = var.test_name
}

resource "cato_dhcp_relay_group" "test" {
  name   = var.test_name
  server = var.test_servers
}

output "dns_forwarding_rule_id" { value = cato_dns_forwarding_rule.test.id }
output "dhcp_option_id" { value = cato_dhcp_option.test.id }
output "dhcp_relay_group_id" { value = cato_dhcp_relay_group.test.id }
