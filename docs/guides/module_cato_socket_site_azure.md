---
subcategory: "Example Modules"
page_title: "Azure VNET Socket Module"
description: |-
  Provides an combined example of creating a virtual socket site in Cato Management Application, and templates for creating an Resource Group with underlying network resources and deploying a virtual socket instance in Azure.
---

# Example Azure Module (cato_socket_site)

The `cato_socket_site` resource contains the configuration parameters necessary to 
add a socket site to the Cato cloud 
([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)).
Documentation for the underlying API used in this resource can be found at
[mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite).

## Example Usage

### Create Azure VNET - Example Module

<details>
<summary>Azure VNET - Example Module</summary>

In your current project working folder, a `1-vnet` subfolder, and add a `main.tf` file with the following contents:

```hcl
## VNET Variables
variable "assetprefix" {
  type = string
  description = "Your asset prefix for resources created"
  default = null
}

variable "location" { 
  type = string
  default = null
}

variable "lan_ip" {
	type = string
	default = null
}

variable "project_name" { 
  type = string
  default = null
}

variable "dns_servers" { 
  type = list(string)
  default = [
    "10.254.254.1", # Cato Cloud DNS
    "168.63.129.16", # Azure DNS
    "1.1.1.1",
    "8.8.8.8"
  ]
}

variable "subnet_range_mgmt" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Management subnet. This subnet will be used initially to access the public internet and register your vSocket to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
  default = null
}

variable "subnet_range_wan" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Public/WAN subnet. This subnet will be used to access the public internet and securely tunnel to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
  default = null
}

variable "subnet_range_lan" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Private/LAN subnet. This subnet will host the target LAN interface of the vSocket so resources in the VPC (or AWS Region) can route to the Cato Cloud.
    The minimum subnet length to support High Availability is /29.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
  default = null
}

variable "vnet_prefix" {
  type = string
  description = <<EOT
  	Choose a unique range for your new VPC that does not conflict with the rest of your Wide Area Network.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
  default = null
}

## VNET Module Resources
provider "azurerm" {
	features {}
}

resource "azurerm_resource_group" "azure-rg" {
  location = var.location
  name = var.project_name
}

resource "azurerm_availability_set" "availability-set" {
  location                     = var.location
  name                         = "${var.assetprefix}-availabilitySet"
  platform_fault_domain_count  = 2
  platform_update_domain_count = 2
  resource_group_name          = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}

## Create Network and Subnets
resource "azurerm_virtual_network" "vnet" {
  address_space       = [var.vnet_prefix]
  location            = var.location
  name                = "${var.assetprefix}-vsNet"
  resource_group_name = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}

resource "azurerm_virtual_network_dns_servers" "dns_servers" {
  virtual_network_id = azurerm_virtual_network.vnet.id
  dns_servers        = var.dns_servers
}

resource "azurerm_subnet" "subnet-mgmt" {
  address_prefixes     = [var.subnet_range_mgmt]
  name                 = "subnetMGMT"
  resource_group_name  = azurerm_resource_group.azure-rg.name
  virtual_network_name = "${var.assetprefix}-vsNet"
  depends_on = [
    azurerm_virtual_network.vnet
  ]
}
resource "azurerm_subnet" "subnet-wan" {
  address_prefixes     = [var.subnet_range_wan]
  name                 = "subnetWAN"
  resource_group_name  = azurerm_resource_group.azure-rg.name
  virtual_network_name = "${var.assetprefix}-vsNet"
  depends_on = [
    azurerm_virtual_network.vnet
  ]
}

resource "azurerm_subnet" "subnet-lan" {
  address_prefixes     = [var.subnet_range_lan]
  name                 = "subnetLAN"
  resource_group_name  = azurerm_resource_group.azure-rg.name
  virtual_network_name = "${var.assetprefix}-vsNet"
  depends_on = [
    azurerm_virtual_network.vnet
  ]
}

# Allocate Public IPs
resource "azurerm_public_ip" "mgmt-public-ip" {
  allocation_method   = "Static"
  location            = var.location
  name                = "${var.assetprefix}-vs0nicMngPublicIP"
  resource_group_name = azurerm_resource_group.azure-rg.name
  sku                 = "Standard"
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}
resource "azurerm_public_ip" "wan-public-ip" {
  allocation_method   = "Static"
  location            = var.location
  name                = "${var.assetprefix}-vs0nicWanPublicIP"
  resource_group_name = azurerm_resource_group.azure-rg.name
  sku                 = "Standard"
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}

# Create Network Interfaces
resource "azurerm_network_interface" "mgmt-nic" {
  location            = var.location
  name                = "${var.assetprefix}-vs0nicMng"
  resource_group_name = azurerm_resource_group.azure-rg.name
  ip_configuration {
    name                          = "vs0nicMngIP"
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.mgmt-public-ip.id
    subnet_id                     = azurerm_subnet.subnet-mgmt.id
  }
  depends_on = [
    azurerm_public_ip.mgmt-public-ip,
    azurerm_subnet.subnet-mgmt
  ]
}

resource "azurerm_network_interface" "wan-nic" {
  enable_ip_forwarding = true
  location             = var.location
  name                 = "${var.assetprefix}-vs0nicWan"
  resource_group_name          = azurerm_resource_group.azure-rg.name
  ip_configuration {
    name                          = "vs0nicWanIP"
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.wan-public-ip.id
    subnet_id                     = azurerm_subnet.subnet-wan.id
  }
  depends_on = [
    azurerm_public_ip.wan-public-ip,
    azurerm_subnet.subnet-wan
  ]
}

resource "azurerm_network_interface" "lan-nic" {
  enable_ip_forwarding = true
  location             = var.location
  name                 = "${var.assetprefix}-vs0nicLan"
  resource_group_name          = azurerm_resource_group.azure-rg.name
  ip_configuration {
    name                          = "lanIPConfig"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurerm_subnet.subnet-lan.id
  }
  depends_on = [
    azurerm_subnet.subnet-lan
  ]
}

resource "azurerm_network_interface_security_group_association" "mgmt-nic-association" {
  network_interface_id      = azurerm_network_interface.mgmt-nic.id
  network_security_group_id = azurerm_network_security_group.mgmt-sg.id
  depends_on = [
    azurerm_network_interface.mgmt-nic,
    azurerm_network_security_group.mgmt-sg
  ]
}

resource "azurerm_network_interface_security_group_association" "wan-nic-association" {
  network_interface_id      = azurerm_network_interface.wan-nic.id
  network_security_group_id = azurerm_network_security_group.wan-sg.id
  depends_on = [
    azurerm_network_interface.wan-nic,
    azurerm_network_security_group.wan-sg
  ]
}

resource "azurerm_network_interface_security_group_association" "lan-nic-association" {
  network_interface_id      = azurerm_network_interface.lan-nic.id
  network_security_group_id = azurerm_network_security_group.lan-sg.id
  depends_on = [
    azurerm_network_interface.lan-nic,
    azurerm_network_security_group.lan-sg
  ]
}

# Create Security Groups
resource "azurerm_network_security_group" "mgmt-sg" {
  location            = var.location
  name                = "${var.assetprefix}-MGMTSecurityGroup"
  resource_group_name = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}
resource "azurerm_network_security_group" "wan-sg" {
  location            = var.location
  name                = "${var.assetprefix}-WANSecurityGroup"
  resource_group_name = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}
resource "azurerm_network_security_group" "lan-sg" {
  location            = var.location
  name                = "${var.assetprefix}-LANSecurityGroup"
  resource_group_name = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}

## Create Route Tables, Routes and Associations 
resource "azurerm_route_table" "private-rt" {
  disable_bgp_route_propagation = true
  location                      = var.location
  name                          = "${var.assetprefix}-viaCato"
  resource_group_name           = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}
resource "azurerm_route" "public-rt" {
  address_prefix      = "23.102.135.246/32"
  name                = "Microsoft-KMS"
  next_hop_type       = "Internet"
  resource_group_name = azurerm_resource_group.azure-rg.name
  route_table_name    = "${var.assetprefix}-viaCato"
  depends_on = [
    azurerm_route_table.private-rt
  ]
}
resource "azurerm_route" "lan-route" {
  address_prefix         = "0.0.0.0/0"
  name                   = "default"
  next_hop_in_ip_address = var.lan_ip
  next_hop_type          = "VirtualAppliance"
  resource_group_name    = azurerm_resource_group.azure-rg.name
  route_table_name       = "${var.assetprefix}-viaCato"
  depends_on = [
    azurerm_route_table.private-rt
  ]
}
resource "azurerm_route_table" "public-rt" {
  disable_bgp_route_propagation = true
  location                      = var.location
  name                          = "${var.assetprefix}-viaInternet"
  resource_group_name           = azurerm_resource_group.azure-rg.name
  depends_on = [
    azurerm_resource_group.azure-rg
  ]
}
resource "azurerm_route" "route-internet" {
  address_prefix      = "0.0.0.0/0"
  name                = "default"
  next_hop_type       = "Internet"
  resource_group_name = azurerm_resource_group.azure-rg.name
  route_table_name    = "${var.assetprefix}-viaInternet"
  depends_on = [
    azurerm_route_table.public-rt
  ]
}

resource "azurerm_subnet_route_table_association" "rt-table-association-mgmt" {
  route_table_id = azurerm_route_table.public-rt.id
  subnet_id      = azurerm_subnet.subnet-mgmt.id
  depends_on = [
    azurerm_route_table.public-rt,
    azurerm_subnet.subnet-mgmt
  ]
}

resource "azurerm_subnet_route_table_association" "rt-table-association-wan" {
  route_table_id = azurerm_route_table.public-rt.id
  subnet_id      = azurerm_subnet.subnet-wan.id
  depends_on = [
    azurerm_route_table.public-rt,
    azurerm_subnet.subnet-wan,
  ]
}

resource "azurerm_subnet_route_table_association" "rt-table-association-lan" {
  route_table_id = azurerm_route_table.private-rt.id
  subnet_id      = azurerm_subnet.subnet-lan.id
  depends_on = [
    azurerm_route_table.private-rt,
    azurerm_subnet.subnet-lan
  ]
}

## VNET Module Outputs:
output "resource-group-name" { value = azurerm_resource_group.azure-rg.name }
output "mgmt-nic-id" { value = azurerm_network_interface.mgmt-nic.id }
output "wan-nic-id" { value = azurerm_network_interface.wan-nic.id }
output "lan-nic-id" { value = azurerm_network_interface.lan-nic.id }
output "lan_subnet_id" { value = azurerm_subnet.subnet-lan.id }
```
</details>

In your current project working folder, add a `main.tf` file with the following contents:

```hcl
terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
  required_version = ">= 0.13"
}

## Create Azure Resource Group and Virtual Network
module "vnet" {
  source = "./1-vnet"
  location = var.location
  project_name = var.project_name
  assetprefix = var.assetprefix
  subnet_range_mgmt = var.subnet_range_mgmt
  subnet_range_wan = var.subnet_range_wan
  subnet_range_lan = var.subnet_range_lan
  lan_ip = var.lan_ip
  vnet_prefix = var.vnet_prefix
}

## Create Cato SocketSite and Deploy vSocket
module "vsocket-azure" {
  source              = "catonetworks/vsocket-azure/cato"
  token               = var.cato_token
  account_id          = var.account_id
  vnet_prefix         = var.vnet_prefix
  lan_ip              = var.lan_ip
  location            = var.location
  resource-group-name = module.vnet.resource-group-name
  mgmt-nic-id         = module.vnet.mgmt-nic-id
  wan-nic-id          = module.vnet.wan-nic-id
  lan-nic-id          = module.vnet.lan-nic-id
  site_name           = var.project_name
  site_description    = var.site_description
  site_type           = var.site_type
  site_location       = var.site_location
}
```