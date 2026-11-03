package sbsection

import "github.com/hashicorp/terraform-plugin-framework/types"

// SocketBypassSection mirrors WAN section layout for Application Control policy.
type SocketBypassSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
