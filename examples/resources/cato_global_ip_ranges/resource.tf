resource "cato_global_ip_ranges" "this" {
  ranges = [
    {
      description =  "example range description"
      ip_range =  "10.200.0.0/16"
      name =  "example range"
    }
  ]
}
