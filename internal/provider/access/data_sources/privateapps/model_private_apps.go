package privateapps

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the Terraform model for the cato_private_apps data source.
type Lookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	IDFilter   types.List `tfsdk:"id_filter"`
	Items      types.List `tfsdk:"items"`
}
