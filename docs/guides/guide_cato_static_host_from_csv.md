---
subcategory: ""
page_title: "Manage Bulk Static Hosts from CSV"
description: |-
  Manage Bulk Static Hosts from CSV
---

# Manage Bulk Static Hosts from CSV

Terraform can natively import csv data using the [csvdecode](https://www.terraform.io/docs/language/functions/csvdecode.html) function. The following example shows how to use the csvdecode function to manage [static hosts](https://api.catonetworks.com/documentation/#mutation-site.addStaticHost) resources in bulk from a csv file.

<details>
<summary>Example CSV file format</summary>

Create a csv file with the following format.  The first row is the header row and the remaining rows are the asset data.  The header row is used to map the column data to the asset attributes.

```csv
ip,name,mac_address
192.168.25.25,my.hostname25,
192.168.25.26,my.hostname26,"00:00:00:00:00:02"
```
</details>

## Example Bulk Import Usage

<details>
<summary>Example Variables for Bulk Import</summary>

## Example Variables for Bulk Import

```
variable "site_id" {
    description = "Site ID"
    type        = number
    default	    = 12345
}

variable "csv_file_path" {
	description =  "Path to the csv file to import"
	type = string
	default = "static_hosts.csv"
}

```
</details>

## Proviers and Resources for Bulk Import

```hcl
locals {
	static_host_csv = csvdecode(file("${path.module}/${var.csv_file_path}"))
}

resource "cato_static_host" "host_csv" {
    for_each = { for static_host in local.static_host_csv : static_host.ip => static_host }
    site_id = var.site_id
    name = each.value.name
    ip = each.value.ip
    mac_address = each.value.mac_address
}
```
