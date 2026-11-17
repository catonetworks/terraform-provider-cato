package socketlanpolicy

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
	_ resource.Resource                = &socketLanPolicyResource{}
	_ resource.ResourceWithConfigure   = &socketLanPolicyResource{}
	_ resource.ResourceWithImportState = &socketLanPolicyResource{}
)

func NewSocketLanPolicyResource() resource.Resource {
	return &socketLanPolicyResource{}
}

type socketLanPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *socketLanPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_policy"
}

func (r *socketLanPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketLanPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "socket_lan")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *socketLanPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SocketLanPolicyModel
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

func (r *socketLanPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *socketLanPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SocketLanPolicyModel
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

func (r *socketLanPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_socket_lan_policy delete is a no-op; policy always exists")
}

func (r *socketLanPolicyResource) updatePolicy(
	ctx context.Context,
	plan SocketLanPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicySocketLanUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicySocketLanUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetSocketLan().GetUpdatePolicy()
	if err := sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors()); err != nil {
		diags.AddError("Policy update failed", err.Error())
		return diags
	}

	diags.Append(r.publishPolicyRevision(ctx)...)
	return diags
}

// publishPolicyRevision publishes the Socket LAN draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound —
// that is treated as success by sharedpolicy.PublishResult.
func (r *socketLanPolicyResource) publishPolicyRevision(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicySocketLanPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishSocketLanPolicyRevision", err.Error())
	}
	return diags
}

func (r *socketLanPolicyResource) readState(ctx context.Context) (SocketLanPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetSocketLan().GetPolicy()
	if pol == nil {
		return out, fmt.Errorf("policy read did not return a policy")
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
