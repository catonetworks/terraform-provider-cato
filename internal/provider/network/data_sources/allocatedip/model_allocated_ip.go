package allocatedip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Lookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	// DescriptionFilter / PopLocationFilter select allocations by their stable
	// cross-tenant identity (description, popLocation) -- the display name
	// embeds the Cato-ASSIGNED ip, so names never match across tenants.
	DescriptionFilter types.List `tfsdk:"description_filter"`
	PopLocationFilter types.List `tfsdk:"pop_location_filter"`
	Items             types.List `tfsdk:"items"`
}

type AllocatedIP struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	CountryCode types.String `tfsdk:"cc"`
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
}
