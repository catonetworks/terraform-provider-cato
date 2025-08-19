package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// activePeriodModifier ensures active_period changes are properly detected
// by comparing configuration values with state values and forcing changes
// when they differ, even if semantic equality would consider them equal
type activePeriodModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m activePeriodModifier) Description(_ context.Context) string {
	return "Forces detection of active_period changes by comparing configuration with API state"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m activePeriodModifier) MarkdownDescription(_ context.Context) string {
	return "Forces detection of `active_period` changes by comparing configuration with API state"
}

// PlanModifyObject implements the plan modification logic
func (m activePeriodModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Handle case where active_period is not present in config (null)
	// but we want to set default values
	if req.ConfigValue.IsNull() {
		tflog.Warn(ctx, "ActivePeriod plan modifier: Config is null, setting default values")
		
		// Create default object with use_effective_from = false and use_expires_at = false
		defaultAttrs := map[string]attr.Value{
			"effective_from":     types.StringNull(),
			"expires_at":         types.StringNull(),
			"use_effective_from": types.BoolValue(false),
			"use_expires_at":     types.BoolValue(false),
		}
		
		// Define the object type for active_period
		objectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"effective_from":     types.StringType,
				"expires_at":         types.StringType,
				"use_effective_from": types.BoolType,
				"use_expires_at":     types.BoolType,
			},
		}
		
		defaultObjectValue, diags := types.ObjectValue(objectType.AttrTypes, defaultAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		resp.PlanValue = defaultObjectValue
		return
	}
	
	// Skip if we don't have state or state is unknown
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Get the configuration and state values
	configAttrs := req.ConfigValue.Attributes()
	stateAttrs := req.StateValue.Attributes()

	tflog.Warn(ctx, "ActivePeriod plan modifier: Checking config vs state", map[string]interface{}{
		"config_effective_from": configAttrs["effective_from"],
		"config_expires_at":     configAttrs["expires_at"],
		"state_effective_from":  stateAttrs["effective_from"],
		"state_expires_at":      stateAttrs["expires_at"],
	})

	// Check if configuration differs from state for key fields
	configEffectiveFrom, configEffectiveFromOk := configAttrs["effective_from"].(types.String)
	configExpiresAt, configExpiresAtOk := configAttrs["expires_at"].(types.String)
	stateEffectiveFrom, stateEffectiveFromOk := stateAttrs["effective_from"].(types.String)
	stateExpiresAt, stateExpiresAtOk := stateAttrs["expires_at"].(types.String)

	if !configEffectiveFromOk || !configExpiresAtOk || !stateEffectiveFromOk || !stateExpiresAtOk {
		tflog.Warn(ctx, "ActivePeriod plan modifier: Failed to extract field values")
		return
	}

	// Check if there's a meaningful difference that should trigger an update
	changeDetected := false

	// Compare effective_from: null config vs non-null state should trigger change
	if configEffectiveFrom.IsNull() && !stateEffectiveFrom.IsNull() && !stateEffectiveFrom.IsUnknown() && stateEffectiveFrom.ValueString() != "" {
		tflog.Warn(ctx, "ActivePeriod plan modifier: Detected effective_from removal", map[string]interface{}{
			"config": "null",
			"state":  stateEffectiveFrom.ValueString(),
		})
		changeDetected = true
	}

	// Compare expires_at: null config vs non-null state should trigger change
	if configExpiresAt.IsNull() && !stateExpiresAt.IsNull() && !stateExpiresAt.IsUnknown() && stateExpiresAt.ValueString() != "" {
		tflog.Warn(ctx, "ActivePeriod plan modifier: Detected expires_at removal", map[string]interface{}{
			"config": "null",
			"state":  stateExpiresAt.ValueString(),
		})
		changeDetected = true
	}

	// Compare effective_from: non-null config vs different state should trigger change
	if !configEffectiveFrom.IsNull() && !configEffectiveFrom.IsUnknown() && configEffectiveFrom.ValueString() != "" {
		if stateEffectiveFrom.IsNull() || stateEffectiveFrom.IsUnknown() ||
			configEffectiveFrom.ValueString() != stateEffectiveFrom.ValueString() {
			tflog.Warn(ctx, "ActivePeriod plan modifier: Detected effective_from change", map[string]interface{}{
				"config": configEffectiveFrom.ValueString(),
				"state":  stateEffectiveFrom.ValueString(),
			})
			changeDetected = true
		}
	}

	// Compare expires_at: non-null config vs different state should trigger change
	if !configExpiresAt.IsNull() && !configExpiresAt.IsUnknown() && configExpiresAt.ValueString() != "" {
		if stateExpiresAt.IsNull() || stateExpiresAt.IsUnknown() ||
			configExpiresAt.ValueString() != stateExpiresAt.ValueString() {
			tflog.Warn(ctx, "ActivePeriod plan modifier: Detected expires_at change", map[string]interface{}{
				"config": configExpiresAt.ValueString(),
				"state":  stateExpiresAt.ValueString(),
			})
			changeDetected = true
		}
	}

	if changeDetected {
		tflog.Warn(ctx, "ActivePeriod plan modifier: Forcing update due to detected changes")

		// Create a new object with computed values based on configuration
		newAttrs := make(map[string]attr.Value)

		// Copy configuration values
		newAttrs["effective_from"] = configEffectiveFrom
		newAttrs["expires_at"] = configExpiresAt

		// Compute use_* flags based on configuration presence
		if configEffectiveFrom.IsNull() || configEffectiveFrom.IsUnknown() || configEffectiveFrom.ValueString() == "" {
			newAttrs["use_effective_from"] = types.BoolValue(false)
		} else {
			newAttrs["use_effective_from"] = types.BoolValue(true)
		}

		if configExpiresAt.IsNull() || configExpiresAt.IsUnknown() || configExpiresAt.ValueString() == "" {
			newAttrs["use_expires_at"] = types.BoolValue(false)
		} else {
			newAttrs["use_expires_at"] = types.BoolValue(true)
		}

		// Create new object value
		objectType := req.ConfigValue.Type(ctx).(types.ObjectType)
		newObjectValue, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.PlanValue = newObjectValue
	}
}

// ActivePeriodModifier returns a new active_period plan modifier
func ActivePeriodModifier() planmodifier.Object {
	return activePeriodModifier{}
}
