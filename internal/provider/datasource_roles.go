package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const rolePageSize = 100

type RoleList struct {
	ID            types.String `tfsdk:"id"`
	AccountID     types.String `tfsdk:"account_id"`
	Filter        types.Object `tfsdk:"filter"`
	SortDirection types.String `tfsdk:"sort_direction"`
	Items         types.List   `tfsdk:"items"`
	Total         types.Int64  `tfsdk:"total"`
}
type rolesDataSource struct{ roleClientConfig }

var _ datasource.DataSourceWithConfigure = &rolesDataSource{}

func NewRolesDataSource() datasource.DataSource { return &rolesDataSource{} }
func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}
func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req.ProviderData, &resp.Diagnostics)
}
func roleStringFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Filter operators, combined using the API's filtering semantics.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"eq":  schema.StringAttribute{Description: "Equal to this value.", Optional: true},
			"neq": schema.StringAttribute{Description: "Not equal to this value.", Optional: true},
			"in": schema.SetAttribute{
				Description: "Match any of these values.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"nin": schema.SetAttribute{
				Description: "Exclude these values.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}
func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all matching custom and predefined IAM roles, automatically fetching every API page.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Description: "Account identifier for this lookup.", Computed: true},
			"account_id": roleDataAccountAttribute(),
			"filter": schema.SingleNestedAttribute{
				Description: "Optional role filters.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"id":   roleStringFilterSchema(),
					"name": roleStringFilterSchema(),
					"predefined": schema.SingleNestedAttribute{
						Description: "Filter predefined or custom roles.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"eq": schema.BoolAttribute{
								Description: "Equal to this predefined status.",
								Optional:    true,
							},
							"neq": schema.BoolAttribute{
								Description: "Exclude this predefined status.",
								Optional:    true,
							},
						},
					},
				},
			},
			"sort_direction": schema.StringAttribute{
				Description: "Sort roles by name in ASC or DESC order. Defaults to ASC.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ASC", "DESC"),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "All matching roles in API sort order.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: roleDataAttributes(),
				},
			},
			"total": schema.Int64Attribute{Description: "Number of matching roles.", Computed: true},
		},
	}
}
func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RoleList
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accountID, err := effectiveRoleAccount(config.AccountID, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	if config.SortDirection.IsNull() {
		config.SortDirection = types.StringValue("ASC")
	}
	input, diags := roleListInput(ctx, config.Filter, config.SortDirection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	client := d.getRoleClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured role client", "Configure the Cato provider before listing roles.")
		return
	}
	items, total, diags := readAllRoles(ctx, client, accountID, input)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Items, diags = types.ListValue(roleObjectType, items)
	resp.Diagnostics.Append(diags...)
	config.ID = types.StringValue(accountID)
	config.AccountID = types.StringValue(accountID)
	config.Total = types.Int64Value(total)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	}
}

func readAllRoles(ctx context.Context, client RoleManagementClient, accountID string,
	input cato_models.RoleManagementRoleListInput,
) ([]attr.Value, int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]attr.Value, 0)
	seen := make(map[string]struct{})
	var total int64
	for {
		result, apiErr := client.RbacRoleManagementRoleList(ctx, accountID, input)
		if apiErr != nil {
			diags.AddError("Failed to list roles", apiErr.Error())
			return nil, 0, diags
		}
		payload := result.GetRbac().GetRoleManagement().GetRoleList()
		if payload == nil || payload.GetPaging() == nil {
			diags.AddError("Invalid role-list response", "The API returned no role list or paging information.")
			return nil, 0, diags
		}
		total = payload.GetPaging().GetTotal()
		if total < 0 {
			diags.AddError("Invalid role-list response", "The API returned a negative total.")
			return nil, 0, diags
		}
		if len(payload.GetItems()) == 0 && int64(len(items)) < total {
			diags.AddError("Role pagination made no progress", "The API returned an empty page before all matching roles were read.")
			return nil, 0, diags
		}
		for _, value := range payload.GetItems() {
			if value == nil {
				diags.AddError("Invalid role-list response", "The API returned a null role.")
				return nil, 0, diags
			}
			if _, duplicate := seen[value.GetID()]; duplicate {
				diags.AddError("Role pagination made no progress", "The API repeated a role identifier. Retry after concurrent changes finish.")
				return nil, 0, diags
			}
			model, dg := roleFromAPI(ctx, accountID, value)
			diags.Append(dg...)
			if diags.HasError() {
				return nil, 0, diags
			}
			object, dg := roleObject(ctx, model)
			diags.Append(dg...)
			if diags.HasError() {
				return nil, 0, diags
			}
			items = append(items, object)
			seen[value.GetID()] = struct{}{}
		}
		if int64(len(items)) >= total {
			break
		}
		input.Paging.From += int64(len(payload.GetItems()))
	}
	return items, total, diags
}
