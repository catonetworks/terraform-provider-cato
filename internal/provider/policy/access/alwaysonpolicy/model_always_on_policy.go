package alwaysonpolicy

import "github.com/hashicorp/terraform-plugin-framework/types"

// PolicyModel is the Terraform model for the account's existing Always On policy.
type PolicyModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
