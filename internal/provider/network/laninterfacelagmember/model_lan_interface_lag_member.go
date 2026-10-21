package laninterfacelagmember

import "github.com/hashicorp/terraform-plugin-framework/types"

type LagMember struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	InterfaceID types.String `tfsdk:"interface_id"`
	Name        types.String `tfsdk:"name"`
	DestType    types.String `tfsdk:"dest_type"`
}
