package dnssettings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SettingsModel contains the account DNS settings represented in Terraform.
type SettingsModel struct {
	ID                                types.String `tfsdk:"id"`
	PrimaryServer                     types.String `tfsdk:"primary_server"`
	SecondaryServer                   types.String `tfsdk:"secondary_server"`
	Suffix                            types.List   `tfsdk:"suffix"`
	AcceptDNSRequestsOnLanInterfaceIP types.Bool   `tfsdk:"accept_dns_requests_on_lan_interface_ip"`
}
