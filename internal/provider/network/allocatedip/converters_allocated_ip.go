package allocatedip

import (
	"context"
	"net/netip"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.PopLocationAddAllocatedIPInput {
	validateType(plan.Type, diags)
	if plan.Description.IsUnknown() {
		diags.AddError("Unknown description", "Description must be known before applying.")
	}
	if plan.PopLocation.IsNull() || plan.PopLocation.IsUnknown() {
		diags.AddError("Missing PoP", "A known PoP reference is required for SYSTEM allocation.")
	}
	pop := idname.PrepareIDName[m.PopLocationRefInput](ctx, plan.PopLocation, diags)
	return m.PopLocationAddAllocatedIPInput{Type: m.PopLocationAllocatedIPType(plan.Type.ValueString()),
		Description: convert.StringPointerForOptionalInput(plan.Description), PopLocation: pop}
}
func identityState(plan resourceModel, id, ip string) resourceModel {
	plan.ID = types.StringValue(id)
	plan.IP = types.StringValue(ip)
	plan.State = types.StringValue("")
	plan.Subnet = types.StringValue("")
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if !plan.PopLocation.IsNull() && !plan.PopLocation.IsUnknown() {
		attrs := plan.PopLocation.Attributes()
		for key, value := range attrs {
			if value.IsUnknown() {
				attrs[key] = types.StringValue("")
			}
		}
		plan.PopLocation = types.ObjectValueMust(idname.RefAttrTypes, attrs)
	}
	return plan
}

type remoteAllocation interface {
	GetID() string
	GetIP() string
	GetDescription() *string
	GetAllocationType() *m.PopLocationAllocatedIPType
	GetSubnet() *string
	GetState() *m.PopLocationByoipSubnetState
}
type popRef interface {
	~struct {
		ID   string `json:"id" graphql:"id"`
		Name string `json:"name" graphql:"name"`
	}
}

func parseAllocation[P popRef](ctx context.Context, item remoteAllocation, pop *P, account, expected string,
	diags *diag.Diagnostics,
) resourceModel {
	state := resourceModel{ID: types.StringValue(item.GetID()), IP: types.StringValue(item.GetIP()),
		Description: convert.ParseOptionalString(item.GetDescription()), Subnet: convert.ParseOptionalString(item.GetSubnet()),
		State: types.StringValue(""), Type: types.StringValue(""), PopLocation: types.ObjectNull(idname.RefAttrTypes)}
	if account != expected {
		diags.AddError("Invalid API response", "Allocation belongs to a different or unspecified account.")
	}
	if kind := item.GetAllocationType(); kind != nil {
		state.Type = types.StringValue(string(*kind))
	}
	validateType(state.Type, diags)
	if value := item.GetState(); value != nil {
		state.State = types.StringValue(string(*value))
	}
	if pop != nil {
		state.PopLocation = idname.ParseIDName(ctx, *pop, diags)
	}
	if state.PopLocation.IsNull() || state.PopLocation.Attributes()["id"].(types.String).ValueString() == "" {
		diags.AddError("Invalid API response", "Allocation has no PoP identity.")
	}
	if ip, err := netip.ParseAddr(item.GetIP()); err != nil || !ip.Is4() {
		diags.AddError("Invalid API response", "Allocation has no valid IPv4 address.")
	}
	return state
}
