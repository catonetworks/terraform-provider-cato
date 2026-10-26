package wfrulesections

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type wfRuleSectionLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}
