package alwaysonsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &alwaysOnSectionResource{}
	_ resource.ResourceWithConfigure   = &alwaysOnSectionResource{}
	_ resource.ResourceWithImportState = &alwaysOnSectionResource{}
)

type alwaysOnSectionResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewAlwaysOnSectionResource returns the cato_always_on_section resource.
func NewAlwaysOnSectionResource() resource.Resource { return &alwaysOnSectionResource{} }

func (r *alwaysOnSectionResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return r.configuredClient()
}

func (r *alwaysOnSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_always_on_section"
}

func (r *alwaysOnSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *alwaysOnSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *alwaysOnSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Section
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildSectionInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.checkConfigured(&resp.Diagnostics) {
		return
	}
	errMsg := fmt.Sprintf("failed to create Always On section %s", input.Section.Name)
	result, err := r.getClient().PolicyZtnaAlwaysOnAddSection(ctx, r.client.AccountId, nil, input)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetAddSection()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, &resp.Diagnostics) ||
		!checkMutationStatus(payload.GetStatus(), errMsg, &resp.Diagnostics) {
		return
	}
	section := payload.GetSection().GetSection()
	if section.GetID() == "" {
		resp.Diagnostics.AddError(errMsg, "API did not return resource ID")
		return
	}
	plan.Section = parseSection(ctx, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alwaysOnSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Section
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := sectionID(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.checkConfigured(&resp.Diagnostics) {
		return
	}
	errMsg := fmt.Sprintf("failed to read Always On section %s", id)
	result, err := r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	section, position, err := findSection(result, id)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	if section == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Section = parseSection(ctx, section, &resp.Diagnostics)
	if !utils.HasValue(state.At) {
		state.At = parseSectionPosition(ctx, position, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alwaysOnSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Section
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := sectionID(ctx, state, &resp.Diagnostics)
	input := buildSectionInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.checkConfigured(&resp.Diagnostics) {
		return
	}
	plan.Section = r.updateSection(ctx, cato_models.PolicyUpdateSectionInput{
		ID: id, Section: &cato_models.PolicyUpdateSectionInfoInput{Name: &input.Section.Name},
	}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.At.Equal(state.At) {
		// Retain the completed rename and old placement if the subsequent move fails.
		resp.Diagnostics.Append(resp.State.Set(ctx, Section{At: state.At, Section: plan.Section})...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Section = r.moveSection(ctx, cato_models.PolicyMoveSectionInput{ID: id, To: input.At}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alwaysOnSectionResource) updateSection(ctx context.Context, input cato_models.PolicyUpdateSectionInput,
	diags *diag.Diagnostics,
) types.Object {
	errMsg := fmt.Sprintf("failed to update Always On section %s", input.ID)
	result, err := r.getClient().PolicyZtnaAlwaysOnUpdateSection(ctx, r.client.AccountId, nil, input)
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return types.ObjectNull(SectionDataAttrTypes)
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetUpdateSection()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, diags) ||
		!checkMutationStatus(payload.GetStatus(), errMsg, diags) {
		return types.ObjectNull(SectionDataAttrTypes)
	}
	section := payload.GetSection().GetSection()
	if section.GetID() != input.ID {
		diags.AddError(errMsg, fmt.Sprintf("API returned resource ID %q; expected %q", section.GetID(), input.ID))
		return types.ObjectNull(SectionDataAttrTypes)
	}
	if input.Section.Name != nil && section.GetName() != *input.Section.Name {
		// updateSection can return the section's previous name after a successful rename.
		current, readErr := readSectionByID(ctx, r.getClient(), r.client.AccountId, input.ID)
		if readErr != nil {
			diags.AddError(errMsg, readErr.Error())
			return types.ObjectNull(SectionDataAttrTypes)
		}
		if current.GetName() != *input.Section.Name {
			diags.AddError(errMsg, "API reported success but the section name did not match the requested name on read.")
			return types.ObjectNull(SectionDataAttrTypes)
		}
		return parseSection(ctx, current, diags)
	}
	return parseSection(ctx, section, diags)
}

func (r *alwaysOnSectionResource) moveSection(ctx context.Context, input cato_models.PolicyMoveSectionInput,
	diags *diag.Diagnostics,
) types.Object {
	errMsg := fmt.Sprintf("failed to move Always On section %s", input.ID)
	result, err := r.getClient().PolicyZtnaAlwaysOnMoveSection(ctx, r.client.AccountId, nil, input)
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return types.ObjectNull(SectionDataAttrTypes)
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetMoveSection()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, diags) ||
		!checkMutationStatus(payload.GetStatus(), errMsg, diags) {
		return types.ObjectNull(SectionDataAttrTypes)
	}
	section := payload.GetSection().GetSection()
	if section.GetID() != input.ID {
		diags.AddError(errMsg, fmt.Sprintf("API returned resource ID %q; expected %q", section.GetID(), input.ID))
		return types.ObjectNull(SectionDataAttrTypes)
	}
	return parseSection(ctx, section, diags)
}

func (r *alwaysOnSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Section
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := sectionID(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.checkConfigured(&resp.Diagnostics) {
		return
	}
	errMsg := fmt.Sprintf("failed to delete Always On section %s", id)
	result, err := r.getClient().PolicyZtnaAlwaysOnRemoveSection(ctx, r.client.AccountId, nil,
		cato_models.PolicyRemoveSectionInput{ID: id})
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetRemoveSection()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, &resp.Diagnostics) ||
		!checkMutationStatus(payload.GetStatus(), errMsg, &resp.Diagnostics) {
		return
	}
	resp.State.RemoveResource(ctx)
}

func checkMutationStatus(status *cato_models.PolicyMutationStatus, summary string, diags *diag.Diagnostics) bool {
	if status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		value := ""
		if status != nil {
			value = string(*status)
		}
		diags.AddError(summary, fmt.Sprintf("API returned mutation status %q", value))
		return false
	}
	return true
}

func (r *alwaysOnSectionResource) checkConfigured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.getClient() == nil {
		diags.AddError("Resource not configured", "The Cato client must be configured before managing an Always On section.")
		return false
	}
	return true
}
