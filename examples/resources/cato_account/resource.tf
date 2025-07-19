# Example 1: Basic Account Creation
# This example creates a basic Cato account with minimal required configuration

resource "cato_account" "example_basic" {
  name     = "My-Customer-Account"
  timezone = "America/New_York"
}

# Example 2: Complete Account Configuration
# This example shows all available configuration options

resource "cato_account" "example_complete" {
  name        = "Enterprise-Customer-Account"
  description = "Primary customer account for Enterprise Corp"
  tenancy     = "SINGLE_TENANT"
  timezone    = "Europe/London"
  type        = "CUSTOMER"
}

# Example 3: Multi-Tenant Account
# This example creates a multi-tenant account

resource "cato_account" "example_multi_tenant" {
  name        = "Multi-Tenant-Account"
  description = "Account supporting multiple tenants"
  tenancy     = "MULTI_TENANT"
  timezone    = "Asia/Tokyo"
  type        = "CUSTOMER"
}

# Example 4: Partner Account
# This example creates a partner account

resource "cato_account" "example_partner" {
  name        = "Partner-Account"
  description = "Account for partner organization"
  tenancy     = "SINGLE_TENANT"
  timezone    = "Pacific/Auckland"
  type        = "PARTNER"
}

