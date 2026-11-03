package alwaysonpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	access "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &ztnaAlwaysOnPolicyResource{}
	_ resource.ResourceWithConfigure   = &ztnaAlwaysOnPolicyResource{}
	_ resource.ResourceWithImportState = &ztnaAlwaysOnPolicyResource{}
)

func NewZtnaAlwaysOnPolicyResource() resource.Resource {
	return &ztnaAlwaysOnPolicyResource{}
}

type ztnaAlwaysOnPolicyResource struct {
	client *catoclient.Service

	apiClient ZtnaPolicyResourceClient
}

func (r *ztnaAlwaysOnPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_ztna_always_on_policy"
}

func (r *ztnaAlwaysOnPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *ztnaAlwaysOnPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "ztna_always_on")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *ztnaAlwaysOnPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ZtnaAlwaysOnPolicyModel
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

func (r *ztnaAlwaysOnPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *ztnaAlwaysOnPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ZtnaAlwaysOnPolicyModel
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

func (r *ztnaAlwaysOnPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_ztna_always_on_policy delete is a no-op; policy always exists")
}

func (r *ztnaAlwaysOnPolicyResource) updatePolicy(
	ctx context.Context,
	plan ZtnaAlwaysOnPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicyZtnaAlwaysOnUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyZtnaAlwaysOnUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetZtnaAlwaysOn().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(access.PublishZtnaAlwaysOnPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *ztnaAlwaysOnPolicyResource) readState(ctx context.Context) (ZtnaAlwaysOnPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetZtnaAlwaysOn().GetPolicy()
	if pol == nil {
		return out, nil
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
