package dnsserverset

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Server types.List   `tfsdk:"server"`
}
