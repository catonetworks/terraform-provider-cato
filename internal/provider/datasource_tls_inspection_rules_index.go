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

func TlsRulesIndexDataSource() datasource.DataSource {
	return &tlsRulesIndexDataSource{}
}

type tlsRulesIndexDataSource struct {
	client *catoClientData
}

func (d *tlsRulesIndexDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tlsRulesIndex"
}

func (d *tlsRulesIndexDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for TLS Inspection Rules.",
		Attributes: map[string]schema.Attribute{
			"rules": schema.ListNestedAttribute{
				Description: "List of TLS Inspection Policy Indexes",
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
							Description: "TLS Inspection section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"section_id": schema.StringAttribute{
							Description: "TLS Inspection section ID housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "TLS Inspection rule description",
							Required:    false,
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Is the TLS Inspection rule enabled?",
							Required:    false,
							Optional:    true,
						},
						"name": schema.StringAttribute{
							Description: "TLS Inspection rule name",
							Required:    false,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (d *tlsRulesIndexDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

var TlsRuleIndexObjectType = types.ObjectType{AttrTypes: TlsRuleIndexAttrTypes}
var TlsRuleIndexAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"action":           types.StringType,
	"index":            types.Int64Type,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"section_id":       types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
	"name":             types.StringType,
}

type TlsRuleIndexLookup struct {
	Rules types.List `tfsdk:"rules"`
}

func (d *tlsRulesIndexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var tlsRuleIndexLookup TlsRuleIndexLookup
	ruleIndexApiData, err := d.client.catov2.Tlsinspectpolicy(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.Tlsinspectpolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ruleIndexApiData),
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
	for _, v := range ruleIndexApiData.Policy.TLSInspect.Policy.Sections {
		sectionIdList[v.Section.ID] = 0
	}

	for _, item := range ruleIndexApiData.Policy.TLSInspect.Policy.Rules {
		sectionIdList[item.Rule.Section.ID]++
		sectionIdListItem := sectionIdList[item.Rule.Section.ID]

		tflog.Warn(ctx, "Read.TLSRuleItem.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(item),
		})

		ruleIndexStateData, diags := types.ObjectValue(
			TlsRuleIndexAttrTypes,
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
			},
		)
		resp.Diagnostics.Append(diags...)

		objects = append(objects, ruleIndexStateData)
	}

	objectlist, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: TlsRuleIndexAttrTypes,
		},
		objects,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tlsRuleIndexLookup.Rules = objectlist

	if diags := resp.State.Set(ctx, &tlsRuleIndexLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

}
