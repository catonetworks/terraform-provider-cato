package rpfsection

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	remoteportfwd "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &remotePortFwdSectionResource{}
	_ resource.ResourceWithConfigure   = &remotePortFwdSectionResource{}
	_ resource.ResourceWithImportState = &remotePortFwdSectionResource{}
)

func NewRemotePortFwdSectionResource() resource.Resource {
	return &remotePortFwdSectionResource{}
}

type remotePortFwdSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *remotePortFwdSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_remote_port_fwd_section"
}

func (r *remotePortFwdSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *remotePortFwdSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *remotePortFwdSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan RemotePortFwdSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyRemotePortFwdAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicyRemotePortFwdAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyRemotePortFwdAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetRemotePortFwd().GetAddSection()
	if err := sharedpolicy.MutationResult(
		add.GetStatus(),
		add.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(remoteportfwd.PublishRemotePortFwdPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *remotePortFwdSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state RemotePortFwdSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().RemotePortFwdPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API RemotePortFwdPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetRemotePortFwd().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "remote port forwarding policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "remote port forwarding section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *remotePortFwdSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan RemotePortFwdSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicyRemotePortFwdMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyRemotePortFwdMoveSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		moveSection.GetPolicy().GetRemotePortFwd().GetMoveSection().GetStatus(),
		moveSection.GetPolicy().GetRemotePortFwd().GetMoveSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	updateSection, err := r.resourceAPI().PolicyRemotePortFwdUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyRemotePortFwdUpdateSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		updateSection.GetPolicy().GetRemotePortFwd().GetUpdateSection().GetStatus(),
		updateSection.GetPolicy().GetRemotePortFwd().GetUpdateSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(remoteportfwd.PublishRemotePortFwdPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *remotePortFwdSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state RemotePortFwdSection
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
	removed, err := r.resourceAPI().PolicyRemotePortFwdRemoveSection(ctx, remove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyRemotePortFwdRemoveSection error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetRemotePortFwd().GetRemoveSection()
	if err := sharedpolicy.MutationResult(
		payload.GetStatus(),
		payload.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(remoteportfwd.PublishRemotePortFwdPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
