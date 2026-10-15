package lannetwork

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/landestination"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/lansite"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/lansource"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SocketLanNetworkRule represents the top-level resource structure
type SocketLanNetworkRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"`
	At   types.Object `tfsdk:"at" json:"at,omitempty"`
}

// SocketLanNetworkRuleData represents the rule data within the resource
type SocketLanNetworkRuleData struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Direction   types.String `tfsdk:"direction" json:"direction,omitempty"`
	Transport   types.String `tfsdk:"transport" json:"transport,omitempty"`
	Site        types.Object `tfsdk:"site" json:"site,omitempty"`
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`
	Destination types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Service     types.Object `tfsdk:"service" json:"service,omitempty"`
	Nat         types.Object `tfsdk:"nat" json:"nat,omitempty"`
}

// SocketLanService represents service criteria for network rules
type SocketLanService struct {
	Simple types.Set  `tfsdk:"simple" json:"simple,omitempty"`
	Custom types.List `tfsdk:"custom" json:"custom,omitempty"`
}

// SocketLanNat represents NAT settings for network rules
type SocketLanNat struct {
	Enabled types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	NatType types.String `tfsdk:"nat_type" json:"nat_type,omitempty"`
}

// Attr types for state management

var SocketLanNetworkRuleObjectType = types.ObjectType{AttrTypes: SocketLanNetworkRuleAttrTypes}
var SocketLanNetworkRuleAttrTypes = map[string]attr.Type{
	"rule": SocketLanNetworkRuleRuleObjectType,
	"at":   policyrule.PositionObjectType,
}

var SocketLanNetworkRuleRuleObjectType = types.ObjectType{AttrTypes: SocketLanNetworkRuleRuleAttrTypes}
var SocketLanNetworkRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	"direction":   types.StringType,
	"transport":   types.StringType,
	"site":        lansite.SocketLanSiteObjectType,
	"source":      lansource.SocketLanSourceObjectType,
	"destination": landestination.SocketLanDestinationObjectType,
	"service":     SocketLanServiceObjectType,
	"nat":         SocketLanNatObjectType,
}

var SocketLanServiceObjectType = types.ObjectType{AttrTypes: SocketLanServiceAttrTypes}
var SocketLanServiceAttrTypes = map[string]attr.Type{
	"simple": types.SetType{ElemType: SimpleServiceObjectType},
	"custom": types.ListType{ElemType: customservice.CustomServiceObjectType},
}

type SimpleService struct {
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

var SimpleServiceObjectType = types.ObjectType{AttrTypes: SimpleServiceAttrTypes}
var SimpleServiceAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

var SocketLanNatObjectType = types.ObjectType{AttrTypes: SocketLanNatAttrTypes}
var SocketLanNatAttrTypes = map[string]attr.Type{
	"enabled":  types.BoolType,
	"nat_type": types.StringType,
}
