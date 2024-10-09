---
subcategory: "Example Modules"
page_title: "AWS VPC Socket Module"
description: |-
  Provides an combined example of creating a virtual socket site in Cato Management Application, and templates for creating a VPC and deploying a virtual socket instance in AWS.
---

# Example AWS Module (cato-oss_socket_site)

The `cato-oss_socket_site` resource contains the configuration parameters necessary to 
add a socket site to the Cato cloud 
([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)).
Documentation for the underlying API used in this resource can be found at
[mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite).

## Example Usage

### Create AWS VPC - Example Module

<details>
<summary>AWS VPC - Example Module</summary>

In your current project working folder, a `1-vpc` subfolder, and add a `main.tf` file with the following contents:

```hcl
## VPC Variables
variable "region" {
  type    = string
  default = "us-east-2"
}

variable "project_name" {
  type    = string
  default = "Cato Lab"
}

variable "ingress_cidr_blocks" {
  type    = list(any)
  default = null
}

variable "subnet_range_mgmt" {
  type    = string
  default = null
}

variable "subnet_range_wan" {
  type    = string
  default = null
}

variable "subnet_range_lan" {
  type    = string
  default = null
}

variable "vpc_range" {
  type    = string
  default = null
}

variable "mgmt_eni_ip" {
  description = "Choose an IP Address within the Management Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "wan_eni_ip" {
  description = "Choose an IP Address within the Public/WAN Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "lan_eni_ip" {
  description = "Choose an IP Address within the LAN Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "vpc_id" {
  description = ""
  type        = string
  default     = null
}

## VPC Module Resources
provider "aws" {
  region = var.region
}

resource "aws_vpc" "cato-lab" {
  cidr_block = var.vpc_range
  tags = {
    Name = "${var.project_name}-VPC"
  }
}

# Lookup data from region and VPC
data "aws_availability_zones" "available" {
  state = "available"
}

# Internet Gateway and Attachment
resource "aws_internet_gateway" "internet_gateway" {}

resource "aws_internet_gateway_attachment" "attach_gateway" {
  internet_gateway_id = aws_internet_gateway.internet_gateway.id
  vpc_id              = aws_vpc.cato-lab.id
}

# Subnets
resource "aws_subnet" "mgmt_subnet" {
  vpc_id            = aws_vpc.cato-lab.id
  cidr_block        = var.subnet_range_mgmt
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "${var.project_name}-MGMT-Subnet"
  }
}

resource "aws_subnet" "wan_subnet" {
  vpc_id            = aws_vpc.cato-lab.id
  cidr_block        = var.subnet_range_wan
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "${var.project_name}-WAN-Subnet"
  }
}

resource "aws_subnet" "lan_subnet" {
  vpc_id            = aws_vpc.cato-lab.id
  cidr_block        = var.subnet_range_lan
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "${var.project_name}-LAN-Subnet"
  }
}

# Internal and External Security Groups
resource "aws_security_group" "internal_sg" {
  name        = "${var.project_name}-Internal-SG"
  description = "CATO LAN Security Group - Allow all traffic Inbound"
  vpc_id      = aws_vpc.cato-lab.id
  ingress = [
    {
      description      = "Allow all traffic Inbound from Ingress CIDR Blocks"
      protocol         = -1
      from_port        = 0
      to_port          = 0
      cidr_blocks      = var.ingress_cidr_blocks
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      security_groups  = []
      self             = false
    }
  ]
  egress = [
    {
      description      = "Allow all traffic Outbound"
      protocol         = -1
      from_port        = 0
      to_port          = 0
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      security_groups  = []
      self             = false
    }
  ]
  tags = {
    name = "${var.project_name}-Internal-SG"
  }
}

resource "aws_security_group" "external_sg" {
  name        = "${var.project_name}-External-SG"
  description = "CATO WAN Security Group - Allow HTTPS In"
  vpc_id      = aws_vpc.cato-lab.id
  ingress = [
    {
      description      = "Allow HTTPS In"
      protocol         = "tcp"
      from_port        = 443
      to_port          = 443
      cidr_blocks      = var.ingress_cidr_blocks
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      security_groups  = []
      self             = false
    },
    {
      description      = "Allow SSH In"
      protocol         = "tcp"
      from_port        = 22
      to_port          = 22
      cidr_blocks      = var.ingress_cidr_blocks
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      security_groups  = []
      self             = false
    }
  ]
  egress = [
    {
      description      = "Allow all traffic Outbound"
      protocol         = -1
      from_port        = 0
      to_port          = 0
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      security_groups  = []
      self             = false
    }
  ]
  tags = {
    name = "${var.project_name}-External-SG"
  }
}

# vSocket Network Interfaces
resource "aws_network_interface" "mgmteni" {
  source_dest_check = "true"
  subnet_id         = aws_subnet.mgmt_subnet.id
  private_ips       = [var.mgmt_eni_ip]
  security_groups   = [aws_security_group.external_sg.id]
  tags = {
    Name = "${var.project_name}-MGMT-INT"
  }
}

resource "aws_network_interface" "waneni" {
  source_dest_check = "true"
  subnet_id         = aws_subnet.wan_subnet.id
  private_ips       = [var.wan_eni_ip]
  security_groups   = [aws_security_group.external_sg.id]
  tags = {
    Name = "${var.project_name}-WAN-INT"
  }
}

resource "aws_network_interface" "laneni" {
  source_dest_check = "false"
  subnet_id         = aws_subnet.lan_subnet.id
  private_ips       = [var.lan_eni_ip]
  security_groups   = [aws_security_group.internal_sg.id]
  tags = {
    Name = "${var.project_name}-LAN-INT"
  }
}

# Elastic IP Addresses
resource "aws_eip" "wanip" {
  tags = {
    Name = "${var.project_name}-WAN-EIP"
  }
}

resource "aws_eip" "mgmteip" {
  tags = {
    Name = "${var.project_name}-MGMT-EIP"
  }
}

# Elastic IP Addresses Association - Required to properly destroy 
resource "aws_eip_association" "wanip_assoc" {
  network_interface_id = aws_network_interface.waneni.id
  allocation_id        = aws_eip.wanip.id
}

resource "aws_eip_association" "mgmteip_assoc" {
  network_interface_id = aws_network_interface.mgmteni.id
  allocation_id        = aws_eip.mgmteip.id
}

# Routing Tables
resource "aws_route_table" "wanrt" {
  vpc_id = aws_vpc.cato-lab.id
  tags = {
    Name = "${var.project_name}-WAN-RT"
  }
}

resource "aws_route_table" "mgmtrt" {
  vpc_id = aws_vpc.cato-lab.id
  tags = {
    Name = "${var.project_name}-MGMT-RT"
  }
}

resource "aws_route_table" "lanrt" {
  vpc_id = aws_vpc.cato-lab.id
  tags = {
    Name = "${var.project_name}-LAN-RT"
  }
}

# Routes
resource "aws_route" "wan_route" {
  route_table_id         = aws_route_table.wanrt.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
}

resource "aws_route" "mgmt_route" {
  route_table_id         = aws_route_table.mgmtrt.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
}

resource "aws_route" "lan_route" {
  route_table_id         = aws_route_table.lanrt.id
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id   = aws_network_interface.laneni.id
}

# Route Table Associations
resource "aws_route_table_association" "mgmt_subnet_route_table_association" {
  subnet_id      = aws_subnet.mgmt_subnet.id
  route_table_id = aws_route_table.mgmtrt.id
}

resource "aws_route_table_association" "wan_subnet_route_table_association" {
  subnet_id      = aws_subnet.wan_subnet.id
  route_table_id = aws_route_table.wanrt.id
}

resource "aws_route_table_association" "lan_subnet_route_table_association" {
  subnet_id      = aws_subnet.lan_subnet.id
  route_table_id = aws_route_table.lanrt.id
}

## The following attributes are exported:
output "internet_gateway_id" { value = aws_internet_gateway.internet_gateway.id }
output "project_name" { value = var.project_name }
output "sg_internal" { value = aws_security_group.internal_sg.id }
output "sg_external" { value = aws_security_group.external_sg.id }
output "mgmt_eni_id" { value = aws_network_interface.mgmteni.id }
output "wan_eni_id" { value = aws_network_interface.waneni.id }
output "lan_eni_id" { value = aws_network_interface.laneni.id }
output "lan_subnet_id" { value = aws_subnet.lan_subnet.id }
output "vpc_id" { value = aws_vpc.cato-lab.id }

```
</details>

### Create Socket Site and vSocket - Example Module

In your current project working folder, a `2-vSocket` subfolder, and add a `main.tf` file with the following contents:

<details>
<summary>vSocket - Example Module</summary>

### Create AWS vSocket - Example Module

```hcl
terraform {
  required_providers {
    cato = {
      source = "cato-networks/cato"
    }
  }
  required_version = ">= 0.13"
}

## vSocket Module Varibables
variable "cato_token" {}

variable "account_id" {
  description = "Account ID"
  type        = number
  default     = null
}

## vSocket Module Varibables
variable "site_description" {
  type        = string
  description = "Site description"
}

variable "project_name" {
  type        = string
  description = "Your Cato Deployment Name Here"
  default     = "Your Cato Deployment Name Here"
}

variable "vpc_range" {
  type        = string
  description = <<EOT
  	Choose a unique range for your new VPC that does not conflict with the rest of your Wide Area Network.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "site_type" {
  description = "The type of the site"
  type        = string
  default     = "DATACENTER"
  validation {
    condition     = contains(["DATACENTER", "BRANCH", "CLOUD_DC", "HEADQUARTERS"], var.site_type)
    error_message = "The site_type variable must be one of 'DATACENTER','BRANCH','CLOUD_DC','HEADQUARTERS'."
  }
}

variable "site_location" {
  type = object({
    city         = string
    country_code = string
    state_code   = string
    timezone     = string
  })
  default = {
    city         = "New York"
    country_code = "US"
    state_code   = "US-NY" ## Optinal, for coutries with states only
    timezone     = "America/New_York"
  }
}


variable "lan_eni_ip" {
  description = "Choose an IP Address within the LAN Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

## Virtual Socket Variables
variable "vpc_id" {
  description = ""
  type        = string
  default     = null
}

variable "ebs_disk_size" {
  description = "Size of disk"
  type        = number
  default     = 32
}

variable "ebs_disk_type" {
  description = "Size of disk"
  type        = string
  default     = "gp2"
}

variable "instance_type" {
  description = "The instance type of the vSocket"
  type        = string
  default     = "c5.xlarge"
  validation {
    condition     = contains(["d2.xlarge", "c3.xlarge", "t3.large", "t3.xlarge", "c4.xlarge", "c5.xlarge", "c5d.xlarge", "c5n.xlarge"], var.instance_type)
    error_message = "The instance_type variable must be one of 'd2.xlarge','c3.xlarge','t3.large','t3.xlarge','c4.xlarge','c5.xlarge','c5d.xlarge','c5n.xlarge'."
  }
}

variable "key_pair" {
  description = "Name of an existing Key Pair for AWS encryption"
  type        = string
  default     = null
}

variable "region" {
  type    = string
  default = "us-east-2"
}

variable "mgmt_eni_id" {
  description = ""
  type        = string
  default     = null
}

variable "wan_eni_id" {
  description = ""
  type        = string
  default     = null
}

variable "lan_eni_id" {
  description = ""
  type        = string
  default     = null
}

provider "aws" {
  region = var.region
}

provider "cato-oss" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

resource "cato-oss_socket_site" "aws-site" {
  connection_type = "SOCKET_AWS1500"
  description     = var.site_description
  name            = var.project_name
  native_range = {
    native_network_range = var.vpc_range
    local_ip             = var.lan_eni_ip
  }
  site_location = var.site_location
  site_type     = var.site_type
}

data "cato-oss_accountSnapshotSite" "aws-site" {
  id = cato-oss_socket_site.aws-site.id
}

## Lookup data from region and VPC
data "aws_ami" "vSocket" {
  most_recent = true
  name_regex  = "VSOCKET_AWS"
  owners      = ["679593333241"]
}

data "aws_availability_zones" "available" {
  state = "available"
}

## vSocket Instance
resource "aws_instance" "vSocket" {
  tenancy           = "default"
  ami               = data.aws_ami.vSocket.id
  availability_zone = data.aws_availability_zones.available.names[0]
  key_name          = var.key_pair
  instance_type     = var.instance_type
  user_data         = base64encode(data.cato-oss_accountSnapshotSite.aws-site.info.sockets[0].serial)
  # Network Interfaces
  # MGMTENI
  network_interface {
    device_index         = 1
    network_interface_id = var.mgmt_eni_id
  }
  # WANENI
  network_interface {
    device_index         = 0
    network_interface_id = var.wan_eni_id
  }
  # LANENI
  network_interface {
    device_index         = 2
    network_interface_id = var.lan_eni_id
  }
  ebs_block_device {
    device_name = "/dev/sda1"
    volume_size = var.ebs_disk_size
    volume_type = var.ebs_disk_type
  }
  tags = {
    Name = "${var.project_name}-vSocket"
  }
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
variable "region" { 
  type = string
  default = "us-east-2" 
}

variable "vpc_id" {
  description = ""
  type        = string
  default	  = null
}

variable "lan_subnet_id" { 
  type = string
  description = "(Required) LAN Subnet ID"
  default = null
}

variable "key_pair" {
  description = "Name of an existing Key Pair for AWS encryption"
  type        = string
  default	  = null
}

provider "aws" {
	region = var.region
}

data "aws_ami" "windows" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2019-English-Full-Base-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "windows_instance" {
  ami                         = data.aws_ami.windows.id
  instance_type               = "t3.micro"
  subnet_id                   = var.lan_subnet_id
  key_name 				            = var.key_pair
  vpc_security_group_ids      = [aws_security_group.windows_sg.id]
  associate_public_ip_address = false

  tags = {
    Name = "windows-vm-instance"
  }
}

resource "aws_security_group" "windows_sg" {
  name        = "windows-vm-sg"
  description = "Allow RDP access"
  vpc_id      = var.vpc_id

  ingress {
    description = "RDP from VPC"
    from_port   = 3389
    to_port     = 3389
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "windows-vm-security-group"
  }
}

```

</details>

### VPC, vSocket, and WindowsVM Module Usage

<details>
<summary>Project Variables</summary>

### Project Variables Example

In your current project working folder, add a `variables.tf` file with the following contents:

```hcl
## VPC Variables
variable "cato_token" {}

variable "account_id" {
  description = "Account ID"
  type        = number
  default     = null
}

## Cato socket site variables
variable "site_description" {
  type        = string
  description = "Site description"
}

variable "project_name" {
  type        = string
  description = "Your Cato Deployment Name Here"
  default     = "Your Cato Deployment Name Here"
}

variable "vpc_range" {
  type        = string
  description = <<EOT
  	Choose a unique range for your new VPC that does not conflict with the rest of your Wide Area Network.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "site_type" {
  description = "The type of the site"
  type        = string
  default     = "DATACENTER"
  validation {
    condition     = contains(["DATACENTER", "BRANCH", "CLOUD_DC", "HEADQUARTERS"], var.site_type)
    error_message = "The site_type variable must be one of 'DATACENTER','BRANCH','CLOUD_DC','HEADQUARTERS'."
  }
}

variable "site_location" {
  type = object({
    city         = string
    country_code = string
    state_code   = string
    timezone     = string
  })
  default = {
    city         = null
    country_code = null
    state_code   = null
    timezone     = null
  }
}

## VPC Module Variables
variable "region" {
  type    = string
  default = "us-east-2"
}

variable "ingress_cidr_blocks" {
  type        = list(any)
  description = <<EOT
  	Set CIDR to receive traffic from the specified IPv4 CIDR address ranges
	For example x.x.x.x/32 to allow one specific IP address access, 0.0.0.0/0 to allow all IP addresses access, or another CIDR range
    Best practice is to allow a few IPs as possible
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT  
  default     = null
}

variable "instance_type" {
  description = "The instance type of the vSocket"
  type        = string
  default     = "c5.xlarge"
  validation {
    condition     = contains(["d2.xlarge", "c3.xlarge", "t3.large", "t3.xlarge", "c4.xlarge", "c5.xlarge", "c5d.xlarge", "c5n.xlarge"], var.instance_type)
    error_message = "The instance_type variable must be one of 'd2.xlarge','c3.xlarge','t3.large','t3.xlarge','c4.xlarge','c5.xlarge','c5d.xlarge','c5n.xlarge'."
  }
}

variable "key_pair" {
  description = "Name of an existing Key Pair for AWS encryption"
  type        = string
  default     = null
}

variable "mgmt_eni_ip" {
  description = "Choose an IP Address within the Management Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "wan_eni_ip" {
  description = "Choose an IP Address within the Public/WAN Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "lan_eni_ip" {
  description = "Choose an IP Address within the LAN Subnet. You CANNOT use the first four assignable IP addresses within the subnet as it's reserved for the AWS virtual router interface. The accepted input format is X.X.X.X"
  type        = string
  default     = null
}

variable "subnet_range_mgmt" {
  type        = string
  description = <<EOT
    Choose a range within the VPC to use as the Management subnet. This subnet will be used initially to access the public internet and register your vSocket to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "subnet_range_wan" {
  type        = string
  description = <<EOT
    Choose a range within the VPC to use as the Public/WAN subnet. This subnet will be used to access the public internet and securely tunnel to the Cato Cloud.
    The minimum subnet length to support High Availability is /28.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

variable "subnet_range_lan" {
  type        = string
  description = <<EOT
    Choose a range within the VPC to use as the Private/LAN subnet. This subnet will host the target LAN interface of the vSocket so resources in the VPC (or AWS Region) can route to the Cato Cloud.
    The minimum subnet length to support High Availability is /29.
    The accepted input format is Standard CIDR Notation, e.g. X.X.X.X/X
	EOT
}

```

</details>

In your current project working folder, add a `main.tf` file with the following contents:

```hcl
terraform {
  required_providers {
    cato = {
      source = "cato-networks/cato-oss"
    }
  }
  required_version = ">= 0.13"
}

module "vpc" {
    source = "./1-vpc"
    region = var.region 
    project_name = var.project_name
    vpc_range = var.vpc_range
    subnet_range_mgmt = var.subnet_range_mgmt
    subnet_range_wan = var.subnet_range_wan
    subnet_range_lan = var.subnet_range_lan
    mgmt_eni_ip = var.mgmt_eni_ip
    wan_eni_ip = var.wan_eni_ip
    lan_eni_ip = var.lan_eni_ip
    ingress_cidr_blocks = var.ingress_cidr_blocks
}

module "vSocket" {
    source = "./2-vSocket"
	cato_token = var.cato_token
	account_id = var.account_id
    instance_type = var.instance_type
    vpc_range = var.vpc_range
    site_description = var.site_description
    site_location = var.site_location
    key_pair = var.key_pair
    mgmt_eni_id = module.vpc.mgmt_eni_id
    wan_eni_id = module.vpc.wan_eni_id
    lan_eni_id = module.vpc.lan_eni_id
    lan_eni_ip = var.lan_eni_ip
    project_name = var.project_name
    region = var.region
    vpc_id = module.vpc.vpc_id
}

module "WindowsVM" {
    source = "./3-WindowsVM"
    lan_subnet_id = module.vpc.lan_subnet_id
    vpc_id = module.vpc.vpc_id  
    region = var.region
    key_pair = var.key_pair
}
```
