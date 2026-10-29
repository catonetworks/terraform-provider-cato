package dhcpoption

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type optionAPI interface {
	GetID() string
	GetTag() int64
	GetValue() string
	GetType() *models.NetworkConfigDhcpOptionType
	GetDescription() *string
}

func parseOption(_ context.Context, plan *OptionModel, item optionAPI, diags *diag.Diagnostics) {
	plan.ID = types.StringValue(item.GetID())
	if !nc.RequirePayload(item.GetType() != nil && item.GetType().IsValid(), diags) {
		return
	}
	plan.Tag = types.Int64Value(item.GetTag())
	plan.Value = types.StringValue(item.GetValue())
	plan.Type = types.StringValue(string(*item.GetType()))
	plan.Description = types.StringValue("")
	if item.GetDescription() != nil {
		plan.Description = types.StringValue(*item.GetDescription())
	}
}

func typePointer(value string) *models.NetworkConfigDhcpOptionType {
	result := models.NetworkConfigDhcpOptionType(value)
	return &result
}

func descriptionPointer(value types.String) *string {
	description := value.ValueString()
	return &description
}

func buildCreateInput(_ context.Context, plan OptionModel, _ *diag.Diagnostics) models.NetworkConfigDhcpOptionCreateInput {
	return models.NetworkConfigDhcpOptionCreateInput{Tag: plan.Tag.ValueInt64(),
		Value:       plan.Value.ValueString(),
		Type:        models.NetworkConfigDhcpOptionType(plan.Type.ValueString()),
		Description: createDescription(plan.Description)}
}

func buildUpdateInput(_ context.Context, plan OptionModel, _ *diag.Diagnostics) models.NetworkConfigDhcpOptionUpdateInput {
	return models.NetworkConfigDhcpOptionUpdateInput{ID: plan.ID.ValueString(),
		Tag:         plan.Tag.ValueInt64Pointer(),
		Value:       plan.Value.ValueStringPointer(),
		Type:        typePointer(plan.Type.ValueString()),
		Description: descriptionPointer(plan.Description)}
}

func createDescription(value types.String) *string {
	if value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}
