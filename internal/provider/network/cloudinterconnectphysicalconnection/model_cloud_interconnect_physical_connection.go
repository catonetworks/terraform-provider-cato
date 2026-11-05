package cloudinterconnectphysicalconnection

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	DownstreamBwLimit     types.String `tfsdk:"downstream_bw_limit"`
	EncapsulationMethod   types.String `tfsdk:"encapsulation_method"`
	HARole                types.String `tfsdk:"ha_role"`
	ID                    types.String `tfsdk:"id"`
	PopLocation           types.Object `tfsdk:"pop_location"`
	PrivateCatoIP         types.String `tfsdk:"private_cato_ip"`
	PrivateSiteIP         types.String `tfsdk:"private_site_ip"`
	QinqVlanConfiguration types.Object `tfsdk:"qinq_vlan_configuration"`
	ServiceProviderName   types.String `tfsdk:"service_provider_name"`
	Site                  types.Object `tfsdk:"site"`
	Subnet                types.String `tfsdk:"subnet"`
	UpstreamBwLimit       types.String `tfsdk:"upstream_bw_limit"`
	Vlan                  types.Int64  `tfsdk:"vlan"`
}
