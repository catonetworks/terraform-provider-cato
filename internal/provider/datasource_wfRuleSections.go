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

type wfRuleSectionLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type wfRuleSection struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func WfRuleSectionsDataSource() datasource.DataSource {
	return &wfRuleSectionsDataSource{}
}

type wfRuleSectionsDataSource struct {
	client *catoClientData
}

func (d *wfRuleSectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wfRuleSections"
}

func (d *wfRuleSectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves WAN Firewall rule sections.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of names to filter",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of wan firewall rule sections",
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

func (d *wfRuleSectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *wfRuleSectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var wfRuleSectionLookup wfRuleSectionLookup
	if diags := req.Config.Get(ctx, &wfRuleSectionLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	result, err := d.client.catov2.PolicyWanFirewallSectionsIndex(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanFirewallSectionsIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallSectionsIndex error", err.Error())
		return
	}

	filterByName := !wfRuleSectionLookup.NameFilter.IsNull() && wfRuleSectionLookup.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range wfRuleSectionLookup.NameFilter.Elements() {
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

	for _, item := range result.Policy.WanFirewall.Policy.Sections {
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

	wfRuleSectionLookup.Items = list
	if diags := resp.State.Set(ctx, &wfRuleSectionLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
