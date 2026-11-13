package tssection

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	terminalserver "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &terminalServerSectionResource{}
	_ resource.ResourceWithConfigure   = &terminalServerSectionResource{}
	_ resource.ResourceWithImportState = &terminalServerSectionResource{}
)

func NewTerminalServerSectionResource() resource.Resource {
	return &terminalServerSectionResource{}
}

type terminalServerSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *terminalServerSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_terminal_server_section"
}

func (r *terminalServerSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *terminalServerSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *terminalServerSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan TerminalServerSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyTerminalServerAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicyTerminalServerAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyTerminalServerAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetTerminalServer().GetAddSection()
	if err := sharedpolicy.MutationResult(
		add.GetStatus(),
		add.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(terminalserver.PublishTerminalServerPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *terminalServerSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state TerminalServerSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyTerminalServerPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyTerminalServerPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetTerminalServer().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "terminal server policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "terminal server section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *terminalServerSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan TerminalServerSection
	var state TerminalServerSection
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
		moveSection, err := r.resourceAPI().PolicyTerminalServerMoveSection(ctx, inputMove, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Cato API PolicyTerminalServerMoveSection error", err.Error())
			return
		}
		if err := sharedpolicy.MutationResult(
			moveSection.GetPolicy().GetTerminalServer().GetMoveSection().GetStatus(),
			moveSection.GetPolicy().GetTerminalServer().GetMoveSection().GetErrors(),
		); err != nil {
			resp.Diagnostics.AddError("Policy mutation failed", err.Error())
			return
		}
	}

	updateSection, err := r.resourceAPI().PolicyTerminalServerUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyTerminalServerUpdateSection error", err.Error())
		return
	}
	if err := sharedpolicy.MutationResult(
		updateSection.GetPolicy().GetTerminalServer().GetUpdateSection().GetStatus(),
		updateSection.GetPolicy().GetTerminalServer().GetUpdateSection().GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(terminalserver.PublishTerminalServerPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *terminalServerSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state TerminalServerSection
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
	removed, err := r.resourceAPI().PolicyTerminalServerRemoveSection(ctx, remove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyTerminalServerRemoveSection error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetTerminalServer().GetRemoveSection()
	if err := sharedpolicy.MutationResult(
		payload.GetStatus(),
		payload.GetErrors(),
	); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(terminalserver.PublishTerminalServerPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
