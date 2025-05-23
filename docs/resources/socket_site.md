---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cato_socket_site Resource - terraform-provider-cato"
subcategory: ""
description: |-
  The cato_socket_site resource contains the configuration parameters necessary to add a socket site to the Cato cloud (virtual socket in AWS/Azure, or physical socket https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites). Documentation for the underlying API used in this resource can be found at mutation.addSocketSite() https://api.catonetworks.com/documentation/#mutation-site.addSocketSite.
  Note: For AWS deployments, please accept the EULA for the Cato Networks AWS Marketplace product https://aws.amazon.com/marketplace/pp?sku=dvfhly9fuuu67tw59c7lt5t3c.
---

# cato_socket_site (Resource)

The `cato_socket_site` resource contains the configuration parameters necessary to add a socket site to the Cato cloud ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite). 

 **Note**: For AWS deployments, please accept the [EULA for the Cato Networks AWS Marketplace product](https://aws.amazon.com/marketplace/pp?sku=dvfhly9fuuu67tw59c7lt5t3c).

## Example Usage

```terraform
// Data Source for site location
data "cato_siteLocation" "ny" {
  filters = [{
    field = "city"
    search = "New York City"
    operation = "startsWith"
  },
  {
    field = "state_name"
    search = "New York"
    operation = "exact"
  },
 {
    field = "country_name"
    search = "United States"
    operation = "contains"
  }]
}

// socket site for AWS
resource "cato_socket_site" "aws_site" {
  name            = "aws_site"
  description     = "site description"
  site_type       = "DATACENTER"
  connection_type = "SOCKET_AWS1500"

  native_range = {
    native_network_range = "192.168.25.0/24"
    local_ip             = "192.168.25.5"
  }

  site_location = {
    city = data.cato_siteLocation.ny.locations[0].city
    country_code = data.cato_siteLocation.ny.locations[0].country_code
    state_code = data.cato_siteLocation.ny.locations[0].state_code
    timezone = data.cato_siteLocation.ny.locations[0].timezone[0]
    address = "555 That Way"
  }
}

// socket site x1500 with DHCP settings
resource "cato_socket_site" "branch_site" {
  name            = "branch_site"
  description     = "site description"
  site_type       = "BRANCH"
  connection_type = "SOCKET_X1500"

  native_range = {
    native_network_range = "192.168.20.0/24"
    local_ip             = "192.168.20.1"
    dhcp_settings = {
      dhcp_type = "DHCP_RANGE"
      ip_range  = "192.168.20.10-192.168.20.22"
    }
  }

  site_location = {
    country_code = "FR"
    timezone     = "Europe/Paris"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connection_type` (String) Connection type for the site (SOCKET_X1500, SOCKET_AWS1500, SOCKET_AZ1500, ...)
- `name` (String) Site name
- `native_range` (Attributes) Site native range settings (see [below for nested schema](#nestedatt--native_range))
- `site_location` (Attributes) Site location (see [below for nested schema](#nestedatt--site_location))
- `site_type` (String) Site type (https://api.catonetworks.com/documentation/#definition-SiteType)

### Optional

- `description` (String) Site description

### Read-Only

- `id` (String) Site ID

<a id="nestedatt--native_range"></a>
### Nested Schema for `native_range`

Required:

- `local_ip` (String) Site native range local ip
- `native_network_range` (String) Site native IP range (CIDR)

Optional:

- `dhcp_settings` (Attributes) Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type) (see [below for nested schema](#nestedatt--native_range--dhcp_settings))
- `native_network_range_id` (String) Site native IP range ID (for update purpose)
- `translated_subnet` (String) Site translated native IP range (CIDR)

<a id="nestedatt--native_range--dhcp_settings"></a>
### Nested Schema for `native_range.dhcp_settings`

Required:

- `dhcp_type` (String) Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)

Optional:

- `ip_range` (String) Network range dhcp range (format "192.168.1.10-192.168.1.20")
- `relay_group_id` (String) Network range dhcp relay group id



<a id="nestedatt--site_location"></a>
### Nested Schema for `site_location`

Required:

- `country_code` (String) Site country code (can be retrieve from entityLookup)
- `timezone` (String) Site timezone (can be retrieve from entityLookup)

Optional:

- `address` (String) Optionnal address
- `city` (String) Optionnal city
- `state_code` (String) Optionnal site state code(can be retrieve from entityLookup)
