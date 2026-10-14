package customservice

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
)

// CustomService - Graphql type CustomService
type CustomService struct {
	Port      types.List   `tfsdk:"port" json:"port,omitempty"`             // []string
	PortRange types.Object `tfsdk:"port_range" json:"port_range,omitempty"` // portrange.PortRange
	Protocol  types.String `tfsdk:"protocol" json:"protocol,omitempty"`
}

var CustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": types.ObjectType{AttrTypes: portrange.PortRangeAttrTypes},
	"protocol":   types.StringType,
}
var CustomServiceObjectType = types.ObjectType{AttrTypes: CustomServiceAttrTypes}
