# Configure an existing AWS vSocket site with an unused secondary slot.
# Provision the AWS ENI, instance and route table separately.
resource "cato_secondary_aws_vsocket" "example" {
  site = {
    id = "12345"
    # name = "AWS branch" # Use name instead of id if preferred.
  }
  eni_ip_address = "10.20.1.10"
  eni_ip_subnet  = "10.20.1.0/24"
  # This table is shared by the site's socket connection configuration.
  route_table_id = "rtb-example"
}
