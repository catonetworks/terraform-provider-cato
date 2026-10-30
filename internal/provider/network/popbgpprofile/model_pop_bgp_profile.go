package popbgpprofile

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Communities types.Set    `tfsdk:"communities"`
}
