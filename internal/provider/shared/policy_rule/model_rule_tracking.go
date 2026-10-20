package policyrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
)

// Tracking - Graphql type Tracking
type Tracking struct {
	Event types.Object `tfsdk:"event" json:"event,omitempty"` // TrackingEvent
	Alert types.Object `tfsdk:"alert" json:"alert,omitempty"` // TrackingAlert
}

var TrackingAttrTypes = map[string]attr.Type{
	"event": types.ObjectType{AttrTypes: TrackingEventAttrTypes},
	"alert": types.ObjectType{AttrTypes: TrackingAlertAttrTypes},
}
var TrackingObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}

// TrackingEvent - Graphql type PolicyRuleTrackingEvent
type TrackingEvent struct {
	Enabled types.Bool `tfsdk:"enabled" json:"enabled,omitempty"`
}

var TrackingEventAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}
var TrackingEventObjectType = types.ObjectType{AttrTypes: TrackingEventAttrTypes}

// TrackingAlert alert settings for the rule
// Graphql type PolicyRuleTrackingAlert
type TrackingAlert struct {
	Enabled           types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Frequency         types.String `tfsdk:"frequency" json:"frequency,omitempty"`
	SubscriptionGroup types.Set    `tfsdk:"subscription_group" json:"subscription_group,omitempty"` // []idname.Ref
	Webhook           types.Set    `tfsdk:"webhook" json:"webhook,omitempty"`                       // []idname.Ref
	MailingList       types.Set    `tfsdk:"mailing_list" json:"mailing_list,omitempty"`             // []idname.Ref
}

var TrackingAlertAttrTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"subscription_group": types.SetType{ElemType: idname.RefObjectType},
	"webhook":            types.SetType{ElemType: idname.RefObjectType},
	"mailing_list":       types.SetType{ElemType: idname.RefObjectType},
}
var TrackingAlertObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
