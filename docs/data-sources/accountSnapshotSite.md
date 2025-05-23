---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cato_accountSnapshotSite Data Source - terraform-provider-cato"
subcategory: ""
description: |-
  Retrieves account snapshot site socket serial number and primary status information.
---

# cato_accountSnapshotSite (Data Source)

Retrieves account snapshot site socket serial number and primary status information.

## Example Usage

```terraform
## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_accountSnapshotSite" "aws-dev-site" {
	id = var.site_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Identifier for the site

### Read-Only

- `info` (Attributes) (see [below for nested schema](#nestedatt--info))

<a id="nestedatt--info"></a>
### Nested Schema for `info`

Read-Only:

- `name` (String) Site Name
- `sockets` (Attributes List) List of sockets attached to the site (see [below for nested schema](#nestedatt--info--sockets))

<a id="nestedatt--info--sockets"></a>
### Nested Schema for `info.sockets`

Read-Only:

- `id` (String) Socket id
- `is_primary` (Boolean) Socket is primary
- `serial` (String) Socket serial number
