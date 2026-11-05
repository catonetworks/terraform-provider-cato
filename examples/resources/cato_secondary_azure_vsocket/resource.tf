# Configure an existing Azure site with an unused secondary slot.
# Use distinct, nonreserved addresses within that site's native network range.
# Provision the Azure VM and networking separately.
resource "cato_secondary_azure_vsocket" "example" {
  site = {
    id = "12345"
    # name = "Azure branch" # Use name instead of id if preferred.
  }
  interface_ip = "10.20.1.10"
  floating_ip  = "10.20.1.20"
}
