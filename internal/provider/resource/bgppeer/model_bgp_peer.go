package bgppeer

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BgpPeer struct {
	ID                     types.String `tfsdk:"id"`
	SiteID                 types.String `tfsdk:"site_id"`
	Name                   types.String `tfsdk:"name"`
	PeerAsn                types.Int64  `tfsdk:"peer_asn"`
	CatoAsn                types.Int64  `tfsdk:"cato_asn"`
	PeerIP                 types.String `tfsdk:"peer_ip"`
	AdvertiseDefaultRoute  types.Bool   `tfsdk:"advertise_default_route"`
	AdvertiseAllRoutes     types.Bool   `tfsdk:"advertise_all_routes"`
	AdvertiseSummaryRoutes types.Bool   `tfsdk:"advertise_summary_routes"`
	SummaryRoute           types.List   `tfsdk:"summary_route"`
	DefaultAction          types.String `tfsdk:"default_action"`
	PerformNat             types.Bool   `tfsdk:"perform_nat"`
	Md5AuthKey             types.String `tfsdk:"md5_auth_key"`
	Metric                 types.Int64  `tfsdk:"metric"`
	HoldTime               types.Int64  `tfsdk:"hold_time"`
	KeepaliveInterval      types.Int64  `tfsdk:"keepalive_interval"`
	BfdEnabled             types.Bool   `tfsdk:"bfd_enabled"`
	BfdSettings            types.Object `tfsdk:"bfd_settings"`
	Tracking               types.Object `tfsdk:"tracking"`
}

type BfdSettingsInput struct {
	TransmitInterval types.Int64 `tfsdk:"transmit_interval"`
	ReceiveInterval  types.Int64 `tfsdk:"receive_interval"`
	Multiplier       types.Int64 `tfsdk:"multiplier"`
}

type BgpSummaryRouteInput struct {
	Route     types.String `tfsdk:"route"`
	Community types.List   `tfsdk:"community"`
}

type BgpCommunityInput struct {
	From types.Int64 `tfsdk:"from"`
	To   types.Int64 `tfsdk:"to"`
}

type BgpTrackingInput struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	AlertFrequency types.String `tfsdk:"alert_frequency"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
}
