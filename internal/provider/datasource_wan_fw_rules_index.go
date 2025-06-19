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

func WanRulesIndexDataSource() datasource.DataSource {
	return &wanRulesIndexDataSource{}
}

type wanRulesIndexDataSource struct {
	client *catoClientData
}

func (d *wanRulesIndexDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wanRulesIndex"
}

func (d *wanRulesIndexDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for WAN Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"rules": schema.ListNestedAttribute{
				Description: "List of WAN Policy Indexes",
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
						"action": schema.StringAttribute{
							Description: "rule action defined by api",
							Required:    false,
							Optional:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    false,
							Optional:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "WAN section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"section_id": schema.StringAttribute{
							Description: "WAN section ID housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "WAN description",
							Required:    false,
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Is the WAN rule enalbed?",
							Required:    false,
							Optional:    true,
						},
						"name": schema.StringAttribute{
							Description: "WAN rule name",
							Required:    false,
							Optional:    true,
						},
						"properties": schema.ListAttribute{
							Description: "WAN section ID housing rule",
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

func (d *wanRulesIndexDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

var WanRuleIndexObjectType = types.ObjectType{AttrTypes: WanRuleIndexAttrTypes}
var WanRuleIndexAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"action":           types.StringType,
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

type WanRuleIndexLookup struct {
	Rules types.List `tfsdk:"rules"`
}

func (d *wanRulesIndexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var wanRuleIndexLookup WanRuleIndexLookup
	ruleIndexApiData, err := d.client.catov2.PolicyWanFirewallRulesIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanFirewallRulesIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ruleIndexApiData),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	sectionIndexApiData, err := d.client.catov2.PolicyWanFirewallSectionsIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanFirewallSectionsIndex.response", map[string]interface{}{
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
	for _, v := range sectionIndexApiData.Policy.WanFirewall.Policy.Sections {
		sectionIdList[v.Section.ID] = 0
	}

	for _, item := range ruleIndexApiData.Policy.WanFirewall.Policy.Rules {
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

			tflog.Warn(ctx, "Read.SystemSectionForItem.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(item),
			})

			ruleIndexStateData, diags := types.ObjectValue(
				WanRuleIndexAttrTypes,
				map[string]attr.Value{
					"id":               types.StringValue(item.Rule.ID),
					"action":           types.StringValue(string(item.Rule.Action)),
					"index":            types.Int64Value(item.Rule.Index),
					"index_in_section": types.Int64Value(sectionIdListItem),
					"section_name":     types.StringValue(item.Rule.Section.GetName()),
					"section_id":       types.StringValue(item.Rule.Section.GetID()),
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
			AttrTypes: WanRuleIndexAttrTypes,
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	wanRuleIndexLookup.Rules = objectlist

	if diags := resp.State.Set(ctx, &wanRuleIndexLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

}
