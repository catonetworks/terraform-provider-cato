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

type HostLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	IPFilter   types.List `tfsdk:"ip_filter"`
	Items      types.List `tfsdk:"items"`
}

type Host struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	IPAddr  types.String `tfsdk:"ip"`
	Network types.String `tfsdk:"net"`
	SiteID  types.String `tfsdk:"site_id"`
}

// terraform field names
const (
	tfNameFilter = "name_filter"
	tfIPFilter   = "ip_filter"
	tfItems      = "items"
	tfID         = "id"
	tfName       = "name"
	tfIP         = "ip"
	tfNet        = "net"
	tfSiteID     = "site_id"
)

var attrTypes = map[string]attr.Type{
	tfID:     types.StringType,
	tfName:   types.StringType,
	tfIP:     types.StringType,
	tfNet:    types.StringType,
	tfSiteID: types.StringType,
}

type hostFilter struct {
	namesMap map[string]struct{}
	ipsMap   map[string]struct{}
}

func HostDataSource() datasource.DataSource {
	return &hostDataSource{}
}

type hostDataSource struct {
	client *catoClientData
}

func (d *hostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *hostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves static host reservations.",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of host names to filter",
				Optional:    true,
			},
			tfIPFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of host IPs to filter",
				Optional:    true,
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of hosts",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						tfID: schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						tfName: schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
						tfIP: schema.StringAttribute{
							Description: "IP address",
							Computed:    true,
						},
						tfNet: schema.StringAttribute{
							Description: "Name of the network (port, iterface)",
							Computed:    true,
						},
						tfSiteID: schema.StringAttribute{
							Description: "Site ID",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *hostDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *hostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var hostLookup HostLookup
	if diags := req.Config.Get(ctx, &hostLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	zeroInt64 := int64(0)
	result, err := d.client.catov2.EntityLookup(ctx, d.client.AccountId, cato_models.EntityTypeHost, &zeroInt64, nil, nil, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	filter := makeFilter(&hostLookup)
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		entity := item.GetEntity()
		ip := cast.ToString(helperFields["ip"])
		name := cast.ToString(entity.Name)
		if filter.accept(ip, name) {
			siteID := cast.ToString(helperFields["siteId"])
			net := cast.ToString(helperFields["interfaceName"])
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					tfID:     types.StringValue(entity.ID),
					tfName:   types.StringValue(name),
					tfIP:     types.StringValue(ip),
					tfNet:    types.StringValue(net),
					tfSiteID: types.StringValue(siteID),
				},
			)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			objects = append(objects, obj)
		}
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	hostLookup.Items = list
	if diags := resp.State.Set(ctx, &hostLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// makeFilter prepares lookup maps for names and IPs;
// if IsNull() then the respective map is nil
func makeFilter(hostLookup *HostLookup) *hostFilter {
	var f hostFilter

	if !hostLookup.NameFilter.IsNull() {
		f.namesMap = make(map[string]struct{})
		for _, value := range hostLookup.NameFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.namesMap[valueStr] = struct{}{}
		}
	}
	if !hostLookup.IPFilter.IsNull() {
		f.ipsMap = make(map[string]struct{})
		for _, value := range hostLookup.IPFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.ipsMap[valueStr] = struct{}{}
		}
	}
	return &f
}

// accept returns true if the item should be returned
func (f *hostFilter) accept(ip, name string) bool {
	if f.ipsMap == nil {
		if f.namesMap == nil {
			return true // no filters applied
		}
		if _, ok := f.namesMap[name]; ok {
			return true // host name matches the filter
		}
		return false // host name did not match
	}

	if _, ok := f.ipsMap[ip]; ok {
		return true // host IP matches the filter
	}
	if f.namesMap == nil {
		return false // host IP did not match and names filter not used
	}
	if _, ok := f.namesMap[name]; ok {
		return true // host name matches the filter
	}
	return false // nor IP nor name matched
}
