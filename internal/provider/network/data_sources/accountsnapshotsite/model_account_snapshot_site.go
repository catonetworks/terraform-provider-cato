package accountsnapshotsite

type SiteSnapshot struct {
	ID       *string `tfsdk:"id"`
	SiteName *string `tfsdk:"site_name"`
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
// 	ID                  *string              `tfsdk:"id"`
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
// 	ID                     *string            `tfsdk:"id"`
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
// 	ID                  string  `tfsdk:"id"`
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
// 	ID        *string `tfsdk:"id"`
// 	Up        *bool   `tfsdk:"up"`
// 	MediaIn   *bool   `tfsdk:"mediaIn"`
// 	LinkSpeed *string `tfsdk:"linkSpeed"`
// 	Duplex    *string `tfsdk:"duplex"`
// }

type SocketInfo struct {
	ID        *string `tfsdk:"id"`
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
