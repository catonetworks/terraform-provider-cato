package activeperiod

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ActivePeriod struct {
	EffectiveFrom    types.String `tfsdk:"effective_from" json:"effective_from,omitempty"`
	ExpiresAt        types.String `tfsdk:"expires_at" json:"expires_at,omitempty"`
	UseEffectiveFrom types.Bool   `tfsdk:"use_effective_from" json:"use_effective_from,omitempty"`
	UseExpiresAt     types.Bool   `tfsdk:"use_expires_at" json:"use_expires_at,omitempty"`
}

var ActivePeriodAttrTypes = map[string]attr.Type{
	"effective_from":     types.StringType,
	"expires_at":         types.StringType,
	"use_effective_from": types.BoolType,
	"use_expires_at":     types.BoolType,
}

var ActivePeriodObjectType = types.ObjectType{AttrTypes: ActivePeriodAttrTypes}
