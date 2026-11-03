package ccsection

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	clientconnectivity "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &clientConnectivitySectionResource{}
	_ resource.ResourceWithConfigure   = &clientConnectivitySectionResource{}
	_ resource.ResourceWithImportState = &clientConnectivitySectionResource{}
)

func NewClientConnectivitySectionResource() resource.Resource {
	return &clientConnectivitySectionResource{}
}

type clientConnectivitySectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *clientConnectivitySectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_client_connectivity_section"
}

func (r *clientConnectivitySectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *clientConnectivitySectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *clientConnectivitySectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ClientConnectivitySection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyClientConnectivityAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicyClientConnectivityAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyClientConnectivityAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetClientConnectivity().GetAddSection()
	if err := sharedpolicy.MutationResult(
		add.GetStatus(),
		add.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(clientconnectivity.PublishClientConnectivityPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *clientConnectivitySectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ClientConnectivitySection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyClientConnectivityPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyClientConnectivityPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetClientConnectivity().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "client connectivity policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "client connectivity section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientConnectivitySectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ClientConnectivitySection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicyClientConnectivityMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyClientConnectivityMoveSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		moveSection.GetPolicy().GetClientConnectivity().GetMoveSection().GetStatus(),
		moveSection.GetPolicy().GetClientConnectivity().GetMoveSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	updateSection, err := r.resourceAPI().PolicyClientConnectivityUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyClientConnectivityUpdateSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		updateSection.GetPolicy().GetClientConnectivity().GetUpdateSection().GetStatus(),
		updateSection.GetPolicy().GetClientConnectivity().GetUpdateSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(clientconnectivity.PublishClientConnectivityPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *clientConnectivitySectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ClientConnectivitySection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	remove := convertDeleteInput(section)
	removed, err := r.resourceAPI().PolicyClientConnectivityRemoveSection(ctx, remove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyClientConnectivityRemoveSection error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetClientConnectivity().GetRemoveSection()
	if err := sharedpolicy.MutationResult(
		payload.GetStatus(),
		payload.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(clientconnectivity.PublishClientConnectivityPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
