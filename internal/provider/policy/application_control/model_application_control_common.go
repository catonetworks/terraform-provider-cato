package appcontrol

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
	ipaddr "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/ip_range"
)

// ApplicationControlSourceAttrTypes matches ApplicationControlSource in the Cato API (WAN-like + country).
var ApplicationControlSourceAttrTypes = map[string]attr.Type{
	"country":             types.SetType{ElemType: idname.RefObjectType},
	"ip":                  types.ListType{ElemType: types.StringType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"site":                types.SetType{ElemType: idname.RefObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: ipaddr.RangeObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"user":                types.SetType{ElemType: idname.RefObjectType},
	"users_group":         types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"system_group":        types.SetType{ElemType: idname.RefObjectType},
}

var ApplicationControlSourceObjectType = types.ObjectType{AttrTypes: ApplicationControlSourceAttrTypes}
