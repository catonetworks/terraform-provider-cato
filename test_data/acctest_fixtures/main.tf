terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = "1.4.0"
    }
  }
}

variable "account_id" {
  type = string
}

# CATO_TOKEN and CATO_BASEURL come from the environment. Keep state outside the
# checkout; the tf-fixture- prefix is deliberately excluded from acctest cleanup.
provider "cato" {
  account_id = var.account_id
}

resource "cato_socket_site" "fixture" {
  count           = 3
  name            = "tf-fixture-site-${count.index}"
  description     = "Persistent acceptance test fixture"
  site_type       = "BRANCH"
  connection_type = "SOCKET_X1500"
  native_range = {
    native_network_range = "10.240.${count.index}.0/24"
    local_ip             = "10.240.${count.index}.1"
  }
  site_location = {
    country_code = "CZ"
    timezone     = "Europe/Prague"
  }
}

resource "cato_static_host" "fixture" {
  count   = 3
  site_id = cato_socket_site.fixture[count.index].id
  name    = "tf-fixture-host-${count.index}"
  ip      = "10.240.${count.index}.10"
}

resource "cato_dhcp_relay_group" "fixture" {
  count  = 3
  name   = "tf-fixture-dhcp-relay-${count.index}"
  server = ["192.0.2.${51 + count.index}"]
}

resource "cato_mailing_list" "fixture" {
  count   = 3
  name    = "tf-fixture-mailing-${count.index}"
  address = ["tf-fixture-${count.index}@example.invalid"]
}

resource "cato_subscription_group" "fixture" {
  count        = 3
  name         = "tf-fixture-subscription-${count.index}"
  mailing_list = [{ id = cato_mailing_list.fixture[count.index].id }]
}

resource "cato_custom_application" "fixture" {
  count = 3
  name  = "tf-fixture-app-${count.index}"
  criteria = [{
    protocol   = "TCP"
    port       = [tostring(50440 + count.index)]
    port_range = []
  }]
}

output "test_vars" {
  value = {
    custom_apps = [for app in cato_custom_application.fixture : { id = app.id, name = app.name }]
  }
}
