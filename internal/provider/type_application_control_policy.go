package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// ApplicationControlPolicyModel holds account-level Application Control policy toggles.
type ApplicationControlPolicyModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	DataControlEnabled types.String `tfsdk:"data_control_enabled"`
	PublishedTime      types.String `tfsdk:"published_time"`
	PublishedBy        types.String `tfsdk:"published_by"`
	RevisionID         types.String `tfsdk:"revision_id"`
	RevisionName       types.String `tfsdk:"revision_name"`
}
