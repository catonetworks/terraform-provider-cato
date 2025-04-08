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

### Create Socket Site and vSocket - Example Module

In your current project working folder, a `2-vSocket` subfolder, and add a `main.tf` file with the following contents:

<details>
<summary>vSocket - Example Module</summary>

### Create Azure vSocket - Example Module

```hcl
terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
  required_version = ">= 0.13"
}

## vSocket Module Varibables
variable cato_token {}

variable "account_id" {
  description = "Account ID"
  type        = number
  default	  = null
}

variable "site_description" {
  type = string
  description = "Site description"
}

variable "site_type" {
  description = "The type of the site"
  type        = string
  default	 = "DATACENTER"
  validation {
    condition = contains(["DATACENTER","BRANCH","CLOUD_DC","HEADQUARTERS"], var.site_type)
    error_message = "The site_type variable must be one of 'DATACENTER','BRANCH','CLOUD_DC','HEADQUARTERS'."
  }
}

variable "project_name" {
  type = string
  description = "Your Cato Deployment Name Here"
  default = null
}

variable "vnet_prefix" {
  type = string
  description = <<EOT
  	Choose a unique range for your new VPC that does not conflict with the rest of your Wide Area Network.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "lan_ip" {
	type = string
  description = "Local IP Address of socket LAN interface"
	default = null
}

## vSocket Params
variable "assetprefix" {
  type = string
  description = "Your asset prefix for resources created"
  default = null
}

variable "location" { 
  type = string
  description = "(Required) The Azure Region where the Resource Group should exist. Changing this forces a new Resource Group to be created."
  default = null
}

variable "resource-group-name" { 
  type = string
  description = "(Required) The Name which should be used for this Resource Group. Changing this forces a new Resource Group to be created."
  default = null
}

variable "mgmt-nic-id" {
  type = string
  default = null
}

variable "wan-nic-id" {
  type = string
  default = null
}

variable "lan-nic-id" {
  type = string
  default = null
}

variable "image_reference_id" {
	type = string
  description = "Path to image used to deploy specific version of the virutal socket"
	default = "/Subscriptions/38b5ec1d-b3b6-4f50-a34e-f04a67121955/Providers/Microsoft.Compute/Locations/eastus/Publishers/catonetworks/ArtifactTypes/VMImage/Offers/cato_socket/Skus/public-cato-socket/Versions/19.0.17805"
}

## vSocket Module Variables
provider "azurerm" {
	features {}
}

provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

resource "cato_socket_site" "azure-site" {
    connection_type  = "SOCKET_AZ1500"
    description = var.site_description
    name = var.project_name
    native_range = {
      native_network_range = var.vnet_prefix
      local_ip = var.lan_ip
    }
    site_location = {
        country_code = "US"
		    state_code = "US-VA"
        timezone = "America/New_York"
    }
    site_type = var.site_type
}

data "cato_accountSnapshotSite" "azure-site" {
	id = cato_socket_site.azure-site.id
}

## Create vSocket Virtual Machine
resource "azurerm_virtual_machine" "vSocket" {
  location                     = var.location
  name                         = "${var.assetprefix}-vSocket"
  network_interface_ids        = [var.mgmt-nic-id, var.wan-nic-id, var.lan-nic-id]
  primary_network_interface_id = var.mgmt-nic-id
  resource_group_name          = var.resource-group-name
  vm_size                      = "Standard_D8ls_v5"
  plan {
    name      = "public-cato-socket"
    product   = "cato_socket"
    publisher = "catonetworks"
  }
  boot_diagnostics {
    enabled     = true
    storage_uri = ""
  }
  storage_os_disk {
    create_option     = "Attach"
    name              = "${var.assetprefix}-vSocket-disk1"
    managed_disk_id   = azurerm_managed_disk.vSocket-disk1.id
    os_type = "Linux"
  }
  
  depends_on = [
    azurerm_managed_disk.vSocket-disk1
  ]
}

resource "azurerm_managed_disk" "vSocket-disk1" {
  name                 = "${var.assetprefix}-vSocket-disk1"
  location             = var.location
  resource_group_name  = var.resource-group-name
  storage_account_type = "Standard_LRS"
  create_option        = "FromImage"
  disk_size_gb         = 8
  os_type              = "Linux"
  image_reference_id   = var.image_reference_id
}

variable "commands" {
  type    = list(string)
  default = [
    "rm /cato/deviceid.txt",
    "rm /cato/socket/configuration/socket_registration.json",
    "nohup /cato/socket/run_socket_daemon.sh &"
   ]
}

resource "azurerm_virtual_machine_extension" "vSocket-custom-script" {
  auto_upgrade_minor_version = true
  name                       = "vSocket-custom-script"
  publisher                  = "Microsoft.Azure.Extensions"
  type                       = "CustomScript"
  type_handler_version       = "2.1"
  virtual_machine_id         = azurerm_virtual_machine.vSocket.id
  settings = <<SETTINGS
 {
  "commandToExecute": "${"echo '${data.cato_accountSnapshotSite.azure-site.info.sockets[0].serial}' > /cato/serial.txt"};${join(";", var.commands)}"
 }
SETTINGS
  depends_on = [
    azurerm_virtual_machine.vSocket
  ]
}
```
</details>

### Create Windows VM - Example Module (optional)

In your current project working folder, a `3-WindowsVM` subfolder, and add a `main.tf` file with the following contents:

<details>
<summary>Windows VM - Example Module</summary>

### Create Windows VM - Example Module

```hcl
## Windows VM Module Variables 
variable "location" { 
  type = string
  description = "(Required) The Azure Region where the Resource Group should exist. Changing this forces a new Resource Group to be created."
  default = null
}

variable "assetprefix" {
  type = string
  description = "Your asset prefix for resources created"
  default = null
}

variable "resource-group-name" { 
  type = string
  description = "(Required) The Name which should be used for this Resource Group. Changing this forces a new Resource Group to be created."
  default = null
}

variable "project_name" {
  type = string
  description = "Your Cato Deployment Name Here"
  default = null
}

variable "lan_subnet_id" { 
  type = string
  description = "(Required) LAN Subnet ID"
  default = null
}

variable "admin_username" {
  type = string
  description = "Admin Username for the VM"
  default = null
}

variable "admin_password" {
  type = string
  description = "Admin Password for the VM"
  default = null
}

provider "azurerm" {
	features {}
}

# Create Network Interfaces
resource "azurerm_network_interface" "lan-nic" {
  location            = var.location
  name                = "${var.windows-assets-prefix}-LAN"
  resource_group_name = var.resource-group-name
  ip_configuration {
    name                          = "LAN-IP"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = var.lan_subnet_id
  }
}

# Create a Windows Virtual Machine
resource "azurerm_virtual_machine" "vm" {
  name                  = var.windows-assets-prefix
  location              = var.location
  resource_group_name   = var.resource-group-name
  network_interface_ids = [azurerm_network_interface.lan-nic.id]
  vm_size               = "Standard_D4S_v3"
  
  storage_os_disk {
    name            = azurerm_managed_disk.os_disk.name
    managed_disk_id = azurerm_managed_disk.os_disk.id
    create_option   = "fromImage"
    os_type         = "Windows"
  }

  os_profile {
    computer_name  = var.windows-assets-prefix
    admin_username = var.admin_username
    admin_password = var.admin_password
  }

  os_profile_windows_config {
    provision_vm_agent        = true
    enable_automatic_upgrades = true
  }

  tags = {
    environment = var.windows-assets-prefix
  }
}

data "azurerm_platform_image" "windows_image" {
  location  = var.location
  publisher = "MicrosoftWindowsServer"
  offer     = "WindowsServer"
  sku       = "2019-Datacenter"
  version   = "17763.6189.240811"
}

resource "azurerm_managed_disk" "os_disk" {
  name                 = "win-osdisk"
  location             = var.location
  resource_group_name  = var.resource-group-name
  storage_account_type = "Standard_LRS"
  create_option        = "FromImage"
  image_reference_id   = data.azurerm_platform_image.windows_image.id
  disk_size_gb         = 127

  tags = {
    environment = var.windows-assets-prefix
  }
}

resource "azurerm_network_security_group" "windows" {
  name                = replace("${var.windows-assets-prefix}SecurityGroup", "-", "")
  location            = var.location
  resource_group_name = var.resource-group-name

  security_rule {
    name                       = "Allow-RDP"
    priority                   = 300
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = {
    environment = var.windows-assets-prefix
  }
}

resource "azurerm_subnet_network_security_group_association" "windows" {
  subnet_id                 = var.lan_subnet_id
  network_security_group_id = azurerm_network_security_group.windows.id
}
```
</details>

### VPC, vSocket, and WindowsVM Module Usage

<details>
<summary>Project Variables</summary>

### Project Variables Example

In your current project working folder, add a `variables.tf` file with the following contents:

```hcl
variable cato_token {}

variable "account_id" {
  description = "Account ID"
  type        = number
  default	  = null
}

variable "location" { 
  type = string
  default = null
}

variable "project_name" {
  type = string
  description = "Your Cato Deployment Name Here"
  default = null
}

variable "assetprefix" {
  type = string
  description = "Your asset prefix for resources created"
  default = null
}

variable "site_description" {
  type = string
  description = "Site description"
}

variable "site_type" {
  description = "The type of the site"
  type        = string
  default	 = "DATACENTER"
  validation {
    condition = contains(["DATACENTER","BRANCH","CLOUD_DC","HEADQUARTERS"], var.site_type)
    error_message = "The site_type variable must be one of 'DATACENTER','BRANCH','CLOUD_DC','HEADQUARTERS'."
  }
}

variable "vnet_prefix" {
  type = string
  description = <<EOT
  	Choose a unique range for your new VPC that does not conflict with the rest of your Wide Area Network.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "subnet_range_mgmt" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Management subnet. This subnet will be used initially to access the public internet and register your vSocket to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "subnet_range_wan" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Public/WAN subnet. This subnet will be used to access the public internet and securely tunnel to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "subnet_range_lan" {
  type = string
  description = <<EOT
    Choose a range within the VPC to use as the Private/LAN subnet. This subnet will host the target LAN interface of the vSocket so resources in the VPC (or AWS Region) can route to the Cato Cloud.
    The minimum subnet length to support High Availability is /29.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "lan_ip" {
	type = string
  description = "Local IP Address of socket LAN interface"
	default = null
}
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
module "vSocket" {
  source = "./2-vSocket"
  cato_token = var.cato_token
  account_id = var.account_id
  site_description = var.site_description
  site_type = var.site_type
  project_name = var.project_name
  vnet_prefix = var.vnet_prefix
  lan_ip = var.lan_ip
  location = var.location
  assetprefix = var.assetprefix
  resource-group-name = module.vnet.resource-group-name
  mgmt-nic-id = module.vnet.mgmt-nic-id
  wan-nic-id = module.vnet.wan-nic-id
  lan-nic-id = module.vnet.lan-nic-id
}

## Create Windows VM on LAN
module "WindowsVM" {
  source = "./3-WindowsVM"
  location = var.location
  assetprefix = var.assetprefix
  resource-group-name = module.vnet.resource-group-name
  project_name = var.project_name
  lan_subnet_id = module.vnet.lan_subnet_id
  admin_username = "your-username-here"
  admin_password = "your-password-here"
}
```
