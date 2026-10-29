# Configure an existing GCP site with an unused secondary slot.
# Provision GCP infrastructure separately and enable the secondary GCP API rollout.
# Use distinct nonreserved addresses within the site's native range.
resource "cato_secondary_gcp_vsocket" "example" {
  site = {
    id = "12345"
    # name = "GCP branch" # Use name instead of id if preferred.
  }
  gcp_config = {
    interface_ip     = "10.20.1.10"
    load_balancer_ip = "10.20.1.20"
  }
}
