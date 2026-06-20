package tfmodel

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GlobalIPRangesModel struct {
	Ranges types.Set `tfsdk:"ranges"` // []GlobalIPRange
}

type GlobalIPRange struct {
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	IPRange     types.String `tfsdk:"ip_range"`
	Name        types.String `tfsdk:"name"`
}

var GlobalIPRangeTypes = map[string]attr.Type{
	"description": types.StringType,
	"id":          types.StringType,
	"ip_range":    types.StringType,
	"name":        types.StringType,
}

func (a GlobalIPRange) Equal(b GlobalIPRange) bool {
	return a.Description.Equal(b.Description) &&
		a.ID.Equal(b.ID) &&
		a.IPRange.Equal(b.IPRange) &&
		a.Name.Equal(b.Name)
}
