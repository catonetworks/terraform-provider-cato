package provider

import (
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAccessRuleModel struct {
	Action            types.String `tfsdk:"action"`             // e.g. "ALLOW"
	ActivePeriod      types.Object `tfsdk:"active_period"`      // PolicyRuleActivePeriod
	Applications      types.List   `tfsdk:"applications"`       // []IdNameRefModel
	ConnectionOrigins types.List   `tfsdk:"connection_origins"` // []string
	Countries         types.List   `tfsdk:"countries"`          // []IdNameRefModel
	Description       types.String `tfsdk:"description"`
	Devices           types.List   `tfsdk:"devices"` // []IdNameRefModel
	Enabled           types.Bool   `tfsdk:"enabled"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Platforms         types.List   `tfsdk:"platforms"`       // []string
	Schedule          types.Object `tfsdk:"schedule"`        // PolicySchedule
	Source            types.Object `tfsdk:"source"`          // Source
	Tracking          types.Object `tfsdk:"tracking"`        // Tracking
	UserAttributes    types.Object `tfsdk:"user_attributes"` // UserAttributes

	// Section          *PrivAccessPolicySection `tfsdk:"section"` -- not available in the 1st phase
}

type PrivAccessPolicySection struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	SubpolicyID types.String `tfsdk:"subpolicy_id"`
}

var PrivAccessPolicySectionTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"subpolicy_id": types.StringType,
}

type Source struct {
	Users      types.List `tfsdk:"users"`       // IdNameRefModel
	UserGroups types.List `tfsdk:"user_groups"` // IdNameRefModel
}

var SourceTypes = map[string]attr.Type{
	"users":       types.ListType{ElemType: types.ObjectType{AttrTypes: parse.IdNameRefModelTypes}},
	"user_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: parse.IdNameRefModelTypes}},
}

type UserAttributes struct {
	RiskScore types.Object `tfsdk:"risk_score"`
}

var UserAttributesTypes = map[string]attr.Type{
	"risk_score": types.ObjectType{AttrTypes: RiskScoreTypes},
}

type RiskScore struct {
	Category types.String `tfsdk:"category"` // e.g. "ANY"
	Operator types.String `tfsdk:"operator"` // e.g. "GTE"
}

var RiskScoreTypes = map[string]attr.Type{
	"category": types.StringType,
	"operator": types.StringType,
}

type PolicySchedule struct {
	ActiveOn        types.String `tfsdk:"active_on"`        // e.g. "ALWAYS"
	CustomRecurring types.Object `tfsdk:"custom_recurring"` // PolicyCustomRecurring
	CustomTimeframe types.Object `tfsdk:"custom_timeframe"` // PolicyCustomTimeframe
}

var PolicyScheduleTypes = map[string]attr.Type{
	"active_on":        types.StringType,
	"custom_recurring": types.ObjectType{AttrTypes: PolicyCustomRecurringTypes},
	"custom_timeframe": types.ObjectType{AttrTypes: PolicyCustomTimeframeTypes},
}

type PolicyCustomRecurring struct {
	Days types.List   `tfsdk:"days"` // []string
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var PolicyCustomRecurringTypes = map[string]attr.Type{
	"days": types.ListType{ElemType: types.StringType},
	"from": types.StringType,
	"to":   types.StringType,
}

type PolicyCustomTimeframe struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var PolicyCustomTimeframeTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

type PolicyRuleActivePeriod struct {
	EffectiveFrom    types.String `tfsdk:"effective_from"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	UseEffectiveFrom types.Bool   `tfsdk:"use_effective_from"`
	UseExpiresAt     types.Bool   `tfsdk:"use_expires_at"`
}

var PolicyRuleActivePeriodTypes = map[string]attr.Type{
	"effective_from":     types.StringType,
	"expires_at":         types.StringType,
	"use_effective_from": types.BoolType,
	"use_expires_at":     types.BoolType,
}

type PolicyRuleTracking struct {
	Event types.Object `tfsdk:"event"` // PolicyRuleTrackingEvent
	Alert types.Object `tfsdk:"alert"` // PoliciRuleTrackingAlert
}

var PolicyRuleTrackingTypes = map[string]attr.Type{
	"event": types.ObjectType{AttrTypes: PolicyRuleTrackingEventTypes},
	"alert": types.ObjectType{AttrTypes: PolicyRuleTrackingAlertTypes},
}

type PolicyRuleTrackingEvent struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

var PolicyRuleTrackingEventTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}

type PoliciRuleTrackingAlert struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Frequency         types.String `tfsdk:"frequency"`
	MailingList       types.List   `tfsdk:"mailing_list"`       // IdNameRefModel
	SubscriptionGroup types.List   `tfsdk:"subscription_group"` // IdNameRefModel
	Webhook           types.List   `tfsdk:"webhook"`            // IdNameRefModel
}

var PolicyRuleTrackingAlertTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"mailing_list":       types.ListType{ElemType: types.ObjectType{AttrTypes: parse.IdNameRefModelTypes}},
	"subscription_group": types.ListType{ElemType: types.ObjectType{AttrTypes: parse.IdNameRefModelTypes}},
	"webhook":            types.ListType{ElemType: types.ObjectType{AttrTypes: parse.IdNameRefModelTypes}},
}
