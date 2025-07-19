package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type Account struct {
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Tenancy     types.String `tfsdk:"tenancy"`
	Timezone    types.String `tfsdk:"timezone"`
	Type        types.String `tfsdk:"type"`
}
