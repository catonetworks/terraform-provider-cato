package rpfsection

import "github.com/hashicorp/terraform-plugin-framework/types"

// RemotePortFwdSection mirrors WAN section layout for Application Control policy.
type RemotePortFwdSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
