package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var PositionObjectType = types.ObjectType{AttrTypes: PositionAttrTypes}
var PositionAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}
var NameIDObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}
var NameIDAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
}
var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
var FromToDaysObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToDaysAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
	"days": types.ListType{ElemType: types.StringType},
}

// Rule -> Tracking
var TrackingObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingAttrTypes = map[string]attr.Type{
	"event": types.ObjectType{AttrTypes: TrackingEventAttrTypes},
	"alert": types.ObjectType{AttrTypes: TrackingAlertAttrTypes},
}
var TrackingEventObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingEventAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}
var TrackingAlertObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingAlertAttrTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"subscription_group": types.SetType{ElemType: NameIDObjectType},
	"webhook":            types.SetType{ElemType: NameIDObjectType},
	"mailing_list":       types.SetType{ElemType: NameIDObjectType},
}

var ScheduleObjectType = types.ObjectType{AttrTypes: ScheduleAttrTypes}
var ScheduleAttrTypes = map[string]attr.Type{
	"active_on":        types.StringType,
	"custom_timeframe": types.ObjectType{AttrTypes: FromToAttrTypes},
	"custom_recurring": types.ObjectType{AttrTypes: FromToDaysAttrTypes},
}

var CustomServiceObjectType = types.ObjectType{AttrTypes: CustomServiceAttrTypes}
var CustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": FromToObjectType,
	"protocol":   types.StringType,
}
