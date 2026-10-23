data "cato_permission_catalog" "available" {
  # account_id = "123456"
}

output "grantable_resources" {
  value = data.cato_permission_catalog.available.resources
}
