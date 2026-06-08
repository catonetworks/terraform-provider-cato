package validators

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// checkLocalIP checks if local_ip is within network_network_range.
// Returns nil if either parameter does not have a value.
// On error update diags and returns an error
func checkLocalIP(diags *diag.Diagnostics, tfLocalIP, tfSubnet types.String) error {
	if !utils.HasValue(tfLocalIP) || !utils.HasValue(tfSubnet) {
		return nil
	}

	localIP := tfLocalIP.ValueString()
	subnet := tfSubnet.ValueString()

	// Parse the local IP
	ip := net.ParseIP(localIP)
	if ip == nil {
		diags.AddError("Invalid Configuration", fmt.Sprintf("local_ip '%s' is not a valid IP address", localIP))
		return ErrConfig
	}

	// Parse the subnet CIDR
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("network_network_range '%s' is not a valid CIDR notation", subnet))
		return ErrConfig
	}

	// Check if the IP is within the subnet
	if !ipNet.Contains(ip) {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("Local IP '%s' is not within the Network Range: '%s'", localIP, subnet))
		return ErrConfig
	}

	return nil
}
