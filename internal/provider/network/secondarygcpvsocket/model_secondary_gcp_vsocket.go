package secondarygcpvsocket

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID        types.String `tfsdk:"id"`
	Site      types.Object `tfsdk:"site"`
	GCPConfig types.Object `tfsdk:"gcp_config"`
}
