package globalipranges

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func GetGlobalIPRangeValidator() GlobalIPRangeValidator {
	return GlobalIPRangeValidator{}
}

// GlobalIPRangeValidator validates the GlobalIPRange settings
type GlobalIPRangeValidator struct{}

func (v GlobalIPRangeValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	var ipRanges []tf.GlobalIPRange

	if !utils.HasValue(req.ConfigValue) || req.ConfigValue.IsUnknown() {
		return
	}

	// get ip ranges as a slice
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.ElementsAs(ctx, &ipRanges, false)) {
		return
	}

	v.ValidateGlobalIPRange(ipRanges, &resp.Diagnostics)
}

func (v GlobalIPRangeValidator) ValidateGlobalIPRange(tfRanges []tf.GlobalIPRange, diags *diag.Diagnostics) {
	uniqueNames := make(map[string]struct{})
	uniqueRanges := make(map[string]struct{})

	for _, ipRange := range tfRanges {
		rangeValue := ipRange.IPRange.ValueString()
		if rangeValue == "" {
			diags.AddError("Invalid Configuration", "ip_range cannot be empty")
			return
		}
		if !validators.IsValidIPRange(rangeValue) {
			diags.AddError("Invalid Configuration", fmt.Sprintf(
				"ip_range '%s' must be a valid IP address, CIDR block, or IP range in start-end format",
				rangeValue,
			))
			return
		}

		// check name
		rangeName := ipRange.Name.ValueString()
		if rangeName == "" {
			diags.AddError("Invalid Configuration", "ip range name cannot be empty")
			return
		}
		if _, exists := uniqueNames[rangeName]; exists {
			diags.AddError("Invalid Configuration", fmt.Sprintf("duplicate ip range name '%s'", rangeName))
			return
		}
		uniqueNames[rangeName] = struct{}{}

		// check for duplicate ranges
		if _, exists := uniqueRanges[rangeValue]; exists {
			diags.AddError("Invalid Configuration", fmt.Sprintf("duplicate ip range '%s'", rangeValue))
			return
		}
		uniqueRanges[rangeValue] = struct{}{}
	}
}

func (v GlobalIPRangeValidator) Description(_ context.Context) string {
	return "global ip range settings validation"
}

func (v GlobalIPRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ModifyPlan computes what ranges should be updated, created or deleted based on the config and state.
func (r *globalIPRangesResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var cfg *tf.GlobalIPRangesModel
	var cfgRanges []tf.GlobalIPRange
	var stateRanges []tf.GlobalIPRange
	state := &tf.GlobalIPRangesModel{} // avoid nil pointer dereference
	stateDefined := !req.State.Raw.IsNull()

	if req.Plan.Raw.IsNull() { // resource destruction
		return
	}

	// get config, validate it
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg == nil || !utils.HasValue(cfg.Ranges) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, cfg.Ranges.ElementsAs(ctx, &cfgRanges, false)) {
		return
	}
	ipRangeValidator := GetGlobalIPRangeValidator()
	ipRangeValidator.ValidateGlobalIPRange(cfgRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state
	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if state != nil && utils.HasValue(state.Ranges) {
			if utils.CheckErr(&resp.Diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
				return
			}
		}
	}

	// prepare plan based on state and config
	plan, _ := r.computePlan(ctx, stateRanges, cfgRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// computePlan compares the config and state and determines which ranges need to be created, updated, deleted or kept as is.
func (r *globalIPRangesResource) computePlan(ctx context.Context, state, config []tf.GlobalIPRange, diags *diag.Diagnostics,
) (newPlan *tf.GlobalIPRangesModel, planDetails ipRangePlanDetails) {
	var plan []tf.GlobalIPRange

	// map of state ranges by name
	stateMap := make(map[string]*tf.GlobalIPRange)
	for i, item := range state {
		stateMap[item.Name.ValueString()] = &state[i]
	}

	// iterate config, if found in state, use ID from state, otherwise ID is unknown (new resource)
	for _, cfgItem := range config {
		planItem := cfgItem
		if planItem.Description.IsNull() {
			planItem.Description = types.StringValue("") // API always returns empty string
		}

		planItem.ID = types.StringUnknown() // ID defaults to unknown, if not found in the state
		if stateItem, ok := stateMap[cfgItem.Name.ValueString()]; ok {
			planItem.ID = stateItem.ID // ID is from state
			if planItem.Equal(*stateItem) {
				planDetails.toKeep = append(planDetails.toKeep, planItem)
			} else {
				planDetails.toUpdate = append(planDetails.toUpdate, planItem)
			}
			delete(stateMap, cfgItem.Name.ValueString())
		} else {
			planDetails.toCreate = append(planDetails.toCreate, planItem)
		}
		plan = append(plan, planItem)
	}

	// items in the state no longer in config should be deleted
	for _, stateItem := range stateMap {
		planDetails.toDelete = append(planDetails.toDelete, *stateItem)
	}

	tfSet, setDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, plan)
	diags.Append(setDiags...)
	if diags.HasError() {
		return nil, planDetails
	}
	return &tf.GlobalIPRangesModel{Ranges: tfSet}, planDetails
}
