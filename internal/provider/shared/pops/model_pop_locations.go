package pops

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type PreferredPopLocationModel struct {
	PreferredOnly types.Bool   `tfsdk:"preferred_only"`
	Automatic     types.Bool   `tfsdk:"automatic"`
	Primary       types.Object `tfsdk:"primary"`   // IDNameRefModel
	Secondary     types.Object `tfsdk:"secondary"` // IDNameRefModel
}

var PreferredPopLocationModelTypes = map[string]attr.Type{
	"preferred_only": types.BoolType,
	"automatic":      types.BoolType,
	"primary":        types.ObjectType{AttrTypes: idname.ModelTypes},
	"secondary":      types.ObjectType{AttrTypes: idname.ModelTypes},
}
