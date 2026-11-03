package sbsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	socketbypass "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &socketBypassSectionResource{}
	_ resource.ResourceWithConfigure   = &socketBypassSectionResource{}
	_ resource.ResourceWithImportState = &socketBypassSectionResource{}
)

func NewSocketBypassSectionResource() resource.Resource {
	return &socketBypassSectionResource{}
}

type socketBypassSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *socketBypassSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_bypass_section"
}

func (r *socketBypassSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketBypassSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *socketBypassSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SocketBypassSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySocketBypassAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicySocketBypassAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySocketBypassAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetSocketBypass().GetAddSection()
	for _, e := range add.GetErrors() {
		if e != nil && e.ErrorCode != nil {
			resp.Diagnostics.AddError("API error: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}
	if add.GetStatus() == nil || *add.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		st := ""
		if add.GetStatus() != nil {
			st = string(*add.GetStatus())
		}
		resp.Diagnostics.AddError("socket bypass addSection failed", st)
		return
	}

	resp.Diagnostics.Append(socketbypass.PublishSocketBypassPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *socketBypassSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SocketBypassSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySocketBypassPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySocketBypassPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSocketBypass().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "socket bypass policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "socket bypass section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *socketBypassSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SocketBypassSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicySocketBypassMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySocketBypassMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetSocketBypass().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetSocketBypass().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetSocketBypass().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.resourceAPI().PolicySocketBypassUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySocketBypassUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetSocketBypass().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetSocketBypass().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetSocketBypass().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(socketbypass.PublishSocketBypassPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *socketBypassSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SocketBypassSection
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
	if _, err := r.resourceAPI().PolicySocketBypassRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicySocketBypassRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(socketbypass.PublishSocketBypassPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
