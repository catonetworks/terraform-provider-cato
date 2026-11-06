package stsection

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	splittunnel "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &splitTunnelSectionResource{}
	_ resource.ResourceWithConfigure   = &splitTunnelSectionResource{}
	_ resource.ResourceWithImportState = &splitTunnelSectionResource{}
)

func NewSplitTunnelSectionResource() resource.Resource {
	return &splitTunnelSectionResource{}
}

type splitTunnelSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *splitTunnelSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_split_tunnel_section"
}

func (r *splitTunnelSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *splitTunnelSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *splitTunnelSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SplitTunnelSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySplitTunnelAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicySplitTunnelAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetSplitTunnel().GetAddSection()
	if err := sharedpolicy.MutationResult(
		add.GetStatus(),
		add.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *splitTunnelSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SplitTunnelSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySplitTunnelPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSplitTunnel().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "split tunnel policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "split tunnel section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *splitTunnelSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SplitTunnelSection
	var state SplitTunnelSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.At.Equal(state.At) {
		moveSection, err := r.resourceAPI().PolicySplitTunnelMoveSection(ctx, inputMove, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Cato API PolicySplitTunnelMoveSection error", err.Error())
			return
		}
		if err := sharedpolicy.MutationResult(
			moveSection.GetPolicy().GetSplitTunnel().GetMoveSection().GetStatus(),
			moveSection.GetPolicy().GetSplitTunnel().GetMoveSection().GetErrors(),
		); err != nil {
			resp.Diagnostics.AddError("Policy mutation failed", err.Error())
			return
		}
	}

	updateSection, err := r.resourceAPI().PolicySplitTunnelUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelUpdateSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		updateSection.GetPolicy().GetSplitTunnel().GetUpdateSection().GetStatus(),
		updateSection.GetPolicy().GetSplitTunnel().GetUpdateSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *splitTunnelSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SplitTunnelSection
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
	removed, err := r.resourceAPI().PolicySplitTunnelRemoveSection(ctx, remove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelRemoveSection error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetSplitTunnel().GetRemoveSection()
	if err := sharedpolicy.MutationResult(
		payload.GetStatus(),
		payload.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
