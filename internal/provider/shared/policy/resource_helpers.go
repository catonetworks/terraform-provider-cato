package policy

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// KnownValue describes one Terraform input that must be resolved before mutation.
type KnownValue struct {
	Name     string
	Value    attr.Value
	Required bool
}

// RevisionState is the subset of Terraform private state used for resource-owned drafts.
type RevisionState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

// PositionModel contains the common policy-relative placement fields.
type PositionModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

var positionAttrTypes = map[string]attr.Type{"position": types.StringType, "ref": types.StringType}

// SectionModel is the common id/name model for policy sections.
type SectionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// ValidID is shared by the independent policy resource implementations.
func ValidID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		diags.AddError("Invalid resource ID", "A known, non-empty resource ID is required in state.")
		return false
	}
	return true
}

// CheckStatus is shared by the independent policy resource implementations.
func CheckStatus(status *cato_models.PolicyMutationStatus, diags *diag.Diagnostics) bool {
	if status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		diags.AddError("Policy mutation failed", "API did not return SUCCESS.")
		return false
	}
	return true
}

// RevisionMissing is shared by the independent policy resource implementations.
func RevisionMissing(err error) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.GqlErrors == nil {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item.Extensions["errorCode"] == "PolicyRevisionNotFound" || item.Message == "Revision was not found" {
			return true
		}
	}
	return false
}

// RememberRevision is shared by the independent policy resource implementations.
func RememberRevision(ctx context.Context, store RevisionState, revision string, diags *diag.Diagnostics) {
	var data []byte
	if revision != "" {
		data = []byte(strconv.Quote(revision))
	}
	diags.Append(store.SetKey(ctx, "pending_revision", data)...)
}

// PendingRevision is shared by the independent policy resource implementations.
func PendingRevision(ctx context.Context, store RevisionState, diags *diag.Diagnostics) string {
	data, d := store.GetKey(ctx, "pending_revision")
	diags.Append(d...)
	if len(data) == 0 || diags.HasError() {
		return ""
	}
	revision, err := strconv.Unquote(string(data))
	if err != nil {
		diags.AddError("Invalid private state", "Cannot decode the pending policy revision ID.")
	}
	return revision
}

// ValidateRulePosition is shared by the independent policy resource implementations.
func ValidateRulePosition(at *cato_models.PolicyRulePositionInput, diags *diag.Diagnostics) {
	if at == nil {
		return
	}
	if at.Position == nil {
		diags.AddError("Missing position", "A position is required when at is configured.")
		return
	}
	position := *at.Position
	if !slices.Contains(cato_models.AllPolicyRulePositionEnum, position) {
		diags.AddError("Invalid position", fmt.Sprintf("Unsupported position %q.", position))
		return
	}
	if position != cato_models.PolicyRulePositionEnumLastInPolicy && position != cato_models.PolicyRulePositionEnumFirstInPolicy {
		if at.Ref == nil || *at.Ref == "" {
			diags.AddError("Missing position reference", "Relative placement requires a known, non-empty ref.")
		}
	}
}

// ValidateSectionPosition is shared by the independent policy resource implementations.
func ValidateSectionPosition(at *cato_models.PolicySectionPositionInput, diags *diag.Diagnostics) {
	if at == nil {
		diags.AddError("Missing position", "Section placement is required.")
		return
	}
	position := at.Position
	if !slices.Contains(cato_models.AllPolicySectionPositionEnum, position) {
		diags.AddError("Invalid position", fmt.Sprintf("Unsupported position %q.", position))
		return
	}
	if position != cato_models.PolicySectionPositionEnumLastInPolicy {
		if at.Ref == nil || *at.Ref == "" {
			diags.AddError("Missing position reference", "Relative placement requires a known, non-empty ref.")
		}
	}
}

// BuildRulePosition is shared by the independent policy resource implementations.
func BuildRulePosition(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRulePositionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "at must be known before applying.")
		return nil
	}
	var model PositionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	ValidateKnownValues(diags,
		KnownValue{Name: "at.position", Value: model.Position, Required: true},
		KnownValue{Name: "at.ref", Value: model.Ref, Required: false},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyRulePositionInput{}
	vPosition := cato_models.PolicyRulePositionEnum(model.Position.ValueString())
	input.Position = &vPosition
	input.Ref = convert.StringPointerForOptionalInput(model.Ref)
	return input
}

// BuildSectionPosition is shared by the independent policy resource implementations.
func BuildSectionPosition(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicySectionPositionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "at must be known before applying.")
		return nil
	}
	var model PositionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	ValidateKnownValues(diags,
		KnownValue{Name: "at.position", Value: model.Position, Required: true},
		KnownValue{Name: "at.ref", Value: model.Ref, Required: false},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicySectionPositionInput{}
	input.Position = cato_models.PolicySectionPositionEnum(model.Position.ValueString())
	input.Ref = convert.StringPointerForOptionalInput(model.Ref)
	return input
}

// BuildSectionCreate is shared by the independent policy resource implementations.
func BuildSectionCreate(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyAddSectionInfoInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "section must be known before applying.")
		return nil
	}
	var model SectionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	ValidateKnownValues(diags,
		KnownValue{Name: "section.name", Value: model.Name, Required: true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyAddSectionInfoInput{}
	input.Name = model.Name.ValueString()
	return input
}

// BuildSectionUpdate is shared by the independent policy resource implementations.
func BuildSectionUpdate(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyUpdateSectionInfoInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "section must be known before applying.")
		return nil
	}
	var model SectionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	ValidateKnownValues(diags,
		KnownValue{Name: "section.name", Value: model.Name, Required: true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyUpdateSectionInfoInput{}
	vName := model.Name.ValueString()
	input.Name = &vName
	return input
}

// SectionAttributes is shared by the independent policy resource implementations.
func SectionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:   "Section ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

// RuleSectionAttributes is shared by the independent policy resource implementations.
func RuleSectionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:   "Section ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Description: "Section name",
			Computed:    true,
		},
	}
}

// SectionPositionAttributes is shared by the independent policy resource implementations.
func SectionPositionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"position": schema.StringAttribute{
			Description: "Position",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicySectionPositionEnum)...)},
		},
		"ref": schema.StringAttribute{
			Description:   "Ref",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
	}
}

// SectionResourceAttributes is shared by the independent policy resource implementations.
func SectionResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "At",
			Required:    true,
			Attributes:  SectionPositionAttributes(),
		},
		"id": schema.StringAttribute{
			Description:   "Section ID; mirrors section.id",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"section": schema.SingleNestedAttribute{
			Description: "Section",
			Required:    true,
			Attributes:  SectionAttributes(),
		},
	}
}

func validateValues(diags *diag.Diagnostics, checkSets bool, values ...KnownValue) {
	for _, value := range values {
		if elements, ok := value.Value.(types.Set); ok && checkSets {
			for _, element := range elements.Elements() {
				if element.IsNull() || element.IsUnknown() {
					diags.AddError("Invalid configuration", value.Name+" must not contain null or unknown elements.")
					break
				}
			}
		}
		if value.Value.IsUnknown() || (value.Required && value.Value.IsNull()) {
			diags.AddError("Invalid configuration", value.Name+" must contain a known value.")
		}
	}
}

// ValidateKnownValues checks required and unknown scalar/object values.
func ValidateKnownValues(d *diag.Diagnostics, values ...KnownValue) {
	validateValues(d, false, values...)
}

// ValidateRuleValues also rejects null and unknown elements in criteria sets.
func ValidateRuleValues(d *diag.Diagnostics, values ...KnownValue) {
	validateValues(d, true, values...)
}

// RulePositionAttributes preserves the resource's placement descriptions.
func RulePositionAttributes(positionDescription, refDescription string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"position": schema.StringAttribute{Description: positionDescription,
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRulePositionEnum)...)}},

		"ref": schema.StringAttribute{Description: refDescription,
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()}},
	}
}
