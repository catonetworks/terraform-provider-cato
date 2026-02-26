package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAccessRuleBulkModel struct {
	RuleData types.Map   `tfsdk:"rule_data"` // map[rule_name]PrivateAccessBulkRule
	Publish  types.Int64 `tfsdk:"publish"`
}

type PrivateAccessBulkRule struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Index    types.Int64  `tfsdk:"index"`
	CMAIndex int64        `tfsdk:"-"`
}

var PrivateAccessBulkRuleTypes = map[string]attr.Type{
	"id":    types.StringType,
	"name":  types.StringType,
	"index": types.Int64Type,
}
