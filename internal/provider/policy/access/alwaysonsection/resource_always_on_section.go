package alwaysonsection

import (
	"context"
	"fmt"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &alwaysOnSectionResource{}
	_ resource.ResourceWithConfigure   = &alwaysOnSectionResource{}
	_ resource.ResourceWithImportState = &alwaysOnSectionResource{}
)

func NewAlwaysOnSectionResource() resource.Resource {
	return &alwaysOnSectionResource{}
}

type alwaysOnSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *alwaysOnSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_always_on_section"
}

func (r *alwaysOnSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *alwaysOnSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *alwaysOnSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlwaysOnSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.PolicyZtnaAlwaysOnAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.resourceAPI().PolicyZtnaAlwaysOnAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicyZtnaAlwaysOnAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyZtnaAlwaysOnAddSection error",
			err.Error(),
		)
		return
	}
	if len(policyChange.Policy.ZtnaAlwaysOn.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.ZtnaAlwaysOn.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// Changes stay in the draft policy - cato_bulk_ao_move_rule publishes.

	populateAlwaysOnSectionID(&plan, policyChange)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetZtnaAlwaysOn().GetAddSection().Section.GetSection().ID,
	)
}

func (r *alwaysOnSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlwaysOnSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyZtnaAlwaysOnPolicy error",
			err.Error(),
		)
		return
	}

	section := convertReadState(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionExist := convertReadState2(&state, body, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !sectionExist {
		tflog.Warn(ctx, "always-on section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *alwaysOnSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlwaysOnSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection, inputMoveSection := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicyZtnaAlwaysOnMoveSection(ctx, inputMoveSection, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyZtnaAlwaysOnMoveSection error",
			err.Error(),
		)
		return
	}
	if moveSection.Policy.ZtnaAlwaysOn.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.ZtnaAlwaysOn.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	updateSection, err := r.resourceAPI().PolicyZtnaAlwaysOnUpdateSection(ctx, inputUpdateSection, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyZtnaAlwaysOnUpdateSection error",
			err.Error(),
		)
		return
	}
	if updateSection.Policy.ZtnaAlwaysOn.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.ZtnaAlwaysOn.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Changes stay in the draft policy - cato_bulk_ao_move_rule publishes.

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *alwaysOnSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlwaysOnSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	section := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := convertDeleteInput2(section)

	removeResp, err := r.resourceAPI().PolicyZtnaAlwaysOnRemoveSection(ctx, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyZtnaAlwaysOnRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResp),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyZtnaAlwaysOnRemoveSection error",
			err.Error(),
		)
		return
	}

	// Changes stay in the draft policy - cato_bulk_ao_move_rule publishes.
}
