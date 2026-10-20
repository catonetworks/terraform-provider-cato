package policyschedule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ScheduleAttrTypes = map[string]attr.Type{
	"active_on":        types.StringType,
	"custom_timeframe": types.ObjectType{AttrTypes: CustomTimeframeAttrTypes},
	"custom_recurring": types.ObjectType{AttrTypes: FromToDaysAttrTypes},
}
var ScheduleObjectType = types.ObjectType{AttrTypes: ScheduleAttrTypes}

var CustomTimeframeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
var CustomTimeframeObjectType = types.ObjectType{AttrTypes: CustomTimeframeAttrTypes}

var FromToDaysAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
	"days": types.ListType{ElemType: types.StringType},
}
var FromToDaysObjectType = types.ObjectType{AttrTypes: FromToDaysAttrTypes}
