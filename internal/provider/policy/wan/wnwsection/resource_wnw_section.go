package wnwsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &wanNetworkSectionResource{}
	_ resource.ResourceWithConfigure   = &wanNetworkSectionResource{}
	_ resource.ResourceWithImportState = &wanNetworkSectionResource{}
)

func NewWanNetworkSectionResource() resource.Resource {
	return &wanNetworkSectionResource{}
}

type wanNetworkSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *wanNetworkSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_wnw_section"
}

func (r *wanNetworkSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *wanNetworkSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *wanNetworkSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanNetworkSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.PolicyWanNetworkAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.resourceAPI().PolicyWanNetworkAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicyWanNetworkAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})

	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanNetworkAddSection error", err.Error())
		return
	}
	if len(policyChange.Policy.WanNetwork.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.WanNetwork.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// publishing new section
	tflog.Info(ctx, "publishing new section")
	_, err = r.resourceAPI().PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	convertCreateInput2(&plan, policyChange)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overriding state with section id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetWanNetwork().GetAddSection().Section.GetSection().ID,
	)
}

func (r *wanNetworkSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanNetworkSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().WanNetworkPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanNetwork.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// retrieve section ID
	section := convertReadState(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionExist, diagstmp := convertReadState2(&state, body, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "wan network section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanNetworkSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanNetworkSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection, inputMoveSection := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Update.PolicyWanNetworkMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.resourceAPI().PolicyWanNetworkMoveSection(ctx, inputMoveSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyWanNetworkMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkMoveSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.WanNetwork.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.WanNetwork.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Debug(ctx, "Update.PolicyWanNetworkUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.resourceAPI().PolicyWanNetworkUpdateSection(ctx, inputUpdateSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyWanNetworkUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkUpdateSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.WanNetwork.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.WanNetwork.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing updated section
	tflog.Info(ctx, "publishing updated section")
	_, err = r.resourceAPI().PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanNetworkSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanNetworkSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve section ID
	section := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := convertDeleteInput2(section)

	tflog.Debug(ctx, "Delete.PolicyWanNetworkRemoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeSection),
	})
	policyWanNetworkRemoveSectionResponse, err := r.resourceAPI().PolicyWanNetworkRemoveSection(ctx, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyWanNetworkRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyWanNetworkRemoveSectionResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	_, err = r.resourceAPI().PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
