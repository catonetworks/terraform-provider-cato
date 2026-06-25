package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// AppTenantRestrictionSection mirrors WAN section layout for Application Control policy.
type AppTenantRestrictionSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
