package siteloc

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SiteLocation struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	City        types.String `tfsdk:"city"`
}

var SiteLocationResourceAttrTypes = map[string]attr.Type{
	"country_code": types.StringType,
	"state_code":   types.StringType,
	"timezone":     types.StringType,
	"address":      types.StringType,
	"city":         types.StringType,
}
