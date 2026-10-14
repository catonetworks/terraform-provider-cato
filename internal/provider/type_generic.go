package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

var ActivePeriodObjectType = types.ObjectType{AttrTypes: ActivePeriodAttrTypes}
var ActivePeriodAttrTypes = map[string]attr.Type{
	"effective_from":     types.StringType,
	"expires_at":         types.StringType,
	"use_effective_from": types.BoolType,
	"use_expires_at":     types.BoolType,
}

var CustomServiceIPObjectType = types.ObjectType{AttrTypes: CustomServiceIPAttrTypes}
var CustomServiceIPAttrTypes = map[string]attr.Type{
	"name":     types.StringType,
	"ip":       types.StringType,
	"ip_range": FromToObjectType,
}
