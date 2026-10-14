package dns

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type forwardingRuleResource struct{ nc.Resource }

func NewForwardingRuleResource() resource.Resource { return &forwardingRuleResource{} }
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
	input := models.NetworkConfigDNSForwardingRuleCreateInput{Domain: plan.Domain.ValueString(),
		Server: nc.Strings(ctx,
			plan.Server,
			&resp.Diagnostics)}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDNSCreateForwardingRule(ctx, r.Client.AccountId,
		models.NetworkConfigDNSCreateForwardingRuleInput{ForwardingRule: []*models.NetworkConfigDNSForwardingRuleCreateInput{&input}},
	)
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
	hydrateForwardingRule(ctx, &plan, items[0], &resp.Diagnostics)
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
	result, err := r.Client.Catov2.NetworkConfigDNSForwardingRule(ctx, r.Client.AccountId,
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
	hydrateForwardingRule(ctx, &state, item, &resp.Diagnostics)
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
	input := models.NetworkConfigDNSForwardingRuleUpdateInput{ID: plan.ID.ValueString(),
		Domain: plan.Domain.ValueStringPointer(),
		Server: nc.Strings(ctx,
			plan.Server,
			&resp.Diagnostics)}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDNSUpdateForwardingRule(ctx, r.Client.AccountId,
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
	hydrateForwardingRule(ctx, &plan, items[0], &resp.Diagnostics)
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
	result, err := r.Client.Catov2.NetworkConfigDNSDeleteForwardingRule(ctx, r.Client.AccountId,
		models.NetworkConfigDNSDeleteForwardingRuleInput{ID: []string{state.ID.ValueString()}},
	)
	if err != nil {
		if !nc.IsNotFound(err, "deleteForwardingRule") {
			nc.APIError(&resp.Diagnostics, "deleteForwardingRule", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.DeleteForwardingRule != nil, &resp.Diagnostics)
}

type forwardingRuleAPI interface {
	GetID() string
	GetDomain() string
	GetServer() []string
}

func hydrateForwardingRule(ctx context.Context, plan *ForwardingRuleModel, item forwardingRuleAPI, diags *diag.Diagnostics) {
	plan.ID = types.StringValue(item.GetID())
	plan.Domain = types.StringValue(item.GetDomain())
	plan.Server = nc.List(ctx, item.GetServer(), diags)
}
