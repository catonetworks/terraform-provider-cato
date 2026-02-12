package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivAccessPolicyModel struct {
	ID       types.String    `tfsdk:"id"`
	Enabled  types.Bool      `tfsdk:"enabled"`
	Audit    *PolicyAudit    `tfsdk:"audit"`
	Revision *PolicyRevision `tfsdk:"revision"`
}

type PolicyAudit struct {
	PublishedBy   types.String `tfsdk:"published_by"`
	PublishedTime types.String `tfsdk:"published_time"`
}

type PolicyRevision struct {
	Changes     types.Int64  `tfsdk:"changes"`
	CreatedTime types.String `tfsdk:"createdTime"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"iD"`
	Name        types.String `tfsdk:"name"`
	UpdatedTime types.String `tfsdk:"updatedTime"`
}
