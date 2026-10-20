package lansite

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
)

// SocketLanSite represents the site scope for the rule
type SocketLanSite struct {
	Site  types.Set `tfsdk:"site" json:"site,omitempty"`
	Group types.Set `tfsdk:"group" json:"group,omitempty"`
}

var SocketLanSiteAttrTypes = map[string]attr.Type{
	"site":  types.SetType{ElemType: idname.RefObjectType},
	"group": types.SetType{ElemType: idname.RefObjectType},
}
var SocketLanSiteObjectType = types.ObjectType{AttrTypes: SocketLanSiteAttrTypes}
