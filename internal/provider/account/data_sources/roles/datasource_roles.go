package roles

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

type rolesDataSource struct{ iam.DataConfig }

var _ datasource.DataSourceWithConfigure = (*rolesDataSource)(nil)

// NewDataSource creates the paginated role list query.
func NewDataSource() datasource.DataSource { return &rolesDataSource{} }
func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}
func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := d.Account(config.AccountID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	input, diags := queryInput(ctx, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	records, err := iam.ListRoles(ctx, d.Reader(), account, input)
	if err != nil {
		resp.Diagnostics.AddError("List roles failed", err.Error())
		return
	}
	values := make([]attr.Value, 0, len(records))
	for _, record := range records {
		model, diags := iam.RoleState(ctx, account, record)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, diags := types.ObjectValueFrom(ctx, iam.RoleAttrTypes, model)
		resp.Diagnostics.Append(diags...)
		values = append(values, value)
	}
	config.ID = types.StringValue(account)
	config.AccountID = types.StringValue(account)
	config.SortDirection = types.StringValue(string(input.Sort.Name.Direction))
	config.Total = types.Int64Value(int64(len(values)))
	config.Items, diags = types.ListValue(iam.RoleObjectType, values)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	}
}
