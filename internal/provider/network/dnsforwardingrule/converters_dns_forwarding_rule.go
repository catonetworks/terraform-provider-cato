package dnsforwardingrule

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type forwardingRuleAPI interface {
	GetID() string
	GetDomain() string
	GetServer() []string
}

func parseForwardingRule(ctx context.Context, plan *ForwardingRuleModel, item forwardingRuleAPI, diags *diag.Diagnostics) {
	plan.ID = types.StringValue(item.GetID())
	plan.Domain = types.StringValue(item.GetDomain())
	plan.Server = nc.List(ctx, item.GetServer(), diags)
}

func buildCreateInput(ctx context.Context,
	plan ForwardingRuleModel,
	diags *diag.Diagnostics) models.NetworkConfigDNSForwardingRuleCreateInput {
	return models.NetworkConfigDNSForwardingRuleCreateInput{Domain: plan.Domain.ValueString(),
		Server: nc.Strings(ctx, plan.Server, diags)}
}

func buildUpdateInput(ctx context.Context,
	plan ForwardingRuleModel,
	diags *diag.Diagnostics) models.NetworkConfigDNSForwardingRuleUpdateInput {
	return models.NetworkConfigDNSForwardingRuleUpdateInput{ID: plan.ID.ValueString(),
		Domain: plan.Domain.ValueStringPointer(),
		Server: nc.Strings(ctx, plan.Server, diags)}
}
