// Create groups with members referenced by name
resource "cato_group" "group_by_name" {
  name        = "Cato Group"
  description = "Example group created via Terraform"

  # The following predicates can not be referenced by name today:
  # floating_subnet, host, network_interface, site_network_subnet
  members = [
    {
      id = "12345"
      type = "FLOATING_SUBNET"
    },
    {
      name = "My Global IP Range"
      type = "GLOBAL_IP_RANGE"
    },
    {
      id = "12345"
      type = "HOST"
    },
    {
      id = "12345"
      type = "NETWORK_INTERFACE"
    },
    {
      name = "X1600 Site"
      type = "SITE"
    },
    {
      id = "AbCdE12345="
      type = "SITE_NETWORK_SUBNET"
    },
  ]
}

// Create groups with members referenced by id
resource "cato_group" "group_by_id" {
  name        = "kitchen sink group by id"
  description = "Example group created via Terraform"

  members = [
    {
      id = "12345"
      type = "FLOATING_SUBNET"
    },
    {
      id = "12345"
      type = "GLOBAL_IP_RANGE"
    },
    {
      id = "12345"
      type = "HOST"
    },
    {
      id = "12345"
      type = "NETWORK_INTERFACE"
    },
    {
      id = "12345"
      type = "SITE"
    },
    {
      id = "AbCdE12345="
      type = "SITE_NETWORK_SUBNET"
    },
  ]
}
