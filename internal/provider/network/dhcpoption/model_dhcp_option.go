package dhcpoption

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OptionModel contains one account DHCP option.
type OptionModel struct {
	ID          types.String `tfsdk:"id"`
	Tag         types.Int64  `tfsdk:"tag"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}
