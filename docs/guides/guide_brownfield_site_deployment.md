---
subcategory: ""
page_title: "Cato Site Brownfield Deployment"
description: |-
  Cato Site Brownfield Deployment
---

# Cato Site Brownfield Deployment

## Overview

Brownfield deployments enable you to bring existing Cato infrastructure under Terraform management without recreating resources or causing downtime. This guide walks you through exporting existing Cato Network site configurations and importing them into Terraform state for infrastructure-as-code management.

### What is a Brownfield Deployment?

- **Greenfield**: Building infrastructure from scratch with no existing resources
- **Brownfield**: Managing existing infrastructure already running in production (e.g., sites configured in the Cato Management Application)

### Benefits

- **Version Control**: Track infrastructure changes in Git
- **Consistency**: Standardized configurations across all sites
- **Automation**: Enable CI/CD pipelines for network infrastructure
- **Disaster Recovery**: Quick restoration from configuration backups
- **Bulk Updates**: Modify multiple sites simultaneously

## Prerequisites

Install the following tools:

```bash
# Verify installations
python3 --version      # Python 3.6 or later
terraform --version    # Terraform 0.13 or later
pip3 install catocli   # Cato CLI tool
git --version          # Git (recommended)
```

### Cato API Credentials

You'll need:
- **API Token**: Generated from the Cato Management Application (see [Generating API Keys](https://support.catonetworks.com/hc/en-us/articles/360011711818))
- **Account ID**: Found in Account > Account Info or in the CMA URL

## Brownfield Workflow

The workflow consists of four phases:

1. **Export**: CMA → catocli → CSV/JSON files
2. **Import**: CSV/JSON files → Terraform State
3. **Modify**: Edit CSV/JSON files with desired changes (optional)
4. **Manage**: Terraform State → Apply → Update CMA

## Implementation Steps

### Step 1: Configure Cato CLI

```bash
# Interactive configuration
catocli configure

# Or use environment variables
export CATO_TOKEN="your-api-token-here"
export CATO_ACCOUNT_ID="your-account-id"

# Verify configuration
catocli configure show
catocli entity site
```

### Step 2: Create Project Directory

```bash
mkdir cato-brownfield-deployment
cd cato-brownfield-deployment
git init  # Optional but recommended
```

### Step 3: Set Up Terraform Configuration

Create `main.tf`:

```hcl
terraform {
  required_version = ">= 0.13"
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = "~> 0.0.58"
    }
  }
}

provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}
```

### Step 4: Export Existing Sites

#### CSV Format (recommended for Excel/Sheets editing)

```bash
catocli export socket_sites \
  -f csv \
  --output-directory=config_data_csv
```

This creates:
- `socket_sites.csv` - Main site configuration
- `sites_config/{site_name}_network_ranges.csv` - Per-site network ranges

#### JSON Format (recommended for programmatic manipulation)

```bash
catocli export socket_sites \
  -f json \
  --output-directory=config_data
```

### Step 5: Add Module to Terraform

#### For CSV:

```hcl
module "sites_from_csv" {
  source = "catonetworks/socket-bulk-sites/cato"
  
  sites_csv_file_path = "config_data_csv/socket_sites.csv"
  sites_csv_network_ranges_folder_path = "config_data_csv/sites_config/"
}
```

#### For JSON:

```hcl
module "sites_from_json" {
  source = "catonetworks/socket-bulk-sites/cato"
  
  sites_json_file_path = "config_data/socket_sites.json"
}
```

### Step 6: Import into Terraform State

```bash
# Initialize Terraform
terraform init

# Import existing resources (CSV example)
catocli import socket_sites_to_tf \
  --data-type csv \
  --csv-file config_data_csv/socket_sites.csv \
  --csv-folder config_data_csv/sites_config/ \
  --module-name module.sites_from_csv \
  --auto-approve

# Verify import (should show no changes)
terraform plan
```

## Best Practices

### 1. Version Control Everything

```bash
git add main.tf config_data_csv/
git commit -m "Initial brownfield import"
```

### 2. Regular Backups

Create automated backup script:

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="backups/$DATE"
mkdir -p "$BACKUP_DIR"
catocli export socket_sites -f json --output-directory="$BACKUP_DIR"
```

### 3. Start Small

- Begin with a single site to validate the process
- Test with `terraform plan` before applying
- Scale to multiple sites after validation

## Troubleshooting

### Import Fails with "Resource Already Exists"

```bash
# List and remove resource from state
terraform state list
terraform state rm 'module.sites_from_csv.cato_socket_site["Site Name"]'
```

### Plan Shows Unexpected Changes

```bash
# Export fresh configuration and compare
catocli export socket_sites -f json --output-directory=config_data_verify
diff config_data/socket_sites.json config_data_verify/socket_sites.json
```

## Additional Resources

- [Cato API Best Practices - Brownfield Deployments](https://connect.catonetworks.com/kb/api-best-practices/brownfield-deployments-for-cato-network-sites/1511)
- [Socket-Bulk-Sites Terraform Module](https://registry.terraform.io/modules/catonetworks/socket-bulk-sites/cato/latest)
- [Cato CLI Documentation](https://github.com/catonetworks/cato-cli)

