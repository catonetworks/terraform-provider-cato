package diasection

import "github.com/hashicorp/terraform-plugin-framework/types"

// DynamicIPAllocationSection mirrors WAN section layout for Application Control policy.
type DynamicIPAllocationSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
