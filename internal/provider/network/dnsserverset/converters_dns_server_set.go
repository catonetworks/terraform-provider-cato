package dnsserverset

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.NetworkConfigDNSCreateServerSetInput {
	validateInputs(plan, diags)
	members := convert.PrepareStringList[string](ctx, plan.Server, diags)
	if members == nil {
		members = []string{}
	}
	return m.NetworkConfigDNSCreateServerSetInput{DNSServerSet: []*m.NetworkConfigDNSServerSetCreateInput{
		{Name: plan.Name.ValueString(), Server: members},
	}}
}
func buildUpdate(ctx context.Context, plan resourceModel, id string, diags *diag.Diagnostics) m.NetworkConfigDNSUpdateServerSetInput {
	input := buildCreate(ctx, plan, diags).DNSServerSet[0]
	return m.NetworkConfigDNSUpdateServerSetInput{DNSServerSet: []*m.NetworkConfigDNSServerSetUpdateInput{
		{ID: id, Name: &input.Name, Server: input.Server},
	}}
}

type remoteRecord interface {
	GetID() string
	GetName() string
	GetServer() []string
}

func parseRecord(ctx context.Context, item remoteRecord, diags *diag.Diagnostics) resourceModel {
	members := item.GetServer()
	if members == nil {
		members = []string{}
	}
	return resourceModel{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName()),
		Server: convert.ParseStringList(ctx, members, diags)}
}
func firstID[T remoteRecord](items []T) string {
	for _, item := range items {
		if item.GetID() != "" {
			return item.GetID()
		}
	}
	return ""
}
