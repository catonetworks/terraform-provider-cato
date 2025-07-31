package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

const socketMappJson = `{
	"SOCKET_X1500": {
		"WAN1": "WAN 01",
		"WAN2": "WAN 02",
		"LAN1": "LAN 01",
		"LAN2": "LAN 02"
	},
	"SOCKET_X1600": {
		"1": "WAN",
		"2": "INT_2",
		"3": "INT_3",
		"4": "INT_4",
		"5": "LAN",
		"6": "INT_6",
		"7": "INT_7",
		"8": "INT_8"
	},
	"SOCKET_X1600_LTE": {
		"1": "WAN",
		"2": "INT_2",
		"3": "INT_3",
		"4": "INT_4",
		"5": "LAN",
		"6": "INT_6",
		"7": "INT_7",
		"8": "INT_8",
		"LTE": "LTE"
	},
	"SOCKET_X1700": {
		"1": "WAN",
		"2": "INT_2",
		"3": "LAN",
		"4": "INT_4",
		"5": "INT_5",
		"6": "INT_6",
		"7": "INT_7",
		"8": "INT_8",
		"9": "INT_9",
		"10": "INT_10",
		"11": "INT_11",
		"12": "INT_12",
		"13": "INT_13",
		"14": "INT_14",
		"15": "INT_15",
		"16": "INT_16"
	}
}`

type NetworkInterfaceLookup struct {
	SiteID                types.String `tfsdk:"site_id"`
	NetworkInterfaceIndex types.String `tfsdk:"network_interface_index"`
	Items                 types.List   `tfsdk:"items"`
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
			"network_interface_index": schema.StringAttribute{
				Description: "Index of the interface to retrieve, example: INT_1, INT_2, etc",
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
	var networkInterfacesDataSource NetworkInterfaceLookup
	if diags := req.Config.Get(ctx, &networkInterfacesDataSource); diags.HasError() {
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

	connType := ""
	ifaceAttrTypes := map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"site_id":                types.StringType,
		"site_name":              types.StringType,
		"subnet":                 types.StringType,
		"socket_interface_index": types.StringType,
		"socket_interface_id":    types.StringType,
		"dest_type":              types.StringType,
	}
	var ifaceList []attr.Value

	// Create a mapping to store map[siteID][interface name]{interface_id, index}
	type InterfaceConfig struct {
		InterfaceID string `json:"interface_id"`
		Index       string `json:"index"`
		DestType    string `json:"dest_type"`
		Name        string `json:"name"`
	}
	ifaceMap := make(map[string]map[string]InterfaceConfig)

	tflog.Warn(ctx, "ifaceMap '"+fmt.Sprintf("%v", ifaceMap)+"'")
	zeroInt64 := int64(0)
	entityLookupResponse, err := d.client.catov2.EntityLookup(ctx, d.client.AccountId, cato_models.EntityTypeNetworkInterface, &zeroInt64, nil, nil, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(entityLookupResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API EntityLookup error", err.Error())
		return
	}

	accountSnapshotSite, err := d.client.catov2.AccountSnapshot(ctx, []string{}, nil, &d.client.AccountId)
	tflog.Debug(ctx, "Read.AccountSnapshot.response", map[string]interface{}{
		"response":       utils.InterfaceToJSONString(accountSnapshotSite),
		"interfaceIndex": utils.InterfaceToJSONString(networkInterfacesDataSource.NetworkInterfaceIndex.ValueString()),
		"siteID":         utils.InterfaceToJSONString(networkInterfacesDataSource.SiteID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	for _, site := range accountSnapshotSite.AccountSnapshot.GetSites() {
		siteID := site.GetID()
		if networkInterfacesDataSource.SiteID.IsNull() || (!networkInterfacesDataSource.SiteID.IsNull() && *siteID == networkInterfacesDataSource.SiteID.ValueString()) {
			connType = site.InfoSiteSnapshot.ConnType.String()
			if socketConf, ok := socketMap[connType]; ok {
				tflog.Debug(ctx, "socketConf", map[string]interface{}{
					"socketConf": utils.InterfaceToJSONString(socketConf),
				})
				if _, exists := ifaceMap[*siteID]; !exists {
					ifaceMap[*siteID] = make(map[string]InterfaceConfig)
				}

				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					fmt.Println("networkInterfacesDataSource.NetworkInterfaceIndex.ValueString() " + fmt.Sprintf("%v", networkInterfacesDataSource.NetworkInterfaceIndex.ValueString()))
					fmt.Println("iface.Id " + fmt.Sprintf("%v", iface.ID))
					fmt.Println("*iface.Name " + fmt.Sprintf("%v", *iface.Name))
					curInterfaceIndex := getInterfaceIndexByConnType(networkInterfacesDataSource.NetworkInterfaceIndex.ValueString(), connType)
					if (networkInterfacesDataSource.NetworkInterfaceIndex.ValueString() == "LAN" ||
						networkInterfacesDataSource.NetworkInterfaceIndex.ValueString() == "LAN1") &&
						"INT_"+iface.ID == curInterfaceIndex {
						fmt.Println("NetworkInterfaceIndex==LAN and curInterfaceIndex=" + fmt.Sprintf("%v", curInterfaceIndex))
						ifaceMap[*siteID][curInterfaceIndex] = InterfaceConfig{
							InterfaceID: socketConf[curInterfaceIndex],
							Index:       curInterfaceIndex,
							DestType:    *iface.DestType + " is LAN",
							Name:        socketConf[iface.ID],
							// Name:        *iface.Name,
						}
					} else if (networkInterfacesDataSource.NetworkInterfaceIndex.ValueString() == "LAN" ||
						networkInterfacesDataSource.NetworkInterfaceIndex.ValueString() == "LAN1") &&
						socketConf[iface.ID] != "" {
						ifaceMap[*siteID]["INT_"+iface.ID] = InterfaceConfig{
							InterfaceID: socketConf["INT_"+iface.ID],
							Index:       iface.ID,
							DestType:    *iface.DestType + " not LAN",
							Name:        socketConf[iface.ID],
							// Name:        *iface.Name,
						}
					}
				}
			} else {
				fmt.Println(fmt.Sprintf("%v", connType) + " not found")
			}
		}
	}

	tflog.Debug(ctx, "ifaceMap", map[string]interface{}{
		"ifaceMap": utils.InterfaceToJSONString(ifaceMap),
	})

	for _, item := range entityLookupResponse.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		// entLookupinterfaceID := cast.ToString(helperFields["interfaceName"])
		entLookupinterfaceID := getInterfaceIndexByConnType(cast.ToString(helperFields["interfaceName"]), connType)
		siteID := cast.ToString(helperFields["siteId"])
		tflog.Warn(ctx, "networkInterfaceLookup.SiteID '"+fmt.Sprintf("%v", networkInterfacesDataSource.SiteID.ValueString())+"'")
		tflog.Warn(ctx, "siteID '"+fmt.Sprintf("%v", siteID)+"'")
		if networkInterfacesDataSource.SiteID.IsNull() || (!networkInterfacesDataSource.SiteID.IsNull() && siteID == networkInterfacesDataSource.SiteID.ValueString()) {
			curInterfaceIndex := getInterfaceIndexByConnType(networkInterfacesDataSource.NetworkInterfaceIndex.ValueString(), connType)
			tflog.Warn(ctx, "entLookupinterfaceID '"+fmt.Sprintf("%v", entLookupinterfaceID)+"'")
			tflog.Warn(ctx, "networkInterfacesDataSource.NetworkInterfaceName.ValueString() '"+fmt.Sprintf("%v", curInterfaceIndex)+"'")
			if networkInterfacesDataSource.NetworkInterfaceIndex.IsNull() ||
				(!networkInterfacesDataSource.NetworkInterfaceIndex.IsNull() &&
					entLookupinterfaceID == curInterfaceIndex) ||
				(!networkInterfacesDataSource.NetworkInterfaceIndex.IsNull() &&
					ifaceMap[siteID][curInterfaceIndex].Name == entLookupinterfaceID) {
				interfaceDestType := types.StringNull()
				interfaceIndex := types.StringNull()
				if _, exists := ifaceMap[siteID]; exists {
					if _, exists := ifaceMap[siteID][entLookupinterfaceID]; exists {
						interfaceDestType = types.StringValue(ifaceMap[siteID][entLookupinterfaceID].DestType)
						interfaceIndex = types.StringValue(ifaceMap[siteID][entLookupinterfaceID].Index)
					}
				}
				ifaceObj, diags := types.ObjectValue(
					ifaceAttrTypes,
					map[string]attr.Value{
						"id":                     types.StringValue(item.GetEntity().GetID()),
						"name":                   types.StringValue(ifaceMap[siteID][entLookupinterfaceID].Name),
						"site_id":                types.StringValue(siteID),
						"site_name":              types.StringValue(cast.ToString(helperFields["siteName"])),
						"subnet":                 types.StringValue(cast.ToString(helperFields["subnet"])),
						"socket_interface_index": interfaceIndex,
						"socket_interface_id":    types.StringValue(entLookupinterfaceID),
						"dest_type":              interfaceDestType,
					},
				)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				ifaceList = append(ifaceList, ifaceObj)
			}
		}
	}

	list, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: ifaceAttrTypes,
		},
		ifaceList,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	networkInterfacesDataSource.Items = list
	if diags := resp.State.Set(ctx, &networkInterfacesDataSource); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func getInterfaceIndexByConnType(interfaceName, connType string) string {
	curInterfaceIndex := interfaceName
	if interfaceName == "LAN1" && connType == "SOCKET_X1500" {
		curInterfaceIndex = "LAN 01"
	} else if interfaceName == "LAN" {
		if connType == "SOCKET_X1600" || connType == "SOCKET_X1600_LTE" {
			curInterfaceIndex = "INT_5"
		} else if connType == "SOCKET_X1700" {
			curInterfaceIndex = "INT_3"
		}
	}
	return curInterfaceIndex
}

// func getInterfaceIndex(interfaceName, connType string) string {
// 	curInterfaceIndex := interfaceName
// 	if interfaceName == "LAN1" && connType == "SOCKET_X1500" {
// 		curInterfaceIndex = "LAN 01"
// 	} else if interfaceName == "LAN" {
// 		if connType == "SOCKET_X1600" || connType == "SOCKET_X1600_LTE" {
// 			curInterfaceIndex = "INT_5"
// 		} else if connType == "SOCKET_X1700" {
// 			curInterfaceIndex = "INT_3"
// 		}
// 	}
// 	return curInterfaceIndex
// }
