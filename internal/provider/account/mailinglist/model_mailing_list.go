package mailinglist

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Address types.Set    `tfsdk:"address"`
	Admin   types.Set    `tfsdk:"admin"`
}
