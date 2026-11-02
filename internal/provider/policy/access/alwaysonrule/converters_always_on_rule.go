package alwaysonrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func buildAlwaysOnRuleInput(ctx context.Context, plan AlwaysOnRule, diags *diag.Diagnostics) cato_models.ZtnaAlwaysOnAddRuleInput {
	var rule AlwaysOnRuleData
	diags.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return cato_models.ZtnaAlwaysOnAddRuleInput{}
	}
	var duration AlwaysOnRuleBypassDuration
	diags.Append(rule.BypassDuration.As(ctx, &duration, basetypes.ObjectAsOptions{})...)
	var source AlwaysOnRuleSource
	diags.Append(rule.Source.As(ctx, &source, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return cato_models.ZtnaAlwaysOnAddRuleInput{}
	}
	antiTamper := cato_models.AntiTamperModeEnum(rule.AntiTamperMode.ValueString())
	input := cato_models.ZtnaAlwaysOnAddRuleInput{
		Rule: &cato_models.ZtnaAlwaysOnAddRuleDataInput{
			Action:          cato_models.ZtnaAlwaysOnRuleActionEnum(rule.Action.ValueString()),
			AllowFailOpen:   rule.AllowFailOpen.ValueBool(),
			AllowUserBypass: rule.AllowUserBypass.ValueBool(),
			AntiTamperMode:  &antiTamper,
			BypassDuration: &cato_models.ZtnaAlwaysOnTimeInput{
				Time: duration.Time.ValueInt64(),
				Unit: cato_models.ZtnaAlwaysOnTimeUnit(duration.Unit.ValueString()),
			},
			Description:          rule.Description.ValueString(),
			Enabled:              rule.Enabled.ValueBool(),
			Name:                 rule.Name.ValueString(),
			DevicePostureProfile: idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, rule.DevicePostureProfile, diags),
			Platform:             convert.PrepareStrings[cato_models.OperatingSystem](ctx, rule.Platform, diags),
			Source: &cato_models.ZtnaAlwaysOnSourceInput{
				User:       idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, source.Users, diags),
				UsersGroup: idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, source.UserGroups, diags),
			},
		},
	}
	// The API's required lists must be JSON arrays even when no criteria are configured.
	if input.Rule.Platform == nil {
		input.Rule.Platform = []cato_models.OperatingSystem{}
	}
	input.At = buildAlwaysOnPosition(ctx, plan.At, diags)
	return input
}

func buildAlwaysOnPosition(ctx context.Context, at types.Object, diags *diag.Diagnostics) *cato_models.PolicyRulePositionInput {
	if at.IsNull() {
		return nil
	}
	var position policyrule.Position
	diags.Append(at.As(ctx, &position, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	positionEnum := cato_models.PolicyRulePositionEnum(position.Position.ValueString())
	return &cato_models.PolicyRulePositionInput{Position: &positionEnum, Ref: convert.StringPointerForOptionalInput(position.Ref)}
}

func buildAlwaysOnRuleID(ctx context.Context, state AlwaysOnRule, diags *diag.Diagnostics) string {
	var rule AlwaysOnRuleData
	diags.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return ""
	}
	if !utils.HasValue(rule.ID) || rule.ID.ValueString() == "" {
		diags.AddError("Invalid Always On rule ID", "A known, non-empty rule ID is required in state.")
		return ""
	}
	return rule.ID.ValueString()
}

func buildAlwaysOnUpdateRuleInput(ctx context.Context, plan, state AlwaysOnRule,
	diags *diag.Diagnostics,
) cato_models.ZtnaAlwaysOnUpdateRuleInput {
	id := buildAlwaysOnRuleID(ctx, state, diags)
	input := buildAlwaysOnRuleInput(ctx, plan, diags)
	if diags.HasError() {
		return cato_models.ZtnaAlwaysOnUpdateRuleInput{}
	}
	rule := input.Rule
	return cato_models.ZtnaAlwaysOnUpdateRuleInput{ID: id, Rule: &cato_models.ZtnaAlwaysOnUpdateRuleDataInput{
		Action: &rule.Action, AllowFailOpen: &rule.AllowFailOpen, AllowUserBypass: &rule.AllowUserBypass,
		AntiTamperMode: rule.AntiTamperMode,
		BypassDuration: &cato_models.ZtnaAlwaysOnTimeUpdateInput{Time: &rule.BypassDuration.Time, Unit: &rule.BypassDuration.Unit},
		Description:    &rule.Description, DevicePostureProfile: rule.DevicePostureProfile,
		Enabled: &rule.Enabled, Name: &rule.Name, Platform: rule.Platform,
		Source: &cato_models.ZtnaAlwaysOnSourceUpdateInput{User: rule.Source.User, UsersGroup: rule.Source.UsersGroup},
	}}
}

func buildAlwaysOnMoveRuleInput(ctx context.Context, plan, state AlwaysOnRule, id string,
	diags *diag.Diagnostics,
) *cato_models.PolicyMoveRuleInput {
	if plan.At.IsNull() || plan.At.Equal(state.At) {
		return nil
	}
	position := buildAlwaysOnPosition(ctx, plan.At, diags)
	if diags.HasError() {
		return nil
	}
	return &cato_models.PolicyMoveRuleInput{ID: id, To: position}
}

func buildAlwaysOnRemoveRuleInput(ctx context.Context, state AlwaysOnRule,
	diags *diag.Diagnostics,
) cato_models.ZtnaAlwaysOnRemoveRuleInput {
	return cato_models.ZtnaAlwaysOnRemoveRuleInput{ID: buildAlwaysOnRuleID(ctx, state, diags)}
}

func parseAlwaysOnRule(ctx context.Context,
	rule *cato.PolicyZtnaAlwaysOnAddRule_Policy_ZtnaAlwaysOn_AddRule_Rule_Rule, diags *diag.Diagnostics,
) types.Object {
	return parseAlwaysOnRuleData(ctx, rule, AlwaysOnRuleData{
		BypassDuration:       parseAlwaysOnBypassDuration(ctx, rule.BypassDuration.Time, rule.BypassDuration.Unit, diags),
		DevicePostureProfile: idname.ParseIDNameSet(ctx, rule.DevicePostureProfile, diags),
		Section:              idname.ParseIDName(ctx, rule.Section, diags),
		Source: parseAlwaysOnSource(ctx, idname.ParseIDNameSet(ctx, rule.Source.User, diags),
			idname.ParseIDNameSet(ctx, rule.Source.UsersGroup, diags), diags),
	}, diags)
}

func parseAlwaysOnPolicyRule(ctx context.Context,
	rule *cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy_Rules_Rule, diags *diag.Diagnostics,
) types.Object {
	return parseAlwaysOnRuleData(ctx, rule, AlwaysOnRuleData{
		BypassDuration:       parseAlwaysOnBypassDuration(ctx, rule.BypassDuration.Time, rule.BypassDuration.Unit, diags),
		DevicePostureProfile: idname.ParseIDNameSet(ctx, rule.DevicePostureProfile, diags),
		Section:              idname.ParseIDName(ctx, rule.Section, diags),
		Source: parseAlwaysOnSource(ctx, idname.ParseIDNameSet(ctx, rule.Source.User, diags),
			idname.ParseIDNameSet(ctx, rule.Source.UsersGroup, diags), diags),
	}, diags)
}

func parseUpdatedAlwaysOnRule(ctx context.Context,
	rule *cato.PolicyZtnaAlwaysOnUpdateRule_Policy_ZtnaAlwaysOn_UpdateRule_Rule_Rule, diags *diag.Diagnostics,
) types.Object {
	return parseAlwaysOnRuleData(ctx, rule, AlwaysOnRuleData{
		BypassDuration:       parseAlwaysOnBypassDuration(ctx, rule.BypassDuration.Time, rule.BypassDuration.Unit, diags),
		DevicePostureProfile: idname.ParseIDNameSet(ctx, rule.DevicePostureProfile, diags),
		Section:              idname.ParseIDName(ctx, rule.Section, diags),
		Source: parseAlwaysOnSource(ctx, idname.ParseIDNameSet(ctx, rule.Source.User, diags),
			idname.ParseIDNameSet(ctx, rule.Source.UsersGroup, diags), diags),
	}, diags)
}

func parseMovedAlwaysOnRule(ctx context.Context,
	rule *cato.PolicyZtnaAlwaysOnMoveRule_Policy_ZtnaAlwaysOn_MoveRule_Rule_Rule, diags *diag.Diagnostics,
) types.Object {
	return parseAlwaysOnRuleData(ctx, rule, AlwaysOnRuleData{
		BypassDuration:       parseAlwaysOnBypassDuration(ctx, rule.BypassDuration.Time, rule.BypassDuration.Unit, diags),
		DevicePostureProfile: idname.ParseIDNameSet(ctx, rule.DevicePostureProfile, diags),
		Section:              idname.ParseIDName(ctx, rule.Section, diags),
		Source: parseAlwaysOnSource(ctx, idname.ParseIDNameSet(ctx, rule.Source.User, diags),
			idname.ParseIDNameSet(ctx, rule.Source.UsersGroup, diags), diags),
	}, diags)
}

func parseAlwaysOnBypassDuration(ctx context.Context, time int64, unit cato_models.ZtnaAlwaysOnTimeUnit,
	diags *diag.Diagnostics,
) types.Object {
	duration, durationDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleBypassDurationAttrTypes, AlwaysOnRuleBypassDuration{
		Time: types.Int64Value(time),
		Unit: types.StringValue(string(unit)),
	})
	diags.Append(durationDiags...)
	return duration
}

func parseAlwaysOnSource(ctx context.Context, users, groups types.Set, diags *diag.Diagnostics) types.Object {
	if users.IsNull() {
		users = types.SetValueMust(idname.RefObjectType, nil)
	}
	if groups.IsNull() {
		groups = types.SetValueMust(idname.RefObjectType, nil)
	}
	source, sourceDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleSourceAttrTypes, AlwaysOnRuleSource{
		Users: users, UserGroups: groups,
	})
	diags.Append(sourceDiags...)
	return source
}

// Both SDK operations expose these scalar fields, but use distinct nested response types.
type alwaysOnRuleFields interface {
	GetAction() *cato_models.ZtnaAlwaysOnRuleActionEnum
	GetAllowFailOpen() bool
	GetAllowUserBypass() bool
	GetAntiTamperMode() *cato_models.AntiTamperModeEnum
	GetDescription() string
	GetEnabled() bool
	GetID() string
	GetIndex() int64
	GetName() string
	GetPlatform() []cato_models.OperatingSystem
}

func parseAlwaysOnRuleData(ctx context.Context, rule alwaysOnRuleFields, state AlwaysOnRuleData, diags *diag.Diagnostics) types.Object {
	antiTamper := cato_models.AntiTamperModeEnumOff
	if mode := rule.GetAntiTamperMode(); mode != nil {
		antiTamper = *mode
	}
	if state.DevicePostureProfile.IsNull() {
		state.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, nil)
	}
	state.Action = types.StringValue(string(*rule.GetAction()))
	state.AllowFailOpen = types.BoolValue(rule.GetAllowFailOpen())
	state.AllowUserBypass = types.BoolValue(rule.GetAllowUserBypass())
	state.AntiTamperMode = types.StringValue(string(antiTamper))
	state.Description = types.StringValue(rule.GetDescription())
	state.Enabled = types.BoolValue(rule.GetEnabled())
	state.ID = types.StringValue(rule.GetID())
	state.Index = types.Int64Value(rule.GetIndex())
	state.Name = types.StringValue(rule.GetName())
	state.Platform = convert.StringSet(ctx, rule.GetPlatform(), diags)
	if state.Platform.IsNull() {
		state.Platform = types.SetValueMust(types.StringType, nil)
	}
	value, stateDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, state)
	diags.Append(stateDiags...)
	return value
}
