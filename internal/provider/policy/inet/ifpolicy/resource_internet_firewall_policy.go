package ifpolicy

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
	_ resource.Resource                = &internetFirewallPolicyResource{}
	_ resource.ResourceWithConfigure   = &internetFirewallPolicyResource{}
	_ resource.ResourceWithImportState = &internetFirewallPolicyResource{}
)

func NewInternetFirewallPolicyResource() resource.Resource {
	return &internetFirewallPolicyResource{}
}

type internetFirewallPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *internetFirewallPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_internet_firewall_policy"
}

func (r *internetFirewallPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *internetFirewallPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "internet_firewall")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *internetFirewallPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan InternetFirewallPolicyModel
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

func (r *internetFirewallPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *internetFirewallPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan InternetFirewallPolicyModel
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

func (r *internetFirewallPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_internet_firewall_policy delete is a no-op; policy always exists")
}

func (r *internetFirewallPolicyResource) updatePolicy(
	ctx context.Context,
	plan InternetFirewallPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyInternetFirewallUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyInternetFirewallUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetInternetFirewall().GetUpdatePolicy()
	if err := sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors()); err != nil {
		diags.AddError("Policy update failed", err.Error())
		return diags
	}

	diags.Append(r.publishPolicyRevision(ctx)...)
	return diags
}

// publishPolicyRevision publishes the Internet Firewall draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound —
// that is treated as success by sharedpolicy.PublishResult.
func (r *internetFirewallPolicyResource) publishPolicyRevision(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyInternetFirewallPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetInternetFirewall().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishInternetFirewallPolicyRevision", err.Error())
	}
	return diags
}

func (r *internetFirewallPolicyResource) readState(ctx context.Context) (InternetFirewallPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().InternetFirewallPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetInternetFirewall().GetPolicy()
	if pol == nil {
		return out, fmt.Errorf("policy read did not return a policy")
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
