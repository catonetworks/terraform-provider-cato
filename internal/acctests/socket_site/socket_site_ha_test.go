//go:build acctest

package socket_site

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSocketSite_InlineVSocketHA(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	if os.Getenv("TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE") != "1" {
		t.Skip("Set TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE=1 on a test account enabled for secondary assignment at site creation")
	}

	tests := []struct {
		cloud, connection, config string
	}{
		{"aws", "SOCKET_AWS1500", `
		secondary_vsocket = {
			eni_ip_address = "%[1]s.10"
			eni_ip_subnet  = "%[1]s.0/24"
			route_table_id = "rtb-acctest"
		}`},
		{"azure", "SOCKET_AZ1500", `
		secondary_vsocket = {
			interface_ip = "%[1]s.10"
			floating_ip  = "%[1]s.20"
		}`},
		{"gcp", "SOCKET_GCP1500", `
		is_cloud_router = true
		secondary_vsocket = {
			interface_ip     = "%[1]s.10"
			load_balancer_ip = "%[1]s.20"
		}`},
	}

	const address = "cato_socket_site.ha"
	for _, tc := range tests {
		t.Run(tc.cloud, func(t *testing.T) {
			prefix := acc.GetRandNetworkPrefix()
			name := acc.GetRandName("inline-ha-" + tc.cloud)
			config := acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_socket_site" "ha" {
	name            = %q
	connection_type = %q
	site_type       = "BRANCH"
	native_range = {
		native_network_range = "%s.0/24"
		local_ip             = "%s.5"
	}
	site_location = {
		country_code = "GB"
		timezone     = "Europe/London"
	}
	%s
}
output "primary_serial" {
	value = one([for socket in cato_socket_site.ha.sockets : socket.serial_number if socket.is_primary])
}
output "secondary_serial" {
	value = one([for socket in cato_socket_site.ha.sockets : socket.serial_number if !socket.is_primary])
}
`, name, tc.connection, prefix, prefix, fmt.Sprintf(tc.config, prefix))

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
				PreCheck:                 acc.CheckCMAVars(t),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(address, "sockets.#", "2"),
							resource.TestCheckResourceAttr(address, "native_range.local_ip", prefix+".5"),
							resource.TestMatchOutput("primary_serial", regexp.MustCompile(`^.+$`)),
							resource.TestMatchOutput("secondary_serial", regexp.MustCompile(`^.+$`)),
						),
					},
					{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
				},
			})
		})
	}
}
