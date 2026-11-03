package alwaysonsection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlwaysOnSection is the Terraform model for the cato_always_on_section
// resource (a ZTNA Always-On policy section).
type AlwaysOnSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`      // policyrule.Position
	Section types.Object `tfsdk:"section"` // idname.Ref
}
