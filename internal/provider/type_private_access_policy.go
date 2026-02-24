package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivAccessPolicyModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Audit   types.Object `tfsdk:"audit"` // PolicyAudit
}

type PolicyAudit struct {
	PublishedBy   types.String `tfsdk:"published_by"`
	PublishedTime types.String `tfsdk:"published_time"`
}

var PolicyAuditTypes = map[string]attr.Type{
	"published_by":   types.StringType,
	"published_time": types.StringType,
}
