package stsection

import "github.com/hashicorp/terraform-plugin-framework/types"

// SplitTunnelSection mirrors WAN section layout for Application Control policy.
type SplitTunnelSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
