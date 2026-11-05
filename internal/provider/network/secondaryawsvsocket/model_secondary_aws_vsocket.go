package secondaryawsvsocket

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID           types.String `tfsdk:"id"`
	Site         types.Object `tfsdk:"site"`
	EniIPAddress types.String `tfsdk:"eni_ip_address"`
	EniIPSubnet  types.String `tfsdk:"eni_ip_subnet"`
	RouteTableID types.String `tfsdk:"route_table_id"`
}
