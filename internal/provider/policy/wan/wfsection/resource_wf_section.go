package wfsection

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
	_ resource.Resource                = &wanFwSectionResource{}
	_ resource.ResourceWithConfigure   = &wanFwSectionResource{}
	_ resource.ResourceWithImportState = &wanFwSectionResource{}
)

func NewWanFwSectionResource() resource.Resource {
	return &wanFwSectionResource{}
}

type wanFwSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *wanFwSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_section"
}

func (r *wanFwSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *wanFwSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *wanFwSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.PolicyWanFirewallAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.resourceAPI().PolicyWanFirewallAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicyWanFirewallAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})

	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallAddSection error", err.Error())
		return
	}
	if len(policyChange.Policy.WanFirewall.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.WanFirewall.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// publishing new section
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	populateWanSectionID(&plan, policyChange)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with rule id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetWanFirewall().GetAddSection().Section.GetSection().ID,
	)
}

func (r *wanFwSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanFirewallSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.resourceAPI().PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanFirewall.response", map[string]interface{}{
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
	section := convertReadState2(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionExist, diagstmp := convertReadState3(ctx, &state, body, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "wan section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanFwSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection, inputMoveSection := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Update.PolicyWanFirewallMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.resourceAPI().PolicyWanFirewallMoveSection(ctx, inputMoveSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyWanFirewallMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.WanFirewall.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.WanFirewall.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Debug(ctx, "Update.PolicyWanFirewallUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.resourceAPI().PolicyWanFirewallUpdateSection(ctx, inputUpdateSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyWanFirewallUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.WanFirewall.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.WanFirewall.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing new section
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
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

func (r *wanFwSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanFirewallSection
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

	tflog.Debug(ctx, "Delete.PolicyWanFirewallRemoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeSection),
	})
	policyWanFirewallRemoveSectionResponse, err := r.resourceAPI().PolicyWanFirewallRemoveSection(ctx, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyWanFirewallRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyWanFirewallRemoveSectionResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
