package groupmembers

import "github.com/hashicorp/terraform-plugin-framework/types"

type Members struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	Members   types.Set    `tfsdk:"members"`
}
