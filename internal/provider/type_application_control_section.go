package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// ApplicationControlSection mirrors WAN section layout for Application Control policy.
type ApplicationControlSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
