package allocatedip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Lookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type AllocatedIP struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	CountryCode types.String `tfsdk:"cc"`
}
