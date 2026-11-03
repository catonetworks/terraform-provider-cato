package tspolicy

import (
	"context"
	"fmt"

	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	terminalserver "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

var (
	_ resource.Resource                = &terminalServerPolicyResource{}
	_ resource.ResourceWithConfigure   = &terminalServerPolicyResource{}
	_ resource.ResourceWithImportState = &terminalServerPolicyResource{}
)

func NewTerminalServerPolicyResource() resource.Resource {
	return &terminalServerPolicyResource{}
}

type terminalServerPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *terminalServerPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_terminal_server_policy"
}

func (r *terminalServerPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *terminalServerPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "terminal_server")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *terminalServerPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan TerminalServerPolicyModel
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

func (r *terminalServerPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *terminalServerPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan TerminalServerPolicyModel
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

func (r *terminalServerPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_terminal_server_policy delete is a no-op; policy always exists")
}

func (r *terminalServerPolicyResource) updatePolicy(
	ctx context.Context,
	plan TerminalServerPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyTerminalServerUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyTerminalServerUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetTerminalServer().GetUpdatePolicy()
	if err := sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors()); err != nil {
		diags.AddError("Policy update failed", err.Error())
		return diags
	}

	diags.Append(terminalserver.PublishTerminalServerPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *terminalServerPolicyResource) readState(ctx context.Context) (TerminalServerPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicyTerminalServerPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetTerminalServer().GetPolicy()
	if pol == nil {
		return out, fmt.Errorf("policy read did not return a policy")
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
