---
subcategory: "Example Modules"
page_title: "AWS VPC Socket Module"
description: |-
  Provides an combined example of creating a virtual socket site in Cato Management Application, and templates for creating a VPC and deploying a virtual socket instance in AWS.
---

# Example AWS Module (cato_socket_site)

The `cato_socket_site` resource contains the configuration parameters necessary to 
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

module "vsocket-aws" {
    source               = "catonetworks/vsocket-aws/cato"
    token                = var.cato_token
    account_id           = var.account_id
    vpc_id               = module.vpc.vpc_id
    key_pair             = var.key_pair
    native_network_range = var.vpc_range
    region               = var.region
    site_name            = "AWS Site ${var.region}"
    site_description     = "AWS Site ${var.region}"
    site_type            = "CLOUD_DC"
    mgmt_eni_id          = module.vpc.mgmt_eni_id
    wan_eni_id           = module.vpc.wan_eni_id
    lan_eni_id           = module.vpc.lan_eni_id
    lan_local_ip         = var.lan_eni_ip
    site_location        = var.site_location
}
```
