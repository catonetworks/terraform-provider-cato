package permissioncatalog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

type catalogDataSource struct{ iam.DataConfig }

var _ datasource.DataSourceWithConfigure = (*catalogDataSource)(nil)

// NewDataSource creates the account permission catalog query.
func NewDataSource() datasource.DataSource { return &catalogDataSource{} }
func (d *catalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_catalog"
}
func (d *catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := d.Account(config.AccountID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid catalog account", err.Error())
		return
	}
	response, err := d.Reader().RbacRoleManagementPermissionCatalog(ctx, account)
	if err != nil {
		resp.Diagnostics.AddError("Read permission catalog failed", err.Error())
		return
	}
	state, diags := catalogState(ctx, account, response)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
