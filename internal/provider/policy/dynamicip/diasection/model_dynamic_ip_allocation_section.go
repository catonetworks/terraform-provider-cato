package diasection

import "github.com/hashicorp/terraform-plugin-framework/types"

// DynamicIpAllocationSection mirrors WAN section layout for Application Control policy.
type DynamicIpAllocationSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
