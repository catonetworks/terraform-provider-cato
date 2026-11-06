package pasection

import "github.com/hashicorp/terraform-plugin-framework/types"

// PrivateAccessSection mirrors WAN section layout for Application Control policy.
type PrivateAccessSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
