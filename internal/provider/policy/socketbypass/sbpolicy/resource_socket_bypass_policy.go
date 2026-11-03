package sbpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	socketbypass "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &socketBypassPolicyResource{}
	_ resource.ResourceWithConfigure   = &socketBypassPolicyResource{}
	_ resource.ResourceWithImportState = &socketBypassPolicyResource{}
)

func NewSocketBypassPolicyResource() resource.Resource {
	return &socketBypassPolicyResource{}
}

type socketBypassPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *socketBypassPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_bypass_policy"
}

func (r *socketBypassPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketBypassPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "socket_bypass")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *socketBypassPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SocketBypassPolicyModel
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

func (r *socketBypassPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *socketBypassPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SocketBypassPolicyModel
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

func (r *socketBypassPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_socket_bypass_policy delete is a no-op; policy always exists")
}

func (r *socketBypassPolicyResource) updatePolicy(
	ctx context.Context,
	plan SocketBypassPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicySocketBypassUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicySocketBypassUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetSocketBypass().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(socketbypass.PublishSocketBypassPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *socketBypassPolicyResource) readState(ctx context.Context) (SocketBypassPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicySocketBypassPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetSocketBypass().GetPolicy()
	if pol == nil {
		return out, nil
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
