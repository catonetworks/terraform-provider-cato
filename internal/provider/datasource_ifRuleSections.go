package provider

import (
	"context"
	"strings"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

type ifRuleSectionLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type ifRuleSection struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func IfRuleSectionsDataSource() datasource.DataSource {
	return &ifRuleSectionsDataSource{}
}

type ifRuleSectionsDataSource struct {
	client *catoClientData
}

func (d *ifRuleSectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ifRuleSections"
}

func (d *ifRuleSectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves Internet Firewall rule sections.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of names to filter",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of internet firewall rule sections",
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
					},
				},
			},
		},
	}
}

func (d *ifRuleSectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *ifRuleSectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var ifRuleSectionLookup ifRuleSectionLookup
	if diags := req.Config.Get(ctx, &ifRuleSectionLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	result, err := d.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallSectionsIndex error", err.Error())
		return
	}

	filterByName := !ifRuleSectionLookup.NameFilter.IsNull() && ifRuleSectionLookup.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range ifRuleSectionLookup.NameFilter.Elements() {
			// Trim any quotes if present
			valueStr := strings.Trim(value.String(), "\"")
			namesMap[valueStr] = struct{}{}
		}
	}

	attrTypes := map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.Policy.InternetFirewall.Policy.Sections {
		section_name := cast.ToString(item.Section.Name)
		section_id := cast.ToString(item.Section.ID)
		if !filterByName || contains(namesMap, section_name) {
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(section_id),
					"name": types.StringValue(section_name),
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

	ifRuleSectionLookup.Items = list
	if diags := resp.State.Set(ctx, &ifRuleSectionLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
