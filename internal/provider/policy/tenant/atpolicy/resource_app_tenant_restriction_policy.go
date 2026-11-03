package atpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	tenant "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &appTenantRestrictionPolicyResource{}
	_ resource.ResourceWithConfigure   = &appTenantRestrictionPolicyResource{}
	_ resource.ResourceWithImportState = &appTenantRestrictionPolicyResource{}
)

func NewAppTenantRestrictionPolicyResource() resource.Resource {
	return &appTenantRestrictionPolicyResource{}
}

type appTenantRestrictionPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *appTenantRestrictionPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_app_tenant_restriction_policy"
}

func (r *appTenantRestrictionPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *appTenantRestrictionPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "app_tenant_restriction")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *appTenantRestrictionPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan AppTenantRestrictionPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *appTenantRestrictionPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *appTenantRestrictionPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan AppTenantRestrictionPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *appTenantRestrictionPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_app_tenant_restriction_policy delete is a no-op; policy always exists")
}

func (r *appTenantRestrictionPolicyResource) updatePolicy(
	ctx context.Context,
	plan AppTenantRestrictionPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyAppTenantRestrictionUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyAppTenantRestrictionUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetAppTenantRestriction().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(tenant.PublishAppTenantRestrictionPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *appTenantRestrictionPolicyResource) readState(ctx context.Context) (AppTenantRestrictionPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
	if pol == nil {
		return out, nil
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
