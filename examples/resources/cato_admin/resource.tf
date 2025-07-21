
# Example: Basic admin user creation
resource "cato_admin" "basic_admin" {
  email                  = "admin@email.com"
  first_name             = "John"
  last_name              = "Doe"
  password_never_expires = true

  managed_roles = [
    {
      id = "1"
    }
  ]
}

# Example: Admin user creation for MSP in Sub Account from Reseller API key
resource "cato_admin" "basic_admin" {
  account_id             = "12345" // Sub account ID for MSP use
  email                  = "admin@email.com"
  first_name             = "John"
  last_name              = "Doe"
  password_never_expires = true

  managed_roles = [
    {
      id = "1"
    }
  ]
}


# Reseller admin user example
resource "cato_admin" "reseller_admin" {
  email                  = "reseller@email.com"
  first_name             = "John"
  last_name              = "Doe"
  password_never_expires = true

  managed_roles = [
    {
      id = "1"
      allowed_accounts = ["1234"]
    }
  ]

  reseller_roles = [
    {
      id = "4"
    }
  ]
}