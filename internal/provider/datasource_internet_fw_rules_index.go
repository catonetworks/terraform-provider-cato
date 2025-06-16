package provider

import (
	"context"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func IfwRulesIndexDataSource() datasource.DataSource {
	return &ifwRulesIndexDataSource{}
}

type ifwRulesIndexDataSource struct {
	client *catoClientData
}

func (d *ifwRulesIndexDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ifwRulesIndex"
}

func (d *ifwRulesIndexDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for Internet Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"rules": schema.ListNestedAttribute{
				Description: "List of IFW Policy Indexes",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "rule id defined by api",
							Required:    false,
							Optional:    true,
						},
						"index": schema.Int64Attribute{
							Description: "Index value provided by API",
							Required:    false,
							Optional:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    false,
							Optional:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "IFW section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"section_id": schema.StringAttribute{
							Description: "IFW section ID housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "IFW description",
							Required:    false,
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Is the IFW rule enalbed?",
							Required:    false,
							Optional:    true,
						},
						"name": schema.StringAttribute{
							Description: "IFW rule name",
							Required:    false,
							Optional:    true,
						},
						"properties": schema.ListAttribute{
							Description: "IFW section ID housing rule",
							Required:    false,
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *ifwRulesIndexDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

var IfwRuleIndexObjectType = types.ObjectType{AttrTypes: IfwRuleIndexAttrTypes}
var IfwRuleIndexAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index":            types.Int64Type,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"section_id":       types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
	"name":             types.StringType,
	"properties": types.ListType{
		ElemType: types.StringType,
	},
}

type IfwRuleIndexLookup struct {
	Rules types.List `tfsdk:"rules"`
}

func (d *ifwRulesIndexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var ifwRuleIndexLookup IfwRuleIndexLookup
	ruleIndexApiData, err := d.client.catov2.PolicyInternetFirewallRulesIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallRulesIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ruleIndexApiData),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	sectionIndexApiData, err := d.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	var objects []attr.Value

	sectionIdList := make(map[string]int64)
	for _, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		sectionIdList[v.Section.ID] = 0
	}

	for _, item := range ruleIndexApiData.Policy.InternetFirewall.Policy.Rules {
		systemSection := false
		propertiesStringSlice := make([]string, 0)
		for _, v := range item.Properties {
			propertiesStringSlice = append(propertiesStringSlice, v.String())
			if v.String() == "SYSTEM" {
				systemSection = true
			}
		}

		if !systemSection {
			propertiesValue, diags := types.ListValueFrom(context.Background(), types.StringType, propertiesStringSlice)
			resp.Diagnostics.Append(diags...)

			sectionIdList[item.Rule.Section.ID]++
			sectionIdListItem := sectionIdList[item.Rule.Section.ID]

			ruleIndexStateData, diags := types.ObjectValue(
				IfwRuleIndexAttrTypes,
				map[string]attr.Value{
					"id":               types.StringValue(item.Rule.ID),
					"index":            types.Int64Value(item.Rule.Index),
					"index_in_section": types.Int64Value(sectionIdListItem),
					"section_name":     types.StringValue(item.Rule.Section.Name),
					"section_id":       types.StringValue(item.Rule.Section.ID),
					"description":      types.StringValue(item.Rule.Description),
					"enabled":          types.BoolValue(item.Rule.Enabled),
					"name":             types.StringValue(item.Rule.Name),
					"properties":       propertiesValue,
				},
			)
			resp.Diagnostics.Append(diags...)

			objects = append(objects, ruleIndexStateData)
		}

	}

	objectlist, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwRuleIndexAttrTypes,
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	ifwRuleIndexLookup.Rules = objectlist

	if diags := resp.State.Set(ctx, &ifwRuleIndexLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

	// rootMap := make(map[string]string)

	// tflog.Warn(context.Background(), jsonTxt)
}
