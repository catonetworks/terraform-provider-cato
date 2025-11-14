---
subcategory: ""
page_title: "Manage Bulk Site Licenses from CSV"
description: |-
  Manage Bulk Site Licenses from CSV
---

# Manage Bulk Site Licenses from CSV

Terraform can natively import csv data using the [csvdecode](https://www.terraform.io/docs/language/functions/csvdecode.html) function. The following example shows how to use the csvdecode function to manage [site_license](https://registry.terraform.io/providers/catonetworks/cato/latest/docs/resources/license) resources in bulk from a csv file.

## IMPORTANT NOTE - License Migration

Important to note that changing a license using the terraform provider will not impact the service of a site.  If a site is migrated from one site license to another, or from a site license to a pooled license, site license is replaced immediately in the same operation and is no interruption in service to that site. 

### Export Site Licenses to CSV

Use the [catocli](https://github.com/catonetworks/cato-cli) to export site licenses to a csv file for ease of reference.

```bash
pip3 install catocli
catocli configure
catocli query licensing -f csv
```

<details>
<summary>Example CSV file format</summary>

Create a csv file with the following format.  The first row is the header row and the remaining rows are the asset data.  The header row is used to map the column data to the asset attributes.

```csv
site_id,license_id,license_bw
441577,ec4c7c13-08bf-420e-ac92-a8b0bc16a4ce,
443224,583cfd38-f268-434e-bbe2-1b0874173e2e
442855,2f10af7d-2471-492c-bf68-bc4b9e2b1211,50
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
	default = "licensing.csv"
}

```
</details>

## Proviers and Resources for Bulk Import

```hcl
locals {
    license_csv = csvdecode(file("${path.module}/${var.csv_file_path}"))
}

resource "cato_license" "licenses" {
  for_each = { for license in local.license_csv : license.site_id => license }
  site_id    = each.value.site_id
  license_id = each.value.license_id
  bw         = each.value.license_bw
}
```
