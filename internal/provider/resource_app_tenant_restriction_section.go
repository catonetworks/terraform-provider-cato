//nolint:lll
package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &appTenantRestrictionSectionResource{}
	_ resource.ResourceWithConfigure   = &appTenantRestrictionSectionResource{}
	_ resource.ResourceWithImportState = &appTenantRestrictionSectionResource{}
)

func NewAppTenantRestrictionSectionResource() resource.Resource {
	return &appTenantRestrictionSectionResource{}
}

type appTenantRestrictionSectionResource struct {
	client *catoClientData
}

func (r *appTenantRestrictionSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_tenant_restriction_section"
}

func (r *appTenantRestrictionSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom section in the Cato app tenant restriction (App & Data Inline Protection) policy. " +
			"Adding a section can reassign existing rules into that section and create a draft revision; see provider documentation. " +
			"Underlying GraphQL is marked @beta.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the section",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy or another section",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_SECTION", "BEFORE_SECTION", "LAST_IN_POLICY"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference section ID when position is AFTER_SECTION or BEFORE_SECTION",
						Optional:    true,
					},
				},
			},
			"section": schema.SingleNestedAttribute{
				Description: "Section name and computed ID",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Section ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Section name",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *appTenantRestrictionSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *appTenantRestrictionSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *appTenantRestrictionSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppTenantRestrictionSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PolicyAddSectionInput{}
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicySectionPositionInput{}
		positionInput := PolicyRulePositionInput{}
		resp.Diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		input.At.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}
	if !plan.Section.IsNull() {
		input.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := PolicyAddSectionInfoInput{}
		resp.Diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		input.Section.Name = sectionInput.Name.ValueString()
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyAppTenantRestrictionAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.client.catov2.PolicyAppTenantRestrictionAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetAppTenantRestriction().GetAddSection()
	for _, e := range add.GetErrors() {
		if e != nil && e.ErrorCode != nil {
			resp.Diagnostics.AddError("API error: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}
	if add.GetStatus() == nil || *add.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		st := ""
		if add.GetStatus() != nil {
			st = string(*add.GetStatus())
		}
		resp.Diagnostics.AddError("app tenant restriction addSection failed", st)
		return
	}

	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := add.GetSection().GetSection().GetID()
	plan.ID = types.StringValue(sid)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *appTenantRestrictionSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppTenantRestrictionSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
		return
	}

	section := PolicyUpdateSectionInfoInput{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "app tenant restriction policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := false
	for _, item := range pol.GetSections() {
		if item == nil || item.GetSection() == nil {
			continue
		}
		if item.GetSection().GetID() != section.ID.ValueString() {
			continue
		}
		found = true
		state.ID = types.StringValue(item.GetSection().GetID())
		cur, d := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(item.GetSection().GetID()),
			"name": types.StringValue(item.GetSection().GetName()),
		})
		resp.Diagnostics.Append(d...)
		state.Section = cur
		break
	}

	// Preserve the at block from existing state; the API does not return positioning
	// info on read. Only default to LAST_IN_POLICY when state has nothing (e.g. import).
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		resp.Diagnostics.Append(d...)
		state.At = atObj
	}

	if !found {
		tflog.Warn(ctx, "app tenant restriction section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *appTenantRestrictionSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppTenantRestrictionSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate := cato_models.PolicyUpdateSectionInput{}
	inputMove := cato_models.PolicyMoveSectionInput{}
	if !plan.Section.IsNull() {
		sectionInput := PolicyUpdateSectionInfoInput{}
		resp.Diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		inputUpdate.ID = sectionInput.ID.ValueString()
		inputUpdate.Section = &cato_models.PolicyUpdateSectionInfoInput{
			Name: sectionInput.Name.ValueStringPointer(),
		}
	}
	if !plan.At.IsNull() {
		positionInput := PolicyRulePositionInput{}
		resp.Diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		inputMove.To = &cato_models.PolicySectionPositionInput{
			Position: cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		}
		inputMove.ID = inputUpdate.ID
	}
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.client.catov2.PolicyAppTenantRestrictionMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.client.catov2.PolicyAppTenantRestrictionUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appTenantRestrictionSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppTenantRestrictionSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	section := PolicyAddSectionInfoInput{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	remove := cato_models.PolicyRemoveSectionInput{ID: section.ID.ValueString()}
	if _, err := r.client.catov2.PolicyAppTenantRestrictionRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
