package tlssection

import "github.com/hashicorp/terraform-plugin-framework/types"

type InspectionSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}
