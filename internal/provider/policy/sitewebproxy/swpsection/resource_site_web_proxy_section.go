package swpsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	sitewebproxy "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &siteWebProxySectionResource{}
	_ resource.ResourceWithConfigure   = &siteWebProxySectionResource{}
	_ resource.ResourceWithImportState = &siteWebProxySectionResource{}
)

func NewSiteWebProxySectionResource() resource.Resource {
	return &siteWebProxySectionResource{}
}

type siteWebProxySectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *siteWebProxySectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_site_web_proxy_section"
}

func (r *siteWebProxySectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *siteWebProxySectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *siteWebProxySectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SiteWebProxySection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySiteWebProxyAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.resourceAPI().PolicySiteWebProxyAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetSiteWebProxy().GetAddSection()
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
		resp.Diagnostics.AddError("site web proxy addSection failed", st)
		return
	}

	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := convertCreateInput2(&plan, add)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *siteWebProxySectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SiteWebProxySection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSiteWebProxy().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "site web proxy policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := convertReadState(&state, section, pol, &resp.Diagnostics)

	if !found {
		tflog.Warn(ctx, "site web proxy section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *siteWebProxySectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SiteWebProxySection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate, inputMove := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.resourceAPI().PolicySiteWebProxyMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetSiteWebProxy().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetSiteWebProxy().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetSiteWebProxy().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.resourceAPI().PolicySiteWebProxyUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetSiteWebProxy().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetSiteWebProxy().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetSiteWebProxy().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *siteWebProxySectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SiteWebProxySection
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
	if _, err := r.resourceAPI().PolicySiteWebProxyRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
