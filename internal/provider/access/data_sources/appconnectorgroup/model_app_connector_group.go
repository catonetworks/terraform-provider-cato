package appconnectorgroup

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupDataSourceModel struct {
	Items []types.String `tfsdk:"items"`
	Total types.Int64    `tfsdk:"total"`
}
