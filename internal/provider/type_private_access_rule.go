package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAccessRuleModel struct {
	ID               types.String     `tfsdk:"id"`
	Name             types.String     `tfsdk:"name"`
	Description      types.String     `tfsdk:"description"`
	Index            types.Int64      `tfsdk:"index"`
	Enabled          types.Bool       `tfsdk:"enabled"`
	Source           *Source          `tfsdk:"source"`
	Platforms        []types.String   `tfsdk:"platforms"`
	Countries        []IdNameRefModel `tfsdk:"countries"`
	Applications     []IdNameRefModel `tfsdk:"applications"`
	ConnectionOrigin []types.String   `tfsdk:"connection_origin"`
	Action           types.String     `tfsdk:"action"` // e.g. "ALLOW"
	Tracking         *Tracking        `tfsdk:"tracking"`
	Device           []IdNameRefModel `tfsdk:"device"`
	UserAttributes   *UserAttributes  `tfsdk:"user_attributes"`
	Schedule         *PolicySchedule  `tfsdk:"schedule"`
	ActivePeriod     types.Object     `tfsdk:"active_period"`

	// Section          *PrivAccessPolicySection `tfsdk:"section"` -- not available in the 1st phase
}

type PrivAccessPolicySection struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	SubpolicyID types.String `tfsdk:"subpolicy_id"`
}

type Source struct {
	Users      []IdNameRefModel `tfsdk:"users"`
	UserGroups []IdNameRefModel `tfsdk:"user_groups"`
}

type UserAttributes struct {
	RiskScore RiskScore `tfsdk:"risk_score"`
}

type RiskScore struct {
	Category types.String `tfsdk:"category"` // e.g. "ANY"
	Operator types.String `tfsdk:"operator"` // e.g. "GTE"
}

type PolicySchedule struct {
	ActiveOn        types.String          `tfsdk:"active_on"` // e.g. "ALWAYS"
	CustomRecurring PolicyCustomRecurring `tfsdk:"custom_recurring"`
	CustomTimeframe PolicyCustomTimeframe `tfsdk:"custom_timeframe"`
}

type PolicyCustomRecurring struct {
	Days []types.String `tfsdk:"days"`
	From types.String   `tfsdk:"from"`
	To   types.String   `tfsdk:"to"`
}

type PolicyCustomTimeframe struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type PolicyRuleActivePeriod struct {
	EffectiveFrom    types.String `tfsdk:"effective_from"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	UseEffectiveFrom types.Bool   `tfsdk:"use_effective_from"`
	UseExpiresAt     types.Bool   `tfsdk:"use_expires_at"`
}

type Tracking struct {
	Event PolicyRuleTrackingEvent `tfsdk:"event"`
	Alert PoliciRuleTrackingAlert `tfsdk:"alert"`
}

type PolicyRuleTrackingEvent struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type PoliciRuleTrackingAlert struct {
	Enabled           types.Bool       `tfsdk:"enabled"`
	Frequency         types.String     `tfsdk:"frequency"`
	MailingList       []IdNameRefModel `tfsdk:"mailing_list"`
	SubscriptionGroup []IdNameRefModel `tfsdk:"subscription_group"`
	Webhook           []IdNameRefModel `tfsdk:"webhook"`
}
