package provider

import (
	"context"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EntityLookup struct {
	Type  types.String `tfsdk:"type"`
	Names types.List   `tfsdk:"names"`
	Items types.List   `tfsdk:"items"`
}

type EntityLookupItem struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func EntityLookupDataSource() datasource.DataSource {
	return &entityLookupDataSource{}
}

type entityLookupDataSource struct {
	client *catoClientData
}

func (d *entityLookupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitylookup"
}

func (d *entityLookupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Entity type",
				Required:    true,
			},
			"names": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Names filter",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of entities",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Entity id",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Entity name",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *entityLookupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *entityLookupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var entityLookup EntityLookup
	if diags := req.Config.Get(ctx, &entityLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	entityType := entityLookup.Type.ValueString()
	result, err := d.client.catov2.EntityLookup(
		ctx,
		d.client.AccountId,
		cato_models.EntityType(entityType),
		nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	filterByName := !entityLookup.Names.IsNull() && entityLookup.Names.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range entityLookup.Names.Elements() {
			namesMap[value.String()] = struct{}{}
		}
	}

	var objects []attr.Value
	for _, item := range result.GetEntityLookup().GetItems() {
		if !filterByName || contains(namesMap, *item.Entity.Name) {
			obj, diags := types.ObjectValue(
				map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
				map[string]attr.Value{
					"id":   types.StringValue(item.Entity.ID),
					"name": types.StringValue(extractValue(entityType, *item.Entity.Name)),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}
	}

	list, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":   types.StringType,
				"name": types.StringType,
			},
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	entityLookup.Items = list
	if diags := resp.State.Set(ctx, &entityLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
