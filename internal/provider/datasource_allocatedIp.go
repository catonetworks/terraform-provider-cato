package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

type AllocatedIpLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type AllocatedIp struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	CountryCode types.String `tfsdk:"cc"`
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
		Description: "Retrieves account allocated IPs.",
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
						"location": schema.StringAttribute{
							Description: "Pop Location",
							Computed:    true,
						},
						"cc": schema.StringAttribute{
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

	zeroInt64 := int64(0)
	result, err := d.client.catov2.EntityLookup(ctx, d.client.AccountId, cato_models.EntityTypeAllocatedIP, &zeroInt64, nil, nil, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
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
		"id":       types.StringType,
		"name":     types.StringType,
		"location": types.StringType,
		"cc":       types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		ip := cast.ToString(helperFields["allocatedIp"])
		if !filterByName || contains(namesMap, ip) {
			location := cast.ToString(helperFields["popLocation"])
			cc := cast.ToString(helperFields["countryCode"])
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":       types.StringValue(item.GetEntity().GetID()),
					"name":     types.StringValue(ip),
					"location": types.StringValue(location),
					"cc":       types.StringValue(cc),
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
