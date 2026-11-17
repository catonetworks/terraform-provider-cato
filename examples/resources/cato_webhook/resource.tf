# cato_webhook: outbound notification webhooks. Each webhook needs exactly ONE
# auth variant (basic | bearer | custom). Below are all three, plus a NO_AUTH-style
# placeholder, showing every optional field (header, response_mapping,
# request_method, related_webhook_ref, enabled).

# --- Basic auth ---
resource "cato_webhook" "basic_auth" {
  name           = "ServiceNow CMDB (basic)"
  url            = "https://example.service-now.com/api/webhook"
  enabled        = true
  request_method = "POST"
  body           = jsonencode({ event = "{{event}}", site = "{{site_name}}" })
  auth = {
    basic = {
      username = "svc_cato"
      password = "REPLACE_ME" # pull from a sensitive variable in real configs
    }
  }
  # Static headers sent with every call.
  header = [
    { name = "X-Source", value = "cato" },
  ]
  # Map fields out of the webhook response (stored as JSONPaths, e.g. "$.result.sys_id").
  response_mapping = [
    { key = "external_id", value = "$.result.sys_id" },
  ]
}

# --- Bearer token auth ---
resource "cato_webhook" "bearer_auth" {
  name           = "Generic SIEM (bearer)"
  url            = "https://siem.example.com/ingest"
  request_method = "POST"
  body           = jsonencode({ message = "{{event}}" })
  auth = {
    bearer = {
      bearer_token = "REPLACE_ME"
    }
  }
}

# --- Custom header auth (also used as the interim NO_AUTH placeholder) ---
resource "cato_webhook" "custom_auth" {
  name    = "Internal Hook (custom header)"
  url     = "https://hooks.internal.example.com/cato"
  enabled = false # defined but not yet active
  body    = jsonencode({ event = "{{event}}" })
  auth = {
    custom = {
      name  = "X-Webhook-Token"
      value = "REPLACE_ME"
    }
  }
  # Chain/relate this webhook to another by id or name.
  related_webhook_ref = {
    by    = "ID"
    input = cato_webhook.basic_auth.id
  }
}
