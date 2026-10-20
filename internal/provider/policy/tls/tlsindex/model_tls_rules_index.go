package tlsindex

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RuleIndexLookup struct {
	Rules types.List `tfsdk:"rules"`
}
