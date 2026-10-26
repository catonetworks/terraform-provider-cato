package policyrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Position Parameters required to define the rule position
// Graphql input PolicyRulePositionInput
type Position struct {
	Position types.String `tfsdk:"position" json:"position,omitempty"`
	Ref      types.String `tfsdk:"ref" json:"ref,omitempty"`
}

var PositionAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}
var PositionObjectType = types.ObjectType{AttrTypes: PositionAttrTypes}
