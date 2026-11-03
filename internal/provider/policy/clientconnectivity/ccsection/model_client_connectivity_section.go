package ccsection

import "github.com/hashicorp/terraform-plugin-framework/types"

// ClientConnectivitySection mirrors WAN section layout for Application Control policy.
type ClientConnectivitySection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
