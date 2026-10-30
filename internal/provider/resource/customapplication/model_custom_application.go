package customapplication

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	Category    types.Set    `tfsdk:"category"`
	Criteria    types.Set    `tfsdk:"criteria"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}
