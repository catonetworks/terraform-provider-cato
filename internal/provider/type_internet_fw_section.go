package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type InternetFirewallSection struct {
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}

type PolicySectionPositionInput struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

type PolicyAddSectionInfoInput struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type PolicyUpdateSectionInfoInput struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
