package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/spf13/cast"
)

type AllocatedIpLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type AllocatedIp struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PopLocation types.String `tfsdk:"pop_location"`
	CountryCode types.String `tfsdk:"country_code"`
}

func AllocatedIpDataSource() datasource.DataSource {
	return &allocatedIpDataSource{}
}

type allocatedIpDataSource struct {
	client *catoClientData
}

func (d *allocatedIpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocatedIp"
}

func (d *allocatedIpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of names to filter",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of allocatedIps",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
						"pop_location": schema.StringAttribute{
							Description: "Pop Location",
							Computed:    true,
						},
						"country_code": schema.StringAttribute{
							Description: "Country code",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *allocatedIpDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *allocatedIpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var allocatedIpLookup AllocatedIpLookup
	if diags := req.Config.Get(ctx, &allocatedIpLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	result, err := d.client.catov2.EntityLookup(
		ctx,
		d.client.AccountId,
		cato_models.EntityTypeAllocatedIP,
		nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	filterByName := !allocatedIpLookup.NameFilter.IsNull() && allocatedIpLookup.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range allocatedIpLookup.NameFilter.Elements() {
			// Trim any quotes if present
			valueStr := strings.Trim(value.String(), "\"")
			namesMap[valueStr] = struct{}{}
		}
	}

	attrTypes := map[string]attr.Type{
		"id":           types.StringType,
		"name":         types.StringType,
		"pop_location": types.StringType,
		"country_code": types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		ip := cast.ToString(helperFields["allocatedIp"])
		if !filterByName || contains(namesMap, ip) {
			popLocation := cast.ToString(helperFields["popLocation"])
			countryCode := cast.ToString(helperFields["countryCode"])
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":           types.StringValue(item.GetEntity().GetID()),
					"name":         types.StringValue(ip),
					"pop_location": types.StringValue(popLocation),
					"country_code": types.StringValue(countryCode),
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
			AttrTypes: attrTypes,
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	allocatedIpLookup.Items = list
	if diags := resp.State.Set(ctx, &allocatedIpLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
