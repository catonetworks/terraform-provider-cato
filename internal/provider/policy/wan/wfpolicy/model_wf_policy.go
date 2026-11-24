package wfpolicy

import "github.com/hashicorp/terraform-plugin-framework/types"

// WanFirewallPolicyModel holds the account-level WAN Firewall policy toggle.
type WanFirewallPolicyModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	PublishedTime types.String `tfsdk:"published_time"`
	PublishedBy   types.String `tfsdk:"published_by"`
	RevisionID    types.String `tfsdk:"revision_id"`
	RevisionName  types.String `tfsdk:"revision_name"`
}
