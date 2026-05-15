package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type InternetFirewallSubPolicy struct {
	ID     types.String `tfsdk:"id"`
	At     types.Object `tfsdk:"at"`
	Policy types.Object `tfsdk:"policy"`
	Scope  types.Object `tfsdk:"scope"`
}

type InternetFirewallSubPolicyInfo struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type InternetFirewallSubPolicyScope struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}
