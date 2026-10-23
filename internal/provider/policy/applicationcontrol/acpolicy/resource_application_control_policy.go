package acpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	applicationcontrol "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &applicationControlPolicyResource{}
	_ resource.ResourceWithConfigure   = &applicationControlPolicyResource{}
	_ resource.ResourceWithImportState = &applicationControlPolicyResource{}
)

func NewApplicationControlPolicyResource() resource.Resource {
	return &applicationControlPolicyResource{}
}

type applicationControlPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *applicationControlPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_application_control_policy"
}

func (r *applicationControlPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *applicationControlPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "application_control")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data_control_enabled"), string(cato_models.PolicyToggleStateDisabled))...)
}

func (r *applicationControlPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ApplicationControlPolicyModel
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

func (r *applicationControlPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ApplicationControlPolicyModel
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

func (r *applicationControlPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_application_control_policy delete is a no-op; policy always exists")
}

func (r *applicationControlPolicyResource) updatePolicy(
	ctx context.Context,
	plan ApplicationControlPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyApplicationControlUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyApplicationControlUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetApplicationControl().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *applicationControlPolicyResource) readState(ctx context.Context) (ApplicationControlPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		return out, nil
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
