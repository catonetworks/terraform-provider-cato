# Example: Basic Admin User Creation
resource "cato_admin" "basic_admin" {
  email                  = "admin@example.com"
  first_name             = "John"
  last_name              = "Doe"
  password_never_expires = true

  managed_roles = [
    {
      id = "role_id_1"
    }
  ]

  reseller_roles = [
    {
      id               = "role_id_2"
      allowed_accounts = ["account_id_1", "account_id_2"]
    }
  ]
}

# Example: Admin User with Optional Fields
resource "cato_admin" "optional_fields_admin" {
  email      = "optional@example.com"
  first_name = "Jane"
  last_name  = "Smith"

  # Optional attributes
  password_never_expires = false

  managed_roles = []

  reseller_roles = [
    {
      id = "reseller_role_id"
    }
  ]
}

# Example: Admin User Without Expiry
resource "cato_admin" "non_expiring_admin" {
  email                  = "noexpire@example.com"
  first_name             = "Alice"
  last_name              = "Jones"
  password_never_expires = true
}

# Example: Complete Configuration
resource "cato_admin" "complete_admin" {
  email                  = "complete@example.com"
  first_name             = "Bob"
  last_name              = "Brown"
  password_never_expires = true

  managed_roles = [
    {
      id = "role_id_3"
    }
  ]

  reseller_roles = [
    {
      id               = "another_role_id"
      allowed_accounts = ["account_id_3"]
    }
  ]
}