package ssiasection

import "github.com/hashicorp/terraform-plugin-framework/types"

// SdpStaticIpAllocationSection mirrors WAN section layout for Application Control policy.
type SdpStaticIpAllocationSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
