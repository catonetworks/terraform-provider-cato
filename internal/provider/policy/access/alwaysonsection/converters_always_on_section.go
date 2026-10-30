package alwaysonsection

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func buildSectionInput(ctx context.Context, plan Section, diags *diag.Diagnostics) cato_models.PolicyAddSectionInput {
	var section SectionData
	diags.Append(plan.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if !utils.HasValue(section.Name) {
		diags.AddError("Invalid Always On section", "The section name must be known before applying the resource.")
	}
	return cato_models.PolicyAddSectionInput{
		At:      buildSectionPosition(ctx, plan.At, diags),
		Section: &cato_models.PolicyAddSectionInfoInput{Name: section.Name.ValueString()},
	}
}

func buildSectionPosition(ctx context.Context, at types.Object, diags *diag.Diagnostics) *cato_models.PolicySectionPositionInput {
	var position policyrule.Position
	diags.Append(at.As(ctx, &position, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	enum := cato_models.PolicySectionPositionEnum(position.Position.ValueString())
	if !utils.HasValue(position.Position) || !enum.IsValid() || position.Ref.IsUnknown() {
		diags.AddError("Invalid section position", "The section position and reference must be known and use a supported position.")
		return nil
	}
	input := &cato_models.PolicySectionPositionInput{Position: enum}
	if position.Ref.ValueString() != "" {
		input.Ref = position.Ref.ValueStringPointer()
	}
	if enum != cato_models.PolicySectionPositionEnumLastInPolicy && input.Ref == nil {
		diags.AddError("Invalid section reference", "A non-empty ref is required for AFTER_SECTION and BEFORE_SECTION.")
	}
	return input
}

func sectionID(ctx context.Context, model Section, diags *diag.Diagnostics) string {
	var section SectionData
	diags.Append(model.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if !utils.HasValue(section.ID) || section.ID.ValueString() == "" {
		diags.AddError("Invalid Always On section ID", "A known, non-empty section ID is required in state.")
	}
	return section.ID.ValueString()
}

type sectionInfo interface {
	GetID() string
	GetName() string
}

func parseSection(ctx context.Context, info sectionInfo, diags *diag.Diagnostics) types.Object {
	value, d := types.ObjectValueFrom(ctx, SectionDataAttrTypes, SectionData{
		ID: types.StringValue(info.GetID()), Name: types.StringValue(info.GetName()),
	})
	diags.Append(d...)
	return value
}

func parseSectionPosition(ctx context.Context, position *cato_models.PolicySectionPositionInput,
	diags *diag.Diagnostics,
) types.Object {
	ref := ""
	if position.Ref != nil {
		ref = *position.Ref
	}
	value, d := types.ObjectValueFrom(ctx, policyrule.PositionAttrTypes, policyrule.Position{
		Position: types.StringValue(string(position.Position)), Ref: types.StringValue(ref),
	})
	diags.Append(d...)
	return value
}
