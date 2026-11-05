package tssection

import "github.com/hashicorp/terraform-plugin-framework/types"

// TerminalServerSection mirrors WAN section layout for Application Control policy.
type TerminalServerSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
