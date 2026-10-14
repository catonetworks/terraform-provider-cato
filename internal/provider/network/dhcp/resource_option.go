package dhcp

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type optionResource struct{ nc.Resource }

func NewOptionResource() resource.Resource { return &optionResource{} }
func (r *optionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_option"
}
func (r *optionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *optionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := models.NetworkConfigDhcpOptionCreateInput{Tag: plan.Tag.ValueInt64(),
		Value:       plan.Value.ValueString(),
		Type:        models.NetworkConfigDhcpOptionType(plan.Type.ValueString()),
		Description: plan.Description.ValueStringPointer()}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDhcpCreateOption(ctx, r.Client.AccountId,
		models.NetworkConfigDhcpCreateOptionInput{Option: []*models.NetworkConfigDhcpOptionCreateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "createOption", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.CreateOption != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.CreateOption.Option
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID != "", &resp.Diagnostics) {
		return
	}
	hydrateOption(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *optionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDhcpOption(ctx, r.Client.AccountId,
		models.NetworkConfigDhcpOptionRefInput{By: models.ObjectRefByID, Input: state.ID.ValueString()},
	)
	if err != nil {
		if nc.IsNotFound(err, "option") {
			resp.State.RemoveResource(ctx)
		} else {
			nc.APIError(&resp.Diagnostics, "option", err)
		}
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.Dhcp.Option
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !nc.RequirePayload(item.ID == state.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	hydrateOption(ctx, &state, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *optionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := models.NetworkConfigDhcpOptionUpdateInput{ID: plan.ID.ValueString(),
		Tag:         plan.Tag.ValueInt64Pointer(),
		Value:       plan.Value.ValueStringPointer(),
		Type:        typePointer(plan.Type.ValueString()),
		Description: descriptionPointer(plan.Description)}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDhcpUpdateOption(ctx, r.Client.AccountId,
		models.NetworkConfigDhcpUpdateOptionInput{Option: []*models.NetworkConfigDhcpOptionUpdateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateOption", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.UpdateOption != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.UpdateOption.Option
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID == plan.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	hydrateOption(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *optionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.Catov2.NetworkConfigDhcpDeleteOption(ctx, r.Client.AccountId,
		models.NetworkConfigDhcpDeleteOptionInput{ID: []string{state.ID.ValueString()}},
	)
	if err != nil {
		if !nc.IsNotFound(err, "deleteOption") {
			nc.APIError(&resp.Diagnostics, "deleteOption", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.DeleteOption != nil, &resp.Diagnostics)
}

type optionAPI interface {
	GetID() string
	GetTag() int64
	GetValue() string
	GetType() *models.NetworkConfigDhcpOptionType
	GetDescription() *string
}

func hydrateOption(_ context.Context, plan *OptionModel, item optionAPI, diags *diag.Diagnostics) {
	plan.ID = types.StringValue(item.GetID())
	if !nc.RequirePayload(item.GetType() != nil, diags) {
		return
	}
	plan.Tag = types.Int64Value(item.GetTag())
	plan.Value = types.StringValue(item.GetValue())
	plan.Type = types.StringValue(string(*item.GetType()))
	plan.Description = types.StringPointerValue(item.GetDescription())
	if plan.Description.ValueString() == "" {
		plan.Description = types.StringNull()
	}
}

func typePointer(value string) *models.NetworkConfigDhcpOptionType {
	result := models.NetworkConfigDhcpOptionType(value)
	return &result
}

// Empty string clears a description using the generated SDK's non-nil pointer.
func descriptionPointer(value types.String) *string {
	description := value.ValueString()
	return &description
}
