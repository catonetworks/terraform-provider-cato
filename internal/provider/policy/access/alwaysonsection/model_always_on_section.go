package alwaysonsection

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// Section is the Terraform model for cato_always_on_section.
type Section struct {
	At      types.Object `tfsdk:"at"`      // policyrule.Position
	Section types.Object `tfsdk:"section"` // SectionData
}

// SectionAttrTypes defines the resource's Terraform attribute types.
var SectionAttrTypes = map[string]attr.Type{
	"at":      policyrule.PositionObjectType,
	"section": SectionDataObjectType,
}

// SectionObjectType defines the resource's Terraform object type.
var SectionObjectType = types.ObjectType{AttrTypes: SectionAttrTypes}

// SectionData contains a section's name and computed identifier.
type SectionData struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// SectionDataAttrTypes defines the nested section attribute types.
var SectionDataAttrTypes = map[string]attr.Type{"id": types.StringType, "name": types.StringType}

// SectionDataObjectType defines the nested section object type.
var SectionDataObjectType = types.ObjectType{AttrTypes: SectionDataAttrTypes}
