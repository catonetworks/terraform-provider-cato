package swppolicy

import (
	"context"
	"fmt"

	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	sitewebproxy "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

var (
	_ resource.Resource                = &siteWebProxyPolicyResource{}
	_ resource.ResourceWithConfigure   = &siteWebProxyPolicyResource{}
	_ resource.ResourceWithImportState = &siteWebProxyPolicyResource{}
)

func NewSiteWebProxyPolicyResource() resource.Resource {
	return &siteWebProxyPolicyResource{}
}

type siteWebProxyPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *siteWebProxyPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_site_web_proxy_policy"
}

func (r *siteWebProxyPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *siteWebProxyPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "site_web_proxy")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *siteWebProxyPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SiteWebProxyPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *siteWebProxyPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *siteWebProxyPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SiteWebProxyPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *siteWebProxyPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_site_web_proxy_policy delete is a no-op; policy always exists")
}

func (r *siteWebProxyPolicyResource) updatePolicy(
	ctx context.Context,
	plan SiteWebProxyPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicySiteWebProxyUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicySiteWebProxyUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetSiteWebProxy().GetUpdatePolicy()
	if err := sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors()); err != nil {
		diags.AddError("Policy update failed", err.Error())
		return diags
	}

	diags.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *siteWebProxyPolicyResource) readState(ctx context.Context) (SiteWebProxyPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetSiteWebProxy().GetPolicy()
	if pol == nil {
		return out, fmt.Errorf("policy read did not return a policy")
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
