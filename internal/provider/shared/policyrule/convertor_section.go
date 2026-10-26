package policyrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func DefaultSectionPosition(ctx context.Context, current types.Object) (types.Object, diag.Diagnostics) {
	position := string(cato_models.PolicyRulePositionEnumLastInPolicy)
	ref := types.StringNull()
	var diags diag.Diagnostics

	if !current.IsNull() && !current.IsUnknown() {
		var currentPosition Position
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
