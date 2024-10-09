---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cato-oss_wan_interface Resource - terraform-provider-cato-oss"
subcategory: ""
description: |-
  The cato-oss_wan_interface resource contains the configuration parameters necessary to add a wan interface to a socket. (virtual socket in AWS/Azure, or physical socket https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites). Documentation for the underlying API used in this resource can be found at mutation.updateSocketInterface() https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface.
---

# cato-oss_wan_interface (Resource)

The `cato-oss_wan_interface` resource contains the configuration parameters necessary to add a wan interface to a socket. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).

## Example Usage

```terraform
// wan interface active
resource "cato-oss_wan_interface" "wan1" {
  site_id              = cato-oss_socket_site.site1.id
  interface_id         = "WAN1"
  name                 = "Interface WAN 1"
  upstream_bandwidth   = "100"
  downstream_bandwidth = "100"
  role                 = "wan_1"
  precedence           = "ACTIVE"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `downstream_bandwidth` (Number) WAN interface downstream bandwitdh
- `interface_id` (String) SocketInterface available ids, INT_# stands for 1,2,3...12 supported ids (https://api.catonetworks.com/documentation/#definition-SocketInterfaceIDEnum)
- `name` (String) WAN interface name
- `precedence` (String) WAN interface precedence (https://api.catonetworks.com/documentation/#definition-SocketInterfacePrecedenceEnum)
- `role` (String) WAN interface role (https://api.catonetworks.com/documentation/#definition-SocketInterfaceRole)
- `site_id` (String) Site ID
- `upstream_bandwidth` (Number) WAN interface upstream bandwitdh