package atsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	tenant "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &appTenantRestrictionSectionResource{}
	_ resource.ResourceWithConfigure   = &appTenantRestrictionSectionResource{}
	_ resource.ResourceWithImportState = &appTenantRestrictionSectionResource{}
)

func NewAppTenantRestrictionSectionResource() resource.Resource {
	return &appTenantRestrictionSectionResource{}
}

type appTenantRestrictionSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *appTenantRestrictionSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_app_tenant_restriction_section"
}

func (r *appTenantRestrictionSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *appTenantRestrictionSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *appTenantRestrictionSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan AppTenantRestrictionSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyAppTenantRestrictionAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicyAppTenantRestrictionAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetAppTenantRestriction().GetAddSection()
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
		resp.Diagnostics.AddError("app tenant restriction addSection failed", st)
		return
	}

	resp.Diagnostics.Append(tenant.PublishAppTenantRestrictionPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *appTenantRestrictionSectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state AppTenantRestrictionSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "app tenant restriction policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "app tenant restriction section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *appTenantRestrictionSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan AppTenantRestrictionSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicyAppTenantRestrictionMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetAppTenantRestriction().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.resourceAPI().PolicyAppTenantRestrictionUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetAppTenantRestriction().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(tenant.PublishAppTenantRestrictionPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appTenantRestrictionSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state AppTenantRestrictionSection
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
	if _, err := r.resourceAPI().PolicyAppTenantRestrictionRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(tenant.PublishAppTenantRestrictionPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
