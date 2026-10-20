package wfrulesindex

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WanRuleIndexLookup struct {
	Rules types.List `tfsdk:"rules"`
}
