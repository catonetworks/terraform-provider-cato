package dnssuffixset

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.NetworkConfigDNSCreateSuffixSetInput {
	validateInputs(plan, diags)
	members := convert.PrepareStringList[string](ctx, plan.Suffix, diags)
	if members == nil {
		members = []string{}
	}
	return m.NetworkConfigDNSCreateSuffixSetInput{DNSSuffixSet: []*m.NetworkConfigDNSSuffixSetCreateInput{
		{Name: plan.Name.ValueString(), Suffix: members},
	}}
}
func buildUpdate(ctx context.Context, plan resourceModel, id string, diags *diag.Diagnostics) m.NetworkConfigDNSUpdateSuffixSetInput {
	input := buildCreate(ctx, plan, diags).DNSSuffixSet[0]
	return m.NetworkConfigDNSUpdateSuffixSetInput{DNSSuffixSet: []*m.NetworkConfigDNSSuffixSetUpdateInput{
		{ID: id, Name: &input.Name, Suffix: input.Suffix},
	}}
}

type remoteRecord interface {
	GetID() string
	GetName() string
	GetSuffix() []string
}

func parseRecord(ctx context.Context, item remoteRecord, diags *diag.Diagnostics) resourceModel {
	members := item.GetSuffix()
	if members == nil {
		members = []string{}
	}
	return resourceModel{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName()),
		Suffix: convert.ParseStringList(ctx, members, diags)}
}
func firstID[T remoteRecord](items []T) string {
	for _, item := range items {
		if item.GetID() != "" {
			return item.GetID()
		}
	}
	return ""
}
