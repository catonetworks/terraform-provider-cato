package ssiasection

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	sdpstaticip "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &sdpStaticIPAllocationSectionResource{}
	_ resource.ResourceWithConfigure   = &sdpStaticIPAllocationSectionResource{}
	_ resource.ResourceWithImportState = &sdpStaticIPAllocationSectionResource{}
)

func NewSdpStaticIPAllocationSectionResource() resource.Resource {
	return &sdpStaticIPAllocationSectionResource{}
}

type sdpStaticIPAllocationSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *sdpStaticIPAllocationSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sdp_static_ip_allocation_section"
}

func (r *sdpStaticIPAllocationSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *sdpStaticIPAllocationSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *sdpStaticIPAllocationSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SdpStaticIPAllocationSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySdpStaticIPAllocationAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicySdpStaticIPAllocationAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySdpStaticIPAllocationAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetSdpStaticIPAllocation().GetAddSection()
	if err := sharedpolicy.MutationResult(
		add.GetStatus(),
		add.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(sdpstaticip.PublishSdpStaticIPAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *sdpStaticIPAllocationSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SdpStaticIPAllocationSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySdpStaticIPAllocationPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySdpStaticIPAllocationPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSdpStaticIPAllocation().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "SDP static IP allocation policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "SDP static IP allocation section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sdpStaticIPAllocationSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SdpStaticIPAllocationSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicySdpStaticIPAllocationMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySdpStaticIPAllocationMoveSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		moveSection.GetPolicy().GetSdpStaticIPAllocation().GetMoveSection().GetStatus(),
		moveSection.GetPolicy().GetSdpStaticIPAllocation().GetMoveSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	updateSection, err := r.resourceAPI().PolicySdpStaticIPAllocationUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySdpStaticIPAllocationUpdateSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		updateSection.GetPolicy().GetSdpStaticIPAllocation().GetUpdateSection().GetStatus(),
		updateSection.GetPolicy().GetSdpStaticIPAllocation().GetUpdateSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(sdpstaticip.PublishSdpStaticIPAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *sdpStaticIPAllocationSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SdpStaticIPAllocationSection
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
	removed, err := r.resourceAPI().PolicySdpStaticIPAllocationRemoveSection(ctx, remove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySdpStaticIPAllocationRemoveSection error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetSdpStaticIPAllocation().GetRemoveSection()
	if err := sharedpolicy.MutationResult(
		payload.GetStatus(),
		payload.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(sdpstaticip.PublishSdpStaticIPAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
