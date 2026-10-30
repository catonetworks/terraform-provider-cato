# Allocates a new public IP. Only this allocation is released on destruction.
resource "cato_allocated_ip" "example" {
  type         = "SYSTEM"
  pop_location = { name = "London" }
  # description = "Terraform managed public IP"
}

# type and pop_location changes require replacement.
# BYOIP allocation/release need private API inputs and are unsupported here.
# terraform import cato_allocated_ip.example ALLOCATION_ID
