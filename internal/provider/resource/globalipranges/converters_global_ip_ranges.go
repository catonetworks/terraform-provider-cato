package globalipranges

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// prepareCreateInput converts the list of ranges to be created to the input format required by the Create API
func (r *globalIPRangesResource) prepareCreateInput(ranges []tf.GlobalIPRange) []*cato_models.CreateGlobalIPRangeInput {
	input := make([]*cato_models.CreateGlobalIPRangeInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.CreateGlobalIPRangeInput{
			IPRange: globalIPRange.IPRange.ValueString(),
			Name:    globalIPRange.Name.ValueString(),
		}
		if utils.HasValue(globalIPRange.Description) {
			inputItem.Description = globalIPRange.Description.ValueStringPointer()
		}
		input = append(input, &inputItem)
	}
	return input
}

// prepareUpdateInput converts the list of ranges to be updated to the input format required by the Update API
func (r *globalIPRangesResource) prepareUpdateInput(ranges []tf.GlobalIPRange) []*cato_models.UpdateGlobalIPRangeInput {
	input := make([]*cato_models.UpdateGlobalIPRangeInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.UpdateGlobalIPRangeInput{
			ID:      globalIPRange.ID.ValueString(),
			IPRange: globalIPRange.IPRange.ValueStringPointer(),
			Name:    globalIPRange.Name.ValueStringPointer(),
		}
		if utils.HasValue(globalIPRange.Description) {
			inputItem.Description = globalIPRange.Description.ValueStringPointer()
		}
		input = append(input, &inputItem)
	}
	return input
}

// prepareDeleteInput converts the list of ranges to be deleted to the input format required by the Delete API
func (r *globalIPRangesResource) prepareDeleteInput(ranges []tf.GlobalIPRange) []*cato_models.GlobalIPRangeRefInput {
	input := make([]*cato_models.GlobalIPRangeRefInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.GlobalIPRangeRefInput{
			By:    cato_models.ObjectRefByID,
			Input: globalIPRange.ID.ValueString(),
		}
		input = append(input, &inputItem)
	}
	return input
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	r *globalIPRangesResource,
	cfg *tf.GlobalIPRangesModel,
	state *tf.GlobalIPRangesModel,
	originalCfgRanges *[]tf.GlobalIPRange,
	originalStateRanges *[]tf.GlobalIPRange,
	diagnostics *diag.Diagnostics,
) (resultPlanDetails ipRangePlanDetails) {
	cfgRanges := *originalCfgRanges
	defer func() { *originalCfgRanges = cfgRanges }()

	stateRanges := *originalStateRanges
	defer func() { *originalStateRanges = stateRanges }()

	if utils.CheckErr(diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
		return resultPlanDetails
	}
	if utils.CheckErr(diagnostics, cfg.Ranges.ElementsAs(ctx, &cfgRanges, false)) {
		return resultPlanDetails
	}
	_, resultPlanDetails = r.computePlan(ctx, stateRanges, cfgRanges, diagnostics)
	return resultPlanDetails
}

// convertCreateValues converts between the Terraform model and SDK values.
func convertCreateValues(
	ctx context.Context,
	diags diag.Diagnostics,
	tfRange tf.GlobalIPRange,
) (resultRangeObj basetypes.ObjectValue) {
	resultRangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
	diags.Append(objDiags...)
	return resultRangeObj
}

// convertUpdateValues converts between the Terraform model and SDK values.
func convertUpdateValues(
	ctx context.Context,
	diags diag.Diagnostics,
	tfRange tf.GlobalIPRange,
) (resultRangeObj basetypes.ObjectValue) {
	resultRangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
	diags.Append(objDiags...)
	return resultRangeObj
}

// convertHydrateValues converts between the Terraform model and SDK values.
func convertHydrateValues(diags *diag.Diagnostics, globalIPRanges []attr.Value) (resultGlobalIPRangesSet basetypes.SetValue) {
	resultGlobalIPRangesSet, setDiags := types.SetValue(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, globalIPRanges)
	diags.Append(setDiags...)
	return resultGlobalIPRangesSet
}

// convertHydrateValues2 converts between the Terraform model and SDK values.
func convertHydrateValues2(
	ctx context.Context,
	diags *diag.Diagnostics,
	item *cato_go_sdk.ObjectGlobalIpRangeList_Object_GlobalIPRangeList_Items,
) (resultRangeObj basetypes.ObjectValue) {
	tfRange := tf.GlobalIPRange{
		Description: types.StringPointerValue(item.GetDescription()),
		ID:          types.StringValue(item.GetID()),
		IPRange:     types.StringValue(item.GetIPRange()),
		Name:        types.StringValue(item.GetName()),
	}
	resultRangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
	diags.Append(objDiags...)
	return resultRangeObj
}
