// Add members to an existing group
resource "cato_group_members" "members" {
  group_name = "kitchen sink group"

  # Members can be referenced by ID or NAME
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
