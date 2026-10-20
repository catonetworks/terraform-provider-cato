package lanindex

import "github.com/hashicorp/terraform-plugin-framework/types"

type SocketLanSection struct {
	ID      types.String `tfsdk:"id"`
	At      types.Object `tfsdk:"at"`
	Section types.Object `tfsdk:"section"` // idname.Ref
}
