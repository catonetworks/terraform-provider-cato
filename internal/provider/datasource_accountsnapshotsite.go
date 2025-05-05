package provider

import (
	"context"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &accountSnapshotSiteDataSource{}
	_ datasource.DataSourceWithConfigure = &accountSnapshotSiteDataSource{}
)

func NewAccountSnapshotSiteDataSource() datasource.DataSource {
	return &accountSnapshotSiteDataSource{}
}

type accountSnapshotSiteDataSource struct {
	client *catoClientData
}

type SiteSnapshot struct {
	Id *string `tfsdk:"id"`
	// ProtoId            *int             `tfsdk:"protoId"`
	// ConnectivityStatus *string          `tfsdk:"connectivityStatus"`
	// HaStatus           *HaStatus        `tfsdk:"haStatus"`
	// OperationalStatus  *string          `tfsdk:"operationalStatus"`
	// LastConnected      *string          `tfsdk:"lastConnected"`
	// ConnectedSince     *string          `tfsdk:"connectedSince"`
	// PopName            *string          `tfsdk:"popName"`
	// Devices            []DeviceSnapshot `tfsdk:"devices"`
	Info *SiteInfo `tfsdk:"info"`
	// HostCount          *int64           `tfsdk:"hostCount"`
	// AltWanStatus       *string          `tfsdk:"altWanStatus"`
}

// type HaStatus struct {
// 	Readiness       *string `tfsdk:"readiness"`
// 	WanConnectivity *string `tfsdk:"wanConnectivity"`
// 	Keepalive       *string `tfsdk:"keepalive"`
// 	SocketVersion   *string `tfsdk:"socketVersion"`
// }

// type DeviceSnapshot struct {
// 	Id                  *string              `tfsdk:"id"`
// 	Name                *string              `tfsdk:"name"`
// 	Identifier          *string              `tfsdk:"identifier"`
// 	Connected           *bool                `tfsdk:"connected"`
// 	HaRole              *string              `tfsdk:"haRole"`
// 	Interfaces          []InterfaceSnapshot  `tfsdk:"interfaces"`
// 	LastConnected       *string              `tfsdk:"lastConnected"`
// 	LastDuration        *int64               `tfsdk:"lastDuration"`
// 	ConnectedSince      *string              `tfsdk:"connectedSince"`
// 	LastPopID           *int64               `tfsdk:"lastPopID"`
// 	LastPopName         *string              `tfsdk:"lastPopName"`
// 	RecentConnections   []RecentConnection   `tfsdk:"recentConnections"`
// 	Type                *string              `tfsdk:"type"`
// 	SocketInfo          *SocketInfo          `tfsdk:"socketInfo"`
// 	InterfacesLinkState []InterfaceLinkState `tfsdk:"interfacesLinkState"`
// 	OsType              *string              `tfsdk:"osType"`
// 	OsVersion           *string              `tfsdk:"osVersion"`
// 	Version             *string              `tfsdk:"version"`
// 	VersionNumber       *int64               `tfsdk:"versionNumber"`
// 	ReleaseGroup        *string              `tfsdk:"releaseGroup"`
// 	MfaExpirationTime   *int64               `tfsdk:"mfaExpirationTime"`
// 	MfaCreationTime     *int64               `tfsdk:"mfaCreationTime"`
// 	InternalIP          *string              `tfsdk:"internalIP"`
// }

// type InterfaceSnapshot struct {
// 	Connected              *bool              `tfsdk:"connected"`
// 	Id                     *string            `tfsdk:"id"`
// 	Name                   *string            `tfsdk:"name"`
// 	PhysicalPort           *int64             `tfsdk:"physicalPort"`
// 	NaturalOrder           *int64             `tfsdk:"naturalOrder"`
// 	PopName                *string            `tfsdk:"popName"`
// 	PreviousPopID          *int64             `tfsdk:"previousPopID"`
// 	PreviousPopName        *string            `tfsdk:"previousPopName"`
// 	TunnelConnectionReason *string            `tfsdk:"tunnelConnectionReason"`
// 	TunnelUptime           *int64             `tfsdk:"tunnelUptime"`
// 	TunnelRemoteIP         *string            `tfsdk:"tunnelRemoteIP"`
// 	TunnelRemoteIPInfo     *IPInfo            `tfsdk:"tunnelRemoteIPInfo"`
// 	Type                   *string            `tfsdk:"type"`
// 	Info                   *InterfaceInfo     `tfsdk:"info"`
// 	CellularInterfaceInfo  *CellularInterface `tfsdk:"cellularInterfaceInfo"`
// }

// type CellularInterface struct {
// 	NetworkType         *string `tfsdk:"networkType"`
// 	SimSlotId           *int64  `tfsdk:"simSlotId"`
// 	ModemStatus         *string `tfsdk:"modemStatus"`
// 	IsModemConnected    bool    `tfsdk:"isModemConnected"`
// 	Iccid               *string `tfsdk:"iccid"`
// 	Imei                *string `tfsdk:"imei"`
// 	OperatorName        *string `tfsdk:"operatorName"`
// 	IsModemSuspended    bool    `tfsdk:"isModemSuspended"`
// 	Apn                 *string `tfsdk:"apn"`
// 	ApnSelectionMethod  *string `tfsdk:"apnSelectionMethod"`
// 	SignalStrength      *string `tfsdk:"signalStrength"`
// 	IsRoamingAllowed    bool    `tfsdk:"isRoamingAllowed"`
// 	SimNumber           *string `tfsdk:"simNumber"`
// 	DisconnectionReason *string `tfsdk:"disconnectionReason"`
// 	IsSimSlot1Detected  bool    `tfsdk:"isSimSlot1Detected"`
// 	IsSimSlot2Detected  bool    `tfsdk:"isSimSlot2Detected"`
// }

// type InterfaceInfo struct {
// 	Id                  string  `tfsdk:"id"`
// 	Name                *string `tfsdk:"name"`
// 	UpstreamBandwidth   *int64  `tfsdk:"upstreamBandwidth"`
// 	DownstreamBandwidth *int64  `tfsdk:"downstreamBandwidth"`
// 	DestType            *string `tfsdk:"destType"`
// }

// type IPInfo struct {
// 	Ip          *string  `tfsdk:"ip"`
// 	CountryCode *string  `tfsdk:"countryCode"`
// 	CountryName *string  `tfsdk:"countryName"`
// 	City        *string  `tfsdk:"city"`
// 	State       *string  `tfsdk:"state"`
// 	Provider    *string  `tfsdk:"provider"`
// 	Latitude    *float64 `tfsdk:"latitude"`
// 	Longitude   *float64 `tfsdk:"longitude"`
// }

// type RecentConnection struct {
// 	Duration      *int64  `tfsdk:"duration"`
// 	InterfaceName *string `tfsdk:"interfaceName"`
// 	DeviceName    *string `tfsdk:"deviceName"`
// 	LastConnected *string `tfsdk:"lastConnected"`
// 	PopName       *string `tfsdk:"popName"`
// 	RemoteIP      *string `tfsdk:"remoteIP"`
// 	RemoteIPInfo  *IPInfo `tfsdk:"remoteIPInfo"`
// }

// type InterfaceLinkState struct {
// 	Id        *string `tfsdk:"id"`
// 	Up        *bool   `tfsdk:"up"`
// 	MediaIn   *bool   `tfsdk:"mediaIn"`
// 	LinkSpeed *string `tfsdk:"linkSpeed"`
// 	Duplex    *string `tfsdk:"duplex"`
// }

type SocketInfo struct {
	Id        *string `tfsdk:"id"`
	Serial    *string `tfsdk:"serial"`
	IsPrimary *bool   `tfsdk:"is_primary"`
	// Platform          *string `tfsdk:"platform"`
	// Version           *string `tfsdk:"version"`
	// VersionUpdateTime *string `tfsdk:"versionUpdateTime"`
}

type SiteInfo struct {
	Name *string `tfsdk:"name"`
	// Type *string `tfsdk:"type"`
	// Description  *string         `tfsdk:"description"`
	// CountryCode  *string         `tfsdk:"countryCode"`
	// Region       *string         `tfsdk:"region"`
	// CountryName  *string         `tfsdk:"countryName"`
	// IsHA         *bool           `tfsdk:"isHA"`
	// ConnType     *string         `tfsdk:"connType"`
	// CreationTime *string         `tfsdk:"creationTime"`
	// Interfaces   []InterfaceInfo `tfsdk:"interfaces"`
	Sockets []SocketInfo `tfsdk:"sockets"`
	// Ipsec        []IPSecInfo     `tfsdk:"ipsec"`
}

// type IPSecInfo struct {
// 	IsPrimary  *bool   `tfsdk:"isPrimary"`
// 	CatoIP     *string `tfsdk:"catoIP"`
// 	RemoteIP   *string `tfsdk:"remoteIP"`
// 	IkeVersion *int64  `tfsdk:"ikeVersion"`
// }

func (d *accountSnapshotSiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accountSnapshotSite"
}

func (d *accountSnapshotSiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for the site",
				Required:    true,
			},
			"info": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Site Name",
						Computed:    true,
					},
					"sockets": schema.ListNestedAttribute{
						Description: "List of sockets attached to the site",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Socket id",
									Computed:    true,
								},
								"serial": schema.StringAttribute{
									Description: "Socket serial number",
									Computed:    true,
								},
								"is_primary": schema.BoolAttribute{
									Description: "Socket is primary",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *accountSnapshotSiteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *accountSnapshotSiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state SiteSnapshot
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountSnapshotSite, err := d.client.catov2.AccountSnapshot(ctx, []string{*state.Id}, nil, &d.client.AccountId)
	tflog.Debug(ctx, "Read.AccountSnapshot.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(accountSnapshotSite),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	if len(accountSnapshotSite.AccountSnapshot.Sites) == 1 {
		state = SiteSnapshot{
			Id: accountSnapshotSite.AccountSnapshot.Sites[0].ID,
			Info: &SiteInfo{
				Name: accountSnapshotSite.AccountSnapshot.Sites[0].InfoSiteSnapshot.Name,
			},
		}

		for _, socket := range accountSnapshotSite.AccountSnapshot.Sites[0].InfoSiteSnapshot.GetSockets() {
			state.Info.Sockets = append(state.Info.Sockets, SocketInfo{
				Id:        socket.ID,
				Serial:    socket.Serial,
				IsPrimary: socket.IsPrimary,
			})
		}

	} else {
		tflog.Error(ctx, "Can't find Site into AccountSnapshot")
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
