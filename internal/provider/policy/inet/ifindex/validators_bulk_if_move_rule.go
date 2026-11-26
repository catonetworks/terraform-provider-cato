package ifindex

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const configuredRuleEnablementKey = "configured_rule_enablement"

// ModifyPlan returns a previously configured value to API ownership when the
// user removes it. Terraform can otherwise propose the prior value unchanged.
func (r *ifwRulesIndexResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	previous, diags := loadConfiguredRuleEnablement(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if len(previous) == 0 || resp.Diagnostics.HasError() {
		return
	}

	var rules types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("rule_data"), &rules)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for key, value := range rules.Elements() {
		if _, wasConfigured := previous[key]; !wasConfigured || value.IsNull() || value.IsUnknown() {
			continue
		}
		if value.(types.Object).Attributes()["enabled"].IsNull() {
			enabled := path.Root("rule_data").AtMapKey(key).AtName("enabled")
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, enabled, types.BoolUnknown())...)
		}
	}
}

func legacyEnablementOnlyUpdate(prior, plan, config IfwRulesIndex) bool {
	if !prior.SectionToStartAfterID.Equal(plan.SectionToStartAfterID) {
		return false
	}
	sameSections, _ := compareBulkRows(prior.SectionData, plan.SectionData, config.SectionData)
	sameRules, enabledChanged := compareBulkRows(prior.RuleData, plan.RuleData, config.RuleData)
	return sameSections && sameRules && enabledChanged
}

// compareBulkRows ignores enablement changes and unknown values that Terraform
// introduced for omitted computed fields. A configured unknown may affect ordering.
func compareBulkRows(prior, plan, config types.Map) (match, enabledChanged bool) {
	if prior.Equal(plan) {
		return true, false
	}
	if prior.IsUnknown() || plan.IsUnknown() {
		return false, false
	}
	priorRows, planRows, configRows := prior.Elements(), plan.Elements(), config.Elements()
	if len(priorRows) != len(planRows) {
		return false, false
	}
	for key, planValue := range planRows {
		priorRow, ok := priorRows[key].(types.Object)
		if !ok {
			return false, false
		}
		if planValue.Equal(priorRow) {
			continue
		}
		if planValue.IsUnknown() || planValue.IsNull() || priorRow.IsUnknown() || priorRow.IsNull() {
			return false, false
		}
		planRow := planValue.(types.Object)
		configRow, _ := configRows[key].(types.Object)
		sameRow, changed := compareBulkRow(priorRow, planRow, configRow)
		if !sameRow {
			return false, false
		}
		enabledChanged = enabledChanged || changed
	}
	return true, enabledChanged
}

func compareBulkRow(prior, plan, config types.Object) (match, enabledChanged bool) {
	priorAttrs, configAttrs := prior.Attributes(), config.Attributes()
	for name, value := range plan.Attributes() {
		if value.Equal(priorAttrs[name]) {
			continue
		}
		if name == "enabled" {
			enabledChanged = true
			continue
		}
		if configured := configAttrs[name]; !value.IsUnknown() || configured == nil || !configured.IsNull() {
			return false, false
		}
	}
	return true, enabledChanged
}

func saveConfiguredRuleEnablement(ctx context.Context, config tfsdk.Config, state interface {
	SetKey(context.Context, string, []byte) diag.Diagnostics
}) diag.Diagnostics {
	var rules types.Map
	diags := config.GetAttribute(ctx, path.Root("rule_data"), &rules)
	if diags.HasError() {
		return diags
	}

	configured := make(map[string]struct{})
	for key, value := range rules.Elements() {
		if !value.IsNull() && !value.IsUnknown() && !value.(types.Object).Attributes()["enabled"].IsNull() {
			configured[key] = struct{}{}
		}
	}
	if len(configured) == 0 {
		return state.SetKey(ctx, configuredRuleEnablementKey, nil)
	}

	data, err := json.Marshal(configured)
	if err != nil {
		diags.AddError("Failed to record configured bulk rule enablement", err.Error())
		return diags
	}
	diags.Append(state.SetKey(ctx, configuredRuleEnablementKey, data)...)
	return diags
}

func loadConfiguredRuleEnablement(ctx context.Context, state interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
}) (map[string]struct{}, diag.Diagnostics) {
	data, diags := state.GetKey(ctx, configuredRuleEnablementKey)
	var configured map[string]struct{}
	if len(data) != 0 && !diags.HasError() {
		if err := json.Unmarshal(data, &configured); err != nil {
			diags.AddError("Failed to read configured bulk rule enablement", err.Error())
		}
	}
	return configured, diags
}
