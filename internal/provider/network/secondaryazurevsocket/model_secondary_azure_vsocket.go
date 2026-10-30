package secondaryazurevsocket

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	Site        types.Object `tfsdk:"site"`
	InterfaceIP types.String `tfsdk:"interface_ip"`
	FloatingIP  types.String `tfsdk:"floating_ip"`
}
