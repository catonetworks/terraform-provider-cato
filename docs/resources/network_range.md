---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cato_network_range Resource - terraform-provider-cato"
subcategory: ""
description: |-
  The cato_network_range resource contains the configuration parameters necessary to add a network range to a cato site. (virtual socket in AWS/Azure, or physical socket https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites). Documentation for the underlying API used in this resource can be found at mutation.addNetworkRange() https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange.
---

# cato_network_range (Resource)

The `cato_network_range` resource contains the configuration parameters necessary to add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()](https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange).

## Example Usage

```terraform
// network range of type VLAN
resource "cato_network_range" "vlan100" {
  site_id    = cato_socket_site.site1.id
  name       = "VLAN100"
  range_type = "VLAN"
  subnet     = "192.168.100.0/24"
  local_ip   = "192.168.100.100"
  vlan       = "100"
}

// network range of type VLAN with DHCP RANGE
resource "cato_network_range" "vlan200" {
  site_id    = cato_socket_site.site1.id
  name       = "VLAN200"
  range_type = "VLAN"
  subnet     = "192.168.200.0/24"
  local_ip   = "192.168.200.1"
  vlan       = "200"
  dhcp_settings = {
    dhcp_type = "DHCP_RANGE"
    ip_range  = "192.168.200.100-192.168.200.150"
  }
}

// routed network 
resource "cato_network_range" "routed250" {
  site_id    = cato_socket_site.site1.id
  name       = "routed250"
  range_type = "Routed"
  subnet     = "192.168.250.0/24"
  gateway   = "192.168.25.1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Network range name
- `range_type` (String) Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)
- `site_id` (String) Site ID
- `subnet` (String) Network range (CIDR)

### Optional

- `dhcp_settings` (Attributes) Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type) (see [below for nested schema](#nestedatt--dhcp_settings))
- `gateway` (String) Network range gateway (Only releveant for Routed range_type)
- `interface_id` (String) Network Interface ID
- `internet_only` (Boolean) Internet only network range (Only releveant for Routed range_type)
- `local_ip` (String) Network range local ip
- `translated_subnet` (String) Network range translated native IP range (CIDR)
- `vlan` (Number) Network range VLAN ID (Only releveant for VLAN range_type)

### Read-Only

- `id` (String) Network Range ID

<a id="nestedatt--dhcp_settings"></a>
### Nested Schema for `dhcp_settings`

Required:

- `dhcp_type` (String) Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)

Optional:

- `ip_range` (String) Network range dhcp range (format "192.168.1.10-192.168.1.20")
- `relay_group_id` (String) Network range dhcp relay group id
