package ifrulesections

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ifRuleSectionLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}
