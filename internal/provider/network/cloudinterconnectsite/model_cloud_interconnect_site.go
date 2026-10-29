package cloudinterconnectsite

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	SiteType     types.String `tfsdk:"site_type"`
	SiteLocation types.Object `tfsdk:"site_location"`
}
