package appconnectors

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the Terraform model for the cato_app_connectors data source.
type Lookup struct {
	NameFilter  types.List `tfsdk:"name_filter"`
	GroupFilter types.List `tfsdk:"group_filter"`
	Items       types.List `tfsdk:"items"`
}
