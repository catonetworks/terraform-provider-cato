package tlspolicy

import "github.com/hashicorp/terraform-plugin-framework/types"

// TLSInspectPolicyModel holds account-level TLS Inspection policy toggles.
type TLSInspectPolicyModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	PublishedTime types.String `tfsdk:"published_time"`
	PublishedBy   types.String `tfsdk:"published_by"`
	RevisionID    types.String `tfsdk:"revision_id"`
	RevisionName  types.String `tfsdk:"revision_name"`
}
