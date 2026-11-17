package tlspolicy

import (
	"context"
	"fmt"

	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

var (
	_ resource.Resource                = &tlsInspectPolicyResource{}
	_ resource.ResourceWithConfigure   = &tlsInspectPolicyResource{}
	_ resource.ResourceWithImportState = &tlsInspectPolicyResource{}
)

func NewTLSInspectPolicyResource() resource.Resource {
	return &tlsInspectPolicyResource{}
}

type tlsInspectPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *tlsInspectPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_tls_policy"
}

func (r *tlsInspectPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *tlsInspectPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "tls")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *tlsInspectPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan TLSInspectPolicyModel
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

func (r *tlsInspectPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *tlsInspectPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan TLSInspectPolicyModel
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

func (r *tlsInspectPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_tls_policy delete is a no-op; policy always exists")
}

func (r *tlsInspectPolicyResource) updatePolicy(
	ctx context.Context,
	plan TLSInspectPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyTLSInspectUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyTLSInspectUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetTLSInspect().GetUpdatePolicy()
	if err := sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors()); err != nil {
		diags.AddError("Policy update failed", err.Error())
		return diags
	}

	diags.Append(r.publishPolicyRevision(ctx)...)
	return diags
}

// publishPolicyRevision publishes the TLS Inspection draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound —
// that is treated as success by sharedpolicy.PublishResult.
func (r *tlsInspectPolicyResource) publishPolicyRevision(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyTLSInspectPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetTLSInspect().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishTLSInspectPolicyRevision", err.Error())
	}
	return diags
}

func (r *tlsInspectPolicyResource) readState(ctx context.Context) (TLSInspectPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetTLSInspect().GetPolicy()
	if pol == nil {
		return out, fmt.Errorf("policy read did not return a policy")
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
