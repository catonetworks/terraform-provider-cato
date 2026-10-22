package lansection

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
	_ resource.Resource                = &socketLanSectionResource{}
	_ resource.ResourceWithConfigure   = &socketLanSectionResource{}
	_ resource.ResourceWithImportState = &socketLanSectionResource{}
)

func NewSocketLanSectionResource() resource.Resource {
	return &socketLanSectionResource{}
}

type socketLanSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *socketLanSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_section"
}

func (r *socketLanSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketLanSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *socketLanSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SocketLanSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.PolicySocketLanAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.resourceAPI().PolicySocketLanAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanAddSection error",
			err.Error(),
		)
		return
	}
	if len(policyChange.Policy.SocketLan.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// publishing new section
	tflog.Info(ctx, "Create.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanPublishPolicyRevision.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	convertCreateInput3(&plan, policyChange)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overriding state with section id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetSocketLan().GetAddSection().Section.GetSection().ID,
	)
}

func (r *socketLanSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	tflog.Debug(ctx, "Read.PolicySocketLanPolicy.response", map[string]interface{}{
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

	sectionExist := convertReadState2(&state, body, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "socket lan section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *socketLanSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SocketLanSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection, inputMoveSection := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)

	// move section
	tflog.Debug(ctx, "Update.PolicySocketLanMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.resourceAPI().PolicySocketLanMoveSection(ctx, inputMoveSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanMoveSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.SocketLan.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.SocketLan.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// update section
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.resourceAPI().PolicySocketLanUpdateSection(ctx, nil, inputUpdateSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanUpdateSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.SocketLan.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.SocketLan.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing section changes
	tflog.Info(ctx, "Update.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
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

func (r *socketLanSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketLanSection
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

	PolicySocketLanRemoveSectionResponse, err := r.resourceAPI().PolicySocketLanRemoveSection(ctx, nil, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(PolicySocketLanRemoveSectionResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Delete.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
