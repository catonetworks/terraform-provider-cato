package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func hydrateFirewallSectionPositionState(ctx context.Context, current types.Object) (types.Object, diag.Diagnostics) {
	position := ifwLastInPolicyPosition
	ref := types.StringNull()
	var diags diag.Diagnostics

	if !current.IsNull() && !current.IsUnknown() {
		var currentPosition PolicyRulePositionInput
		currentDiags := current.As(ctx, &currentPosition, basetypes.ObjectAsOptions{})
		diags.Append(currentDiags...)
		if !currentDiags.HasError() {
			if !currentPosition.Position.IsNull() && !currentPosition.Position.IsUnknown() {
				position = currentPosition.Position.ValueString()
			}
			if !currentPosition.Ref.IsNull() && !currentPosition.Ref.IsUnknown() {
				ref = currentPosition.Ref
			}
		}
	}

	result, resultDiags := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue(position),
			"ref":      ref,
		},
	)
	diags.Append(resultDiags...)
	return result, diags
}
