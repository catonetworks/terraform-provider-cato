package allocatedip

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	IP          types.String `tfsdk:"ip"`
	Description types.String `tfsdk:"description"`
	PopLocation types.Object `tfsdk:"pop_location"`
	Type        types.String `tfsdk:"type"`
	Subnet      types.String `tfsdk:"subnet"`
	State       types.String `tfsdk:"state"`
}
