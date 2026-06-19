# App tenant restriction rule (GraphQL @beta).
resource "cato_app_tenant_restriction_rule" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name        = "Example tenant restriction rule"
    description = "Managed by Terraform"
    enabled     = true
    action      = "INJECT_HEADERS"
    severity    = "LOW"
    application = {
      # id   = "office365"
      name = "Office365"
    }
    # headers = [{ name = "X-Example", value = "secret" }] # value is sensitive in schema
    schedule = {}
    source   = {}
  }
}
