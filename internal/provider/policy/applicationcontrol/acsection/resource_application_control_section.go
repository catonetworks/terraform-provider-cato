package acsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	applicationcontrol "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &applicationControlSectionResource{}
	_ resource.ResourceWithConfigure   = &applicationControlSectionResource{}
	_ resource.ResourceWithImportState = &applicationControlSectionResource{}
)

func NewApplicationControlSectionResource() resource.Resource {
	return &applicationControlSectionResource{}
}

type applicationControlSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *applicationControlSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_application_control_section"
}

func (r *applicationControlSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *applicationControlSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *applicationControlSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ApplicationControlSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyApplicationControlAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicyApplicationControlAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetApplicationControl().GetAddSection()
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
		resp.Diagnostics.AddError("Application Control addSection failed", st)
		return
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *applicationControlSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationControlSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "application control policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "application control section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *applicationControlSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ApplicationControlSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicyApplicationControlMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.resourceAPI().PolicyApplicationControlUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationControlSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ApplicationControlSection
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
	if _, err := r.resourceAPI().PolicyApplicationControlRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
