package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivAccessPolicyModel struct {
	Audit   types.Object `tfsdk:"audit"` // PolicyAudit
	Enabled types.Bool   `tfsdk:"enabled"`
	ID      types.String `tfsdk:"id"`
}

type PolicyAudit struct {
	PublishedBy   types.String `tfsdk:"published_by"`
	PublishedTime types.String `tfsdk:"published_time"`
}

var PolicyAuditTypes = map[string]attr.Type{
	"published_by":   types.StringType,
	"published_time": types.StringType,
}
