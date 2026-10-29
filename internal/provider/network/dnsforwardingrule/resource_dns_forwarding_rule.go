package dnsforwardingrule

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type forwardingRuleResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDNSForwardingRuleResource returns the standalone DNS forwarding rule resource.
func NewDNSForwardingRuleResource() resource.Resource { return &forwardingRuleResource{} }

func (r *forwardingRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_forwarding_rule"
}

func (r *forwardingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *forwardingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ForwardingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDNSCreateForwardingRule(ctx, r.client.AccountId,
		models.NetworkConfigDNSCreateForwardingRuleInput{ForwardingRule: []*models.NetworkConfigDNSForwardingRuleCreateInput{&input}},
	)
	if id := createdID(result); id != "" {
		retained := plan
		retained.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &retained)...)
	}

	if err != nil {
		nc.APIError(&resp.Diagnostics, "createForwardingRule", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.DNS.CreateForwardingRule != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.DNS.CreateForwardingRule.ForwardingRule
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID != "", &resp.Diagnostics) {
		return
	}
	parseForwardingRule(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *forwardingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ForwardingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDNSForwardingRule(ctx, r.client.AccountId,
		models.NetworkConfigDNSForwardingRuleRefInput{By: models.ObjectRefByID, Input: state.ID.ValueString()},
	)
	if err != nil {
		if nc.IsNotFound(err, "forwardingRule") {
			resp.State.RemoveResource(ctx)
		} else {
			nc.APIError(&resp.Diagnostics, "forwardingRule", err)
		}
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.DNS.ForwardingRule
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !nc.RequirePayload(item.ID == state.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseForwardingRule(ctx, &state, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *forwardingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ForwardingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDNSUpdateForwardingRule(ctx, r.client.AccountId,
		models.NetworkConfigDNSUpdateForwardingRuleInput{ForwardingRule: []*models.NetworkConfigDNSForwardingRuleUpdateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateForwardingRule", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.DNS.UpdateForwardingRule != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.DNS.UpdateForwardingRule.ForwardingRule
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID == plan.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseForwardingRule(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *forwardingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ForwardingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDNSDeleteForwardingRule(ctx, r.client.AccountId,
		models.NetworkConfigDNSDeleteForwardingRuleInput{ID: []string{state.ID.ValueString()}},
	)
	if err != nil {
		if !nc.IsNotFound(err, "deleteForwardingRule") && !r.confirmDeleted(ctx, state.ID.ValueString(), err) {
			nc.APIError(&resp.Diagnostics, "deleteForwardingRule", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.DeleteForwardingRule != nil,
		&resp.Diagnostics)
}

func (r *forwardingRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}
