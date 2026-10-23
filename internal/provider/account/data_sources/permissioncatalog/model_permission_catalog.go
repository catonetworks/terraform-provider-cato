package permissioncatalog

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model holds the permission catalog for one account.
type Model struct {
	ID        types.String `tfsdk:"id"`
	AccountID types.String `tfsdk:"account_id"`
	Resources types.Set    `tfsdk:"resources"`
}

var catalogAttrTypes = map[string]attr.Type{"resource": types.StringType, "supported_actions": types.SetType{ElemType: types.StringType}}
var catalogObjectType = types.ObjectType{AttrTypes: catalogAttrTypes}
