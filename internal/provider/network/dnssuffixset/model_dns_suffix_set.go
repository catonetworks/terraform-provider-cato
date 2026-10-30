package dnssuffixset

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Suffix types.List   `tfsdk:"suffix"`
}
