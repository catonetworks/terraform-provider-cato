terraform {
  required_providers {
    cato = { source = "catonetworks/cato" }
  }
}

variable "token" {
  type = string
  sensitive = true
  description = "API token supplied through TF_VAR_token."
}

variable "account_id" {
  type = string
  description = "Cato account ID supplied through TF_VAR_account_id."
}

provider "cato" {
  token = var.token
  account_id = var.account_id
}

# Changes are automatically published. Use a test account when evaluating beta APIs.
resource "cato_socket_bypass_rule" "example" {
  # Optional placement; omission on update preserves the current position.
  # at = { position = "LAST_IN_POLICY" }
  rule = {
    action = {
      event = {
        enabled = false
      }
      port = "AUTOMATIC"
    }
    description = "Managed by Terraform"
    destination = {
      application = [{ id = "replace-with-object-id" }]
      custom_app = [{ id = "replace-with-object-id" }]
      domain = ["example.com"]
      fqdn = ["example.com"]
      global_ip_range = [{ id = "replace-with-object-id" }]
      ip = ["192.0.2.241"]
      ip_range = [{
        from = "192.0.2.1"
        to = "192.0.2.2"
      }]
      subnet = ["192.0.2.0/24"]
    }
    enabled = false
    exception = [{
      destination = {
        application = [{ id = "replace-with-object-id" }]
        custom_app = [{ id = "replace-with-object-id" }]
        domain = ["example.com"]
        fqdn = ["example.com"]
        global_ip_range = [{ id = "replace-with-object-id" }]
        ip = ["192.0.2.241"]
        ip_range = [{
          from = "192.0.2.1"
          to = "192.0.2.2"
        }]
        subnet = ["192.0.2.0/24"]
      }
      name = "terraform-socket-bypass-rule"
      service = {
        custom = [{
          port = ["443"]
          port_range = {
            from = "80"
            to = "81"
          }
          protocol = "ANY"
        }]
        simple = [{
          name = "terraform-socket-bypass-rule"
        }]
      }
      site = {
        group = [{ id = "replace-with-object-id" }]
        site = [{ id = "replace-with-object-id" }]
      }
      source = {
        floating_subnet = [{ id = "replace-with-object-id" }]
        global_ip_range = [{ id = "replace-with-object-id" }]
        group = [{ id = "replace-with-object-id" }]
        host = [{ id = "replace-with-object-id" }]
        ip = ["192.0.2.241"]
        ip_range = [{
          from = "192.0.2.1"
          to = "192.0.2.2"
        }]
        network_interface = [{ id = "replace-with-object-id" }]
        site_network_subnet = [{ id = "replace-with-object-id" }]
        subnet = ["192.0.2.0/24"]
        vlan = [1]
      }
    }]
    name = "terraform-socket-bypass-rule"
    service = {
      custom = [{
        port = ["443"]
        port_range = {
          from = "80"
          to = "81"
        }
        protocol = "ANY"
      }]
      simple = [{
        name = "terraform-socket-bypass-rule"
      }]
    }
    site = {
      group = [{ id = "replace-with-object-id" }]
      site = [{ id = "replace-with-object-id" }]
    }
    source = {
      floating_subnet = [{ id = "replace-with-object-id" }]
      global_ip_range = [{ id = "replace-with-object-id" }]
      group = [{ id = "replace-with-object-id" }]
      host = [{ id = "replace-with-object-id" }]
      ip = ["192.0.2.241"]
      ip_range = [{
        from = "192.0.2.1"
        to = "192.0.2.2"
      }]
      network_interface = [{ id = "replace-with-object-id" }]
      site_network_subnet = [{ id = "replace-with-object-id" }]
      subnet = ["192.0.2.0/24"]
      vlan = [1]
    }
  }
}
