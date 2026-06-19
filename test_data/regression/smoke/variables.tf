# ## Cato provider variables
variable "baseurl" { type = string }
variable "cato_token" { type = string }
variable "account_id" { type = string }

# Socket site variables
variable "socket_name" { default = "TF-Test-Socket-X1700-Site" }
variable "description" { default = "Test Socket X1700 Site created by Terraform" }
variable "site_type" { default = "BRANCH" }
variable "connection_type" { default = "SOCKET_X1700" }
variable "socket_native_network_range" { default = "248.248.100.0/24" }
variable "socket_local_ip" { default = "248.248.100.1" }
variable "socket_dhcp_type" { default = "DHCP_RANGE" }
variable "socket_ip_range" { default = "248.248.100.50-248.248.100.100" }
variable "socket_country_code" { default = "FR" }
variable "socket_timezone" { default = "Europe/Paris" }

# VLAN range variables
variable "interface_index" { default = "INT_3" }
variable "vlan_range_name" { default = "TF-Test-Socket-X1700-Site_VLAN_Range" }
variable "range_type" { default = "VLAN" }
variable "vlan_range_subnet" { default = "248.249.100.0/24" }
variable "vlan_range_local_ip" { default = "248.249.100.2" }
variable "vlan_range_vlan" { default = 202 }
variable "vlan_range_ip_range" { default = "248.249.100.20-248.249.100.32" }
variable "vlan_range_dhcp_microsegmentation" { default = true }
