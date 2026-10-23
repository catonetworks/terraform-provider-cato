package role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

type roleDataSource struct{ iam.DataConfig }

var _ datasource.DataSourceWithConfigure = (*roleDataSource)(nil)

// NewDataSource creates the single-role lookup.
func NewDataSource() datasource.DataSource { return &roleDataSource{} }
func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}
func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := iam.RoleDataAttributes()
	attrs["account_id"] = iam.DataAccountAttribute()
	attrs["role_id"] = schema.StringAttribute{Description: "ID of the custom or predefined role to read.",
		Required:   true,
		Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
	resp.Schema = schema.Schema{Description: "Reads a custom or predefined RBAC role by ID.", Attributes: attrs}
}
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
	if config.RoleID.IsNull() || config.RoleID.IsUnknown() || config.RoleID.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid role ID", "role_id must be known and nonempty.")
		return
	}
	value, err := iam.LookupRole(ctx, d.Reader(), account, config.RoleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read role failed", err.Error())
		return
	}
	if value == nil {
		resp.Diagnostics.AddError("Role not found", "No role with this ID exists in the account scope.")
		return
	}
	state, diags := iam.RoleState(ctx, account, value)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
