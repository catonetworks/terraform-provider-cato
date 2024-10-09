---
subcategory: ""
page_title: "Manage Bulk Network Ranges from CSV"
description: |-
  Manage Bulk Network Ranges from CSV
---

# Manage Bulk Network Ranges from CSV

Terraform can natively import csv data using the [csvdecode](https://www.terraform.io/docs/language/functions/csvdecode.html) function. The following example shows how to use the csvdecode function to manage [network_ranges](https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange) resources in bulk from a csv file.

<details>
<summary>Example CSV file format</summary>

Create a csv file with the following format.  The first row is the header row and the remaining rows are the asset data.  The header row is used to map the column data to the asset attributes.

```csv
site_id,name,range_type,subnet,local_ip,gateway,vlan
98538,Net1Routed,Routed,10.0.1.0/24,,10.0.1.254,
98538,Net2VLAN,VLAN,10.0.2.0/24,,,2
98538,Net3Direct,Direct,10.0.3.0/24,10.0.1.5,,
```
</details>

## Example Bulk Import Usage

<details>
<summary>Example Variables for Bulk Import</summary>

## Example Variables for Bulk Import

```hcl
variable "csv_file_path" {
	description =  "Path to the csv file to import"
	type = string
	default = "network_ranges.csv"
}

```
</details>

## Proviers and Resources for Bulk Import

```hcl
locals {
	network_range_csv = csvdecode(file("${path.module}/${var.csv_file_path}"))
}

resource "cato-oss_network_range" "networks" {
    for_each = { for network_range in local.network_range_csv : network_range.subnet => network_range }
    site_id = each.value.site_id
    name = each.value.name
    range_type = each.value.range_type
    subnet = each.value.subnet
    local_ip = trimspace(each.value.local_ip) == "" ? null : each.value.local_ip
    gateway = trimspace(each.value.gateway) == "" ? null : each.value.gateway
    vlan = trimspace(each.value.vlan) == "" ? null : each.value.vlan
}
```
