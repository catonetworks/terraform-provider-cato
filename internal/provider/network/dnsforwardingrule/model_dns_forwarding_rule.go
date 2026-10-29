package dnsforwardingrule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ForwardingRuleModel contains one forwarding domain and its ordered server list.
type ForwardingRuleModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Server types.List   `tfsdk:"server"`
}
