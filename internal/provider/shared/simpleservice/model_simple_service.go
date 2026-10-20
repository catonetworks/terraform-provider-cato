package simpleservice

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SimpleService represents a simple named service
// Graphql type SimpleService
type SimpleService struct {
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

var SimpleServiceAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}
var SimpleServiceObjectType = types.ObjectType{AttrTypes: SimpleServiceAttrTypes}
