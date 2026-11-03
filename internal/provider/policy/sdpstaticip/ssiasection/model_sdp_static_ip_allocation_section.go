package ssiasection

import "github.com/hashicorp/terraform-plugin-framework/types"

// SdpStaticIPAllocationSection mirrors WAN section layout for Application Control policy.
type SdpStaticIPAllocationSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
