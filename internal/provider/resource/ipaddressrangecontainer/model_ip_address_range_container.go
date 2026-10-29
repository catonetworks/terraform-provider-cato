package ipaddressrangecontainer

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Mode        types.String `tfsdk:"mode"`
	FileType    types.String `tfsdk:"file_type"`
	Size        types.Int64  `tfsdk:"size"`
	SyncData    types.Object `tfsdk:"sync_data"`
	Values      types.Set    `tfsdk:"values"`
}
