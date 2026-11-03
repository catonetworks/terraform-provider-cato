package rpfpolicy

import "github.com/hashicorp/terraform-plugin-framework/types"

// RemotePortFwdPolicyModel holds account-level Remote Port Forwarding policy toggles.
type RemotePortFwdPolicyModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	PublishedTime types.String `tfsdk:"published_time"`
	PublishedBy   types.String `tfsdk:"published_by"`
	RevisionID    types.String `tfsdk:"revision_id"`
	RevisionName  types.String `tfsdk:"revision_name"`
}
