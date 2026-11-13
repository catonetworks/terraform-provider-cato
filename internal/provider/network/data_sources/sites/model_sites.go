package sites

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Lookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type Site struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Type              types.String `tfsdk:"type"`
	ConnectionType    types.String `tfsdk:"connection_type"`
	OperationalStatus types.String `tfsdk:"operational_status"`
}

type siteFilter struct {
	namesMap map[string]struct{}
}
