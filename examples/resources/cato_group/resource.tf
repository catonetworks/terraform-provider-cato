// Create a group with members referenced by name
resource "cato_group" "group_by_name" {
  name        = "kitchen sink group by name"
  description = "Example group created via Terraform"

  members = [
    {
      # id   = "12345"
      name = "Floating Subnet 1"
      type = "FLOATING_SUBNET"
    },
    {
      # id   = "12345"
      name = "My Global IP Range"
      type = "GLOBAL_IP_RANGE"
    },
    {
      # id   = "12345"
      name = "DC01"
      type = "HOST"
    },
    {
      # id   = "12345"
      name = "X1600 Site:WAN1"
      type = "NETWORK_INTERFACE"
    },
    {
      # id   = "12345"
      name = "X1600 Site"
      type = "SITE"
    },
    {
      # id   = "AbCdE12345="
      name = "X1600 Site:LAN Subnet"
      type = "SITE_NETWORK_SUBNET"
    },
  ]
}

// Create a kitchen-sink group with members referenced by id
resource "cato_group" "group_by_id" {
  name        = "kitchen sink group by id"
  description = "Example group created via Terraform"

  members = [
    {
      id   = "12345"
      type = "FLOATING_SUBNET"
    },
    {
      id   = "12345"
      type = "GLOBAL_IP_RANGE"
    },
    {
      id   = "12345"
      type = "HOST"
    },
    {
      id   = "12345"
      type = "NETWORK_INTERFACE"
    },
    {
      id   = "12345"
      type = "SITE"
    },
    {
      id   = "AbCdE12345="
      type = "SITE_NETWORK_SUBNET"
    },
  ]
}
