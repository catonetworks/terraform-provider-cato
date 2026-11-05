package clientconnectivitysection

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// resourceModel contains Terraform values for client_connectivity_section.
type resourceModel struct {
	At      types.Object `tfsdk:"at"`
	ID      types.String `tfsdk:"id"`
	Section types.Object `tfsdk:"section"`
}

// AtModel contains Terraform values for at.
type AtModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

var atAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}

// SectionModel contains Terraform values for section.
type SectionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var sectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}
