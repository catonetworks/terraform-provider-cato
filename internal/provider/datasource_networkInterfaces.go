package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

const socketMappJson = `{
	"SOCKET_X1600": {
		"1": "WAN_1",
		"2": "WAN_2",
		"3": "LAN_1",
		"4": "LAN_2",
		"5": "INT_1",
		"6": "INT_2",
		"7": "INT_3",
		"8": "INT_4"
	},
	"SOCKET_X1600_LTE": {
		"1": "WAN_1",
		"2": "WAN_2",
		"3": "LAN_1",
		"4": "LAN_2",
		"5": "INT_1",
		"6": "INT_2",
		"7": "INT_3",
		"8": "INT_4",
		"LTE": "LTE"
	},
	"SOCKET_X1700": {
		"1": "WAN_1",
		"2": "WAN_2",
		"3": "LAN_1",
		"4": "LAN_2",
		"5": "INT_1",
		"6": "INT_2",
		"7": "INT_3",
		"8": "INT_4",
		"9": "INT_5",
		"10": "INT_6",
		"11": "INT_7",
		"12": "INT_8",
		"13": "INT_9",
		"14": "INT_10",
		"15": "INT_11",
		"16": "INT_12"
	}
}`

type NetworkInterfaceLookup struct {
	SiteID types.String `tfsdk:"site_id"`
	Items  types.List   `tfsdk:"items"`
}

type NetworkInterface struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SiteID            types.String `tfsdk:"site_id"`
	SiteName          types.String `tfsdk:"site_name"`
	Subnet            types.String `tfsdk:"subnet"`
	SocketInterfaceId types.String `tfsdk:"socket_interface_id"`
}

func NetworkInterfacesDataSource() datasource.DataSource {
	return &networkInterfacesDataSource{}
}

type networkInterfacesDataSource struct {
	client *catoClientData
}

func (d *networkInterfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networkInterfaces"
}

func (d *networkInterfacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "ID of the site to retrieve network interfaces for",
				Required:    false,
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of network interfaces",
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
						"site_id": schema.StringAttribute{
							Description: "Site ID",
							Computed:    true,
						},
						"site_name": schema.StringAttribute{
							Description: "Site Name",
							Computed:    true,
						},
						"subnet": schema.StringAttribute{
							Description: "Subnet",
							Computed:    true,
						},
						"socket_interface_index": schema.StringAttribute{
							Description: "Socket interface index",
							Computed:    true,
						},
						"socket_interface_id": schema.StringAttribute{
							Description: "Unique name of the socket interface",
							Computed:    true,
						},
						"dest_type": schema.StringAttribute{
							Description: "Interface destination type",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *networkInterfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *networkInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var networkInterfaceLookup NetworkInterfaceLookup
	if diags := req.Config.Get(ctx, &networkInterfaceLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Planned to use this to filter results by siteID, but have to do so
	// on the client side as invlalid siteID returns unintuitive error..
	// parent := &cato_models.EntityInput{
	// 	ID:   networkInterfaceLookup.SiteID.ValueString(),
	// 	Type: cato_models.EntityTypeSite,
	// }

	// Unmarshal socketMapping into a nested map
	var socketMap map[string]map[string]string
	err := json.Unmarshal([]byte(socketMappJson), &socketMap)
	if err != nil {
		panic(err)
	}

	accountSnapshotSite, err := d.client.catov2.AccountSnapshot(ctx, []string{}, nil, &d.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}
	// Create a mapping to store map[siteID][interface name]{interface_id, index}
	type InterfaceConfig struct {
		InterfaceID string `json:"interface_id"`
		Index       string `json:"index"`
		DestType    string `json:"dest_type"`
	}
	ifaceMap := make(map[string]map[string]InterfaceConfig)
	for _, site := range accountSnapshotSite.AccountSnapshot.GetSites() {
		siteID := site.GetID()
		if networkInterfaceLookup.SiteID.IsNull() || (!networkInterfaceLookup.SiteID.IsNull() && *siteID == networkInterfaceLookup.SiteID.ValueString()) {
			connType := site.InfoSiteSnapshot.ConnType
			if socketConf, ok := socketMap[connType.String()]; ok {
				if _, exists := ifaceMap[*siteID]; !exists {
					ifaceMap[*siteID] = make(map[string]InterfaceConfig)
				}
				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					ifaceMap[*siteID][*iface.Name] = InterfaceConfig{
						InterfaceID: socketConf[iface.ID],
						Index:       iface.ID,
						DestType:    *iface.DestType,
					}
				}
			} else {
				fmt.Println(fmt.Sprintf("%v", connType) + " not found")
			}
		}
	}

	tflog.Warn(ctx, "ifaceMap '"+fmt.Sprintf("%v", ifaceMap)+"'")

	result, err := d.client.catov2.EntityLookup(
		ctx,
		d.client.AccountId,
		cato_models.EntityTypeNetworkInterface,
		nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	attrTypes := map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"site_id":                types.StringType,
		"site_name":              types.StringType,
		"subnet":                 types.StringType,
		"socket_interface_index": types.StringType,
		"socket_interface_id":    types.StringType,
		"dest_type":              types.StringType,
	}
	var objects []attr.Value

	for _, item := range result.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		interfaceName := cast.ToString(helperFields["interfaceName"])
		siteID := cast.ToString(helperFields["siteId"])
		if networkInterfaceLookup.SiteID.IsNull() || (!networkInterfaceLookup.SiteID.IsNull() && siteID == networkInterfaceLookup.SiteID.ValueString()) {
			interfaceId := types.StringNull()
			interfaceIndex := types.StringNull()
			interfaceDestType := types.StringNull()
			if _, exists := ifaceMap[siteID]; exists {
				if _, exists := ifaceMap[siteID][interfaceName]; exists {
					interfaceId = types.StringValue(ifaceMap[siteID][interfaceName].InterfaceID)
					interfaceIndex = types.StringValue(ifaceMap[siteID][interfaceName].Index)
					interfaceDestType = types.StringValue(ifaceMap[siteID][interfaceName].DestType)
				}
			}
			obj, diags := types.ObjectValue(
				attrTypes,
				map[string]attr.Value{
					"id":                     types.StringValue(item.GetEntity().GetID()),
					"name":                   types.StringValue(interfaceName),
					"site_id":                types.StringValue(siteID),
					"site_name":              types.StringValue(cast.ToString(helperFields["siteName"])),
					"subnet":                 types.StringValue(cast.ToString(helperFields["subnet"])),
					"socket_interface_index": interfaceIndex,
					"socket_interface_id":    interfaceId,
					"dest_type":              interfaceDestType,
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

	networkInterfaceLookup.Items = list
	if diags := resp.State.Set(ctx, &networkInterfaceLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
