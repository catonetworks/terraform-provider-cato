package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var roleCatalogAttrTypes = map[string]attr.Type{
	"resource": types.StringType,
	"supported_actions": types.SetType{
		ElemType: types.StringType,
	},
}
var roleCatalogObjectType = types.ObjectType{AttrTypes: roleCatalogAttrTypes}

type PermissionCatalog struct {
	ID        types.String `tfsdk:"id"`
	AccountID types.String `tfsdk:"account_id"`
	Resources types.Set    `tfsdk:"resources"`
}
type permissionCatalogDataSource struct{ roleClientConfig }

var _ datasource.DataSourceWithConfigure = &permissionCatalogDataSource{}

func NewPermissionCatalogDataSource() datasource.DataSource { return &permissionCatalogDataSource{} }
func (d *permissionCatalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_catalog"
}
func (d *permissionCatalogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req.ProviderData, &resp.Diagnostics)
}
func (d *permissionCatalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads permission resources and supported actions available when defining a custom IAM role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Account identifier for this lookup.",
				Computed:    true,
			},
			"account_id": roleDataAccountAttribute(),
			"resources": schema.SetNestedAttribute{
				Description: "Resources available in this account's permission catalog.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource": schema.StringAttribute{
							Description: "Resource identifier used in role permissions.",
							Computed:    true,
						},
						"supported_actions": schema.SetAttribute{
							Description: "Actions grantable on this resource.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}
func (d *permissionCatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PermissionCatalog
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accountID, err := effectiveRoleAccount(config.AccountID, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	client := d.getRoleClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured role client", "Configure the Cato provider before reading the permission catalog.")
		return
	}
	result, err := client.RbacRoleManagementPermissionCatalog(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read permission catalog", err.Error())
		return
	}
	catalog := result.GetRbac().GetRoleManagement().GetPermissionCatalog()
	if catalog == nil {
		resp.Diagnostics.AddError("Invalid permission-catalog response", "The API returned no permission catalog.")
		return
	}
	resources := make([]attr.Value, 0, len(catalog.GetResource()))
	for _, item := range catalog.GetResource() {
		if item == nil || item.GetResource() == "" {
			resp.Diagnostics.AddError("Invalid permission-catalog response", "The API returned an incomplete catalog resource.")
			return
		}
		actions := make([]attr.Value, 0, len(item.GetSupportedAction()))
		for _, action := range item.GetSupportedAction() {
			actions = append(actions, types.StringValue(string(action)))
		}
		supportedActions, diags := types.SetValue(types.StringType, actions)
		resp.Diagnostics.Append(diags...)
		object, diags := types.ObjectValue(roleCatalogAttrTypes, map[string]attr.Value{
			"resource":          types.StringValue(item.GetResource()),
			"supported_actions": supportedActions,
		})
		resp.Diagnostics.Append(diags...)
		resources = append(resources, object)
	}
	value, diags := types.SetValue(roleCatalogObjectType, resources)
	resp.Diagnostics.Append(diags...)
	config.Resources = value
	config.ID = types.StringValue(accountID)
	config.AccountID = types.StringValue(accountID)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	}
}
