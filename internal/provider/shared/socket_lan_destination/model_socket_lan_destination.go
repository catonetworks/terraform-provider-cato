package landestination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
	ipaddr "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/ip_range"
)

// SocketLanDestination represents destination criteria for network rules
type SocketLanDestination struct {
	Vlan              types.List `tfsdk:"vlan" json:"vlan,omitempty"`
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
}

var SocketLanDestinationAttrTypes = map[string]attr.Type{
	"vlan":                types.ListType{ElemType: types.Int64Type},
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: ipaddr.RangeObjectType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"system_group":        types.SetType{ElemType: idname.RefObjectType},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
}
var SocketLanDestinationObjectType = types.ObjectType{AttrTypes: SocketLanDestinationAttrTypes}
