package allocatedip

// Terraform value construction for cato_allocated_ip (API → state mapping).

import (
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// displayName renders the CMA entity display name for an allocation.
func displayName(popLocation, ip string) string {
	return fmt.Sprintf("%s - %s", popLocation, ip)
}

// applyAllocationToState maps a listed allocation onto the Terraform model.
// The description keeps its null-vs-empty handling: a null config attribute
// must not flap between null and "".
func applyAllocationToState(
	state *AllocatedIP,
	item *cato_go_sdk.PopLocationsAllocatedIpList_PopLocations_AllocatedIPList_AllocatedIP,
) {
	state.ID = types.StringValue(item.GetID())
	state.IP = types.StringValue(item.GetIP())
	state.PopLocationName = types.StringValue(item.GetPopLocation().GetName())
	state.Name = types.StringValue(displayName(item.GetPopLocation().GetName(), item.GetIP()))
	state.Type = types.StringValue(string(item.AllocationType))
	if desc := item.GetDescription(); desc != nil && *desc != "" {
		state.Description = types.StringValue(*desc)
	} else if !state.Description.IsNull() && state.Description.ValueString() != "" {
		state.Description = types.StringNull()
	}
}
