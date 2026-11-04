package ssiapolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	sdpstaticip "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &sdpStaticIpAllocationPolicyResource{}
	_ resource.ResourceWithConfigure   = &sdpStaticIpAllocationPolicyResource{}
	_ resource.ResourceWithImportState = &sdpStaticIpAllocationPolicyResource{}
)

func NewSdpStaticIpAllocationPolicyResource() resource.Resource {
	return &sdpStaticIpAllocationPolicyResource{}
}

type sdpStaticIpAllocationPolicyResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *sdpStaticIpAllocationPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sdp_static_ip_allocation_policy"
}

func (r *sdpStaticIpAllocationPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *sdpStaticIpAllocationPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "sdp_static_ip_allocation")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *sdpStaticIpAllocationPolicyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SdpStaticIpAllocationPolicyModel
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

func (r *sdpStaticIpAllocationPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *sdpStaticIpAllocationPolicyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SdpStaticIpAllocationPolicyModel
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

func (r *sdpStaticIpAllocationPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_sdp_static_ip_allocation_policy delete is a no-op; policy always exists")
}

func (r *sdpStaticIpAllocationPolicyResource) updatePolicy(
	ctx context.Context,
	plan SdpStaticIpAllocationPolicyModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	in := convertUpdatePolicyValues(plan)
	res, err := r.resourceAPI().PolicySdpStaticIPAllocationUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicySdpStaticIPAllocationUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetSdpStaticIPAllocation().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(sdpstaticip.PublishSdpStaticIpAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	return diags
}

func (r *sdpStaticIpAllocationPolicyResource) readState(ctx context.Context) (SdpStaticIpAllocationPolicyModel, error) {
	out := convertReadStateValues()
	body, err := r.resourceAPI().PolicySdpStaticIPAllocationPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetSdpStaticIPAllocation().GetPolicy()
	if pol == nil {
		return out, nil
	}
	convertReadStateValues2(&out, pol)
	return out, nil
}
