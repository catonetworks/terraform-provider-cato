package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type roleDataSource struct{ roleClientConfig }

var _ datasource.DataSourceWithConfigure = &roleDataSource{}

func NewRoleDataSource() datasource.DataSource { return &roleDataSource{} }
func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}
func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req.ProviderData, &resp.Diagnostics)
}
func roleDataAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Account-scoped Terraform identifier, ACCOUNT_ID:ROLE_ID.",
			Computed:    true,
		},
		"account_id": schema.StringAttribute{
			Description: "Account used for this lookup.",
			Computed:    true,
		},
		"role_id": schema.StringAttribute{
			Description: "Stable API role identifier.",
			Computed:    true,
		},
		"name": schema.StringAttribute{Description: "Role name.", Computed: true},
		"description": schema.StringAttribute{
			Description: "Role description; an absent description is represented as an empty string.",
			Computed:    true,
		},
		"predefined": schema.BoolAttribute{
			Description: "Whether this role is predefined and immutable.",
			Computed:    true,
		},
		"account_type": schema.StringAttribute{
			Description: "Account type applicable to the role.",
			Computed:    true,
		},
		"is_used_on_external_access": schema.BoolAttribute{
			Description: "Whether external access currently uses the role.",
			Computed:    true,
		},
		"permissions": schema.SetNestedAttribute{
			Description: "Permissions granted by the role.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"resource": schema.StringAttribute{Description: "Permission resource identifier.", Computed: true},
					"action":   schema.StringAttribute{Description: "Granted RBAC action.", Computed: true},
				},
			},
		},
	}
}
func roleDataAccountAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Account ID. Defaults to the provider account_id.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
}
func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleDataAttributes()
	attrs["account_id"] = roleDataAccountAttribute()
	attrs["role_id"] = schema.StringAttribute{
		Description: "Role identifier to look up; both custom and predefined roles can be read.",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
	resp.Schema = schema.Schema{
		Description: "Reads one custom or predefined IAM role by identifier.",
		Attributes:  attrs,
	}
}
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Role
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accountID, err := effectiveRoleAccount(config.AccountID, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	model, diags := lookupRole(ctx, d.getRoleClient(), accountID, config.RoleID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model == nil {
		resp.Diagnostics.AddError("Role not found", "No role with this identifier exists in the selected account.")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
