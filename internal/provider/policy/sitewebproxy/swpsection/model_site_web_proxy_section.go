package swpsection

import "github.com/hashicorp/terraform-plugin-framework/types"

// SiteWebProxySection mirrors WAN section layout for Application Control policy.
type SiteWebProxySection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
