package allocatedip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AllocatedIP is the Terraform model for the cato_allocated_ip resource.
//
// Cato assigns the IP at allocation time -- the operator only chooses the PoP
// location (city), description, and type. The display name ("<pop> - <ip>")
// and the IP itself are therefore computed, tenant-specific values.
type AllocatedIP struct {
	ID              types.String `tfsdk:"id"`
	PopLocationName types.String `tfsdk:"pop_location_name"`
	Description     types.String `tfsdk:"description"`
	Type            types.String `tfsdk:"type"`
	IP              types.String `tfsdk:"ip"`
	Name            types.String `tfsdk:"name"`
}
