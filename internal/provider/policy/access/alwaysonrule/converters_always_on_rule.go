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
			DevicePostureProfile: buildAlwaysOnReferences[cato_models.DeviceProfileRefInput](ctx, rule.DevicePostureProfile, diags),
			Platform:             convert.PrepareStrings[cato_models.OperatingSystem](ctx, rule.Platform, diags),
			Source: &cato_models.ZtnaAlwaysOnSourceInput{
				User:       buildAlwaysOnReferences[cato_models.UserRefInput](ctx, source.User, diags),
				UsersGroup: buildAlwaysOnReferences[cato_models.UsersGroupRefInput](ctx, source.UsersGroup, diags),
			},
		},
	}
	// The API's required lists must be JSON arrays even when no criteria are configured.
	if input.Rule.Platform == nil {
		input.Rule.Platform = []cato_models.OperatingSystem{}
	}
	if utils.HasValue(plan.At) {
		var position policyrule.Position
		diags.Append(plan.At.As(ctx, &position, basetypes.ObjectAsOptions{})...)
		positionEnum := cato_models.PolicyRulePositionEnum(position.Position.ValueString())
		input.At = &cato_models.PolicyRulePositionInput{
			Position: &positionEnum, Ref: convert.StringPointerForOptionalInput(position.Ref),
		}
	}
	return input
}

// buildAlwaysOnReferences resolves references by ID first, then by name, and preserves empty arrays.
func buildAlwaysOnReferences[T ~struct {
	By    cato_models.ObjectRefBy `json:"by"`
	Input string                  `json:"input"`
}](ctx context.Context, set types.Set, diags *diag.Diagnostics) []*T {
	var refs []idname.Ref
	diags.Append(set.ElementsAs(ctx, &refs, false)...)
	inputs := make([]*T, 0, len(refs))
	for _, ref := range refs {
		switch {
		case utils.HasValue(ref.ID) && ref.ID.ValueString() != "":
			inputs = append(inputs, &T{By: cato_models.ObjectRefByID, Input: ref.ID.ValueString()})
		case utils.HasValue(ref.Name) && ref.Name.ValueString() != "":
			inputs = append(inputs, &T{By: cato_models.ObjectRefByName, Input: ref.Name.ValueString()})
		default:
			diags.AddError("Invalid Always On rule reference", "Each reference must contain a known, non-empty ID or name.")
		}
	}
	return inputs
}

func parseAlwaysOnRule(ctx context.Context,
	rule *cato.PolicyZtnaAlwaysOnAddRule_Policy_ZtnaAlwaysOn_AddRule_Rule_Rule, diags *diag.Diagnostics,
) types.Object {
	duration, durationDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleBypassDurationAttrTypes, AlwaysOnRuleBypassDuration{
		Time: types.Int64Value(rule.BypassDuration.Time),
		Unit: types.StringValue(string(rule.BypassDuration.Unit)),
	})
	diags.Append(durationDiags...)
	section, sectionDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleSectionAttrTypes, AlwaysOnRuleSection{
		ID:   types.StringValue(rule.Section.ID),
		Name: types.StringValue(rule.Section.Name),
	})
	diags.Append(sectionDiags...)
	users := idname.ParseIDNameSet(ctx, rule.Source.User, diags)
	groups := idname.ParseIDNameSet(ctx, rule.Source.UsersGroup, diags)
	profiles := idname.ParseIDNameSet(ctx, rule.DevicePostureProfile, diags)
	if users.IsNull() {
		users = types.SetValueMust(idname.RefObjectType, nil)
	}
	if groups.IsNull() {
		groups = types.SetValueMust(idname.RefObjectType, nil)
	}
	if profiles.IsNull() {
		profiles = types.SetValueMust(idname.RefObjectType, nil)
	}
	source, sourceDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleSourceAttrTypes, AlwaysOnRuleSource{
		User: users, UsersGroup: groups,
	})
	diags.Append(sourceDiags...)
	antiTamper := cato_models.AntiTamperModeEnumOff
	if rule.AntiTamperMode != nil {
		antiTamper = *rule.AntiTamperMode
	}
	state, stateDiags := types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, AlwaysOnRuleData{
		Action:               types.StringValue(string(rule.Action)),
		AllowFailOpen:        types.BoolValue(rule.AllowFailOpen),
		AllowUserBypass:      types.BoolValue(rule.AllowUserBypass),
		AntiTamperMode:       types.StringValue(string(antiTamper)),
		BypassDuration:       duration,
		Description:          types.StringValue(rule.Description),
		DevicePostureProfile: profiles,
		Enabled:              types.BoolValue(rule.Enabled),
		ID:                   types.StringValue(rule.ID),
		Index:                types.Int64Value(rule.Index),
		Name:                 types.StringValue(rule.Name),
		Section:              section,
		Source:               source,
		Platform:             convert.StringSet(ctx, append([]cato_models.OperatingSystem{}, rule.Platform...), diags),
	})
	diags.Append(stateDiags...)
	return state
}
