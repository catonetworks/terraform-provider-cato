package wanindex

import "github.com/hashicorp/terraform-plugin-framework/types"

type WanFirewallSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"`
}

// PolicyAddSectionInfoInput for adding section info to a policy
type PolicyAddSectionInfoInput struct {
	Name string `json:"name"`
}

// type PolicySectionPositionInput struct {
// 	Position types.String `tfsdk:"position"`
// 	Ref      types.String `tfsdk:"ref"`
// }

// type PolicyAddSectionInfoInput struct {
// 	Id   types.String `tfsdk:"id"`
// 	Name types.String `tfsdk:"name"`
// }

// type PolicyUpdateSectionInfoInput struct {
// 	Id   types.String `tfsdk:"id"`
// 	Name types.String `tfsdk:"name"`
// }
