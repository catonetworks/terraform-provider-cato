# Criteria are ORed; members within each criterion are unordered.
resource "cato_custom_application" "example" {
  name = "Example custom application"
  # description = "Optional description"
  # category = [{ id = "category-id" }]
  # Use a category name instead of its ID if preferred.
  criteria = [{
    protocol = "TCP"
    port = ["443"]
    port_range = [] # Required; use [] if unused.
    # destination = {
    #   domain = ["example.com"]
    #   fqdn = ["api.example.com"]
    #   destination_ip = {
    #     ip = ["192.0.2.1"]
    #     subnet = ["198.51.100.0/24"]
    #     ip_range = [{ from = "203.0.113.1", to = "203.0.113.10" }]
    #   }
    # }
  }]
}
