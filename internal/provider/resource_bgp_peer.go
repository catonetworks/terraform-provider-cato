package provider

import (
	"context"
	"strconv"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &bgpPeerResource{}
	_ resource.ResourceWithConfigure   = &bgpPeerResource{}
	_ resource.ResourceWithImportState = &bgpPeerResource{}
)

func NewBgpPeerResource() resource.Resource {
	return &bgpPeerResource{}
}

type bgpPeerResource struct {
	client *catoClientData
}

type BgpPeer struct {
	ID                     types.String `tfsdk:"id"`
	SiteId                 types.String `tfsdk:"site_id"`
	Name                   types.String `tfsdk:"name"`
	PeerAsn                types.Int64  `tfsdk:"peer_asn"`
	CatoAsn                types.Int64  `tfsdk:"cato_asn"`
	PeerIp                 types.String `tfsdk:"peer_ip"`
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
	SubscriptionId types.String `tfsdk:"subscription_id"`
}

func (r *bgpPeerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_bgp_peer`resource contains the configuration parameters necessary to add a BGP peer to a Cato site. Documentation for the underlying API used in this resource can be found at [mutation.site.AddBgpPeerPayload()](https://api.catonetworks.com/documentation/#definition-AddBgpPeerPayload).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the BGP peer.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site Id",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the BGP configuration entity",
				Required:    true,
			},
			"peer_asn": schema.Int64Attribute{
				Description: "The AS number of the peer BGP endpoint.",
				Required:    true,
			},
			"cato_asn": schema.Int64Attribute{
				Description: "The AS number of Cato's BGP endpoint.",
				Required:    true,
			},
			"peer_ip": schema.StringAttribute{
				Description: "The IP address of the BGP peer, this is the configured ip from the Site->IPSec Tunnel (Primary or Secondary)->Private IPs->Site",
				Required:    true,
			},
			"advertise_default_route": schema.BoolAttribute{
				Description: "Advertise the default route (0.0.0.0/0) if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"advertise_all_routes": schema.BoolAttribute{
				Description: "Advertise all routes if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"advertise_summary_routes": schema.BoolAttribute{
				Description: "Advertise summarized routes if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"summary_route": schema.ListNestedAttribute{
				Description: "Summarized routes to advertise.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"route": schema.StringAttribute{
							Description: "Subnet of the summarized route to be advertised.",
							Required:    true,
						},
						"community": schema.ListNestedAttribute{
							Description: "Community values to associate with the summarized route.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"from": schema.Int64Attribute{
										Description: "Start of the community range.",
										Required:    true,
									},
									"to": schema.Int64Attribute{
										Description: "End of the community range.",
										Required:    true,
									},
								},
							},
							Optional: true,
						},
					},
				},
			},
			"default_action": schema.StringAttribute{
				Description: "Default action for routes not matching filters (ACCEPT or DROP).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ACCEPT", "DROP"),
				},
			},
			"perform_nat": schema.BoolAttribute{
				Description: "Perform NAT if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"md5_auth_key": schema.StringAttribute{
				Description: "MD5 authentication key for secure sessions.",
				Optional:    true,
				Sensitive:   true,
			},
			"metric": schema.Int64Attribute{
				Description: "Route preference metric; lower values are given precedence.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(50),
			},
			"hold_time": schema.Int64Attribute{
				Description: "Time (in seconds) before declaring the peer unreachable.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"keepalive_interval": schema.Int64Attribute{
				Description: "Time (in seconds) between keepalive messages.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
			},
			"bfd_enabled": schema.BoolAttribute{
				Description: "Enable BFD for session failure detection if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"bfd_settings": schema.SingleNestedAttribute{
				Description: "Required BFD configuration if BFD is enabled.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"transmit_interval": schema.Int64Attribute{
						Description: "Time interval (in milliseconds) between BFD packets sent by this peer.",
						Optional:    true,
					},
					"receive_interval": schema.Int64Attribute{
						Description: "Time interval (in milliseconds) in which this peer expects to receive BFD packets.",
						Optional:    true,
					},
					"multiplier": schema.Int64Attribute{
						Description: "Number of missed BFD packets before considering the session down.",
						Optional:    true,
					},
				},
			},
			"tracking": schema.SingleNestedAttribute{
				Description: "Configuration for tracking the health and status of the BGP peer.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Indicates if tracking is enabled.",
						Optional:    true,
					},
					"alert_frequency": schema.StringAttribute{
						Description: "Frequency of health alerts.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("HOURLY", "DAILY", "WEEKLY", "IMMEDIATE"),
						},
					},
					"subscription_id": schema.StringAttribute{
						Description: "Subscription ID associated with this tracking rule.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *bgpPeerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bgp_peer"
}

func (r *bgpPeerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *bgpPeerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *bgpPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BgpPeer
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.AddBgpPeerInput{
		Name:          plan.Name.ValueString(),
		PeerAsn:       scalars.Asn32(plan.PeerAsn.String()),
		CatoAsn:       scalars.Asn16(plan.CatoAsn.String()),
		PeerIP:        plan.PeerIp.ValueString(),
		DefaultAction: cato_models.BgpDefaultAction(plan.DefaultAction.ValueString()),
		Site: &cato_models.SiteRefInput{
			By:    cato_models.ObjectRefByID,
			Input: plan.SiteId.ValueString(),
		},
	}

	// Optional fields
	if !plan.AdvertiseAllRoutes.IsNull() {
		input.AdvertiseAllRoutes = plan.AdvertiseAllRoutes.ValueBool()
	}
	if !plan.AdvertiseDefaultRoute.IsNull() {
		input.AdvertiseDefaultRoute = plan.AdvertiseDefaultRoute.ValueBool()
	}
	if !plan.AdvertiseSummaryRoutes.IsNull() {
		input.AdvertiseSummaryRoutes = plan.AdvertiseSummaryRoutes.ValueBool()
	}
	if !plan.PerformNat.IsNull() {
		input.PerformNat = plan.PerformNat.ValueBool()
	}
	if !plan.Md5AuthKey.IsNull() {
		input.Md5AuthKey = plan.Md5AuthKey.ValueStringPointer()
	}
	if !plan.Metric.IsNull() {
		input.Metric = plan.Metric.ValueInt64()
	}
	if !plan.HoldTime.IsNull() {
		input.HoldTime = plan.HoldTime.ValueInt64()
	}
	if !plan.KeepaliveInterval.IsNull() {
		input.KeepaliveInterval = plan.KeepaliveInterval.ValueInt64()
	}
	if !plan.BfdEnabled.IsNull() {
		input.BfdEnabled = plan.BfdEnabled.ValueBool()
		if !plan.BfdSettings.IsNull() {
			var bfdSettingsInput cato_models.BfdSettingsInput
			bfdSettings := BfdSettingsInput{}
			diags = plan.BfdSettings.As(ctx, &bfdSettings, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			bfdSettingsInput.TransmitInterval = bfdSettings.TransmitInterval.ValueInt64()
			bfdSettingsInput.ReceiveInterval = bfdSettings.ReceiveInterval.ValueInt64()
			bfdSettingsInput.Multiplier = bfdSettings.Multiplier.ValueInt64()
			input.BfdSettings = &bfdSettingsInput
		}
	}
	if !plan.SummaryRoute.IsNull() {
		input.SummaryRoute = make([]*cato_models.BgpSummaryRouteInput, 0)
		summaryRoutes := make([]BgpSummaryRouteInput, 0, len(plan.SummaryRoute.Elements()))
		diags = plan.SummaryRoute.ElementsAs(ctx, &summaryRoutes, false)
		resp.Diagnostics.Append(diags...)

		for _, summaryRoute := range summaryRoutes {
			var summaryRouteInput cato_models.BgpSummaryRouteInput
			summaryRouteInput.Route = summaryRoute.Route.ValueString()
			summaryRouteInput.Community = make([]*cato_models.BgpCommunityInput, 0)
			communities := make([]BgpCommunityInput, 0, len(summaryRoute.Community.Elements()))
			diags = summaryRoute.Community.ElementsAs(ctx, &communities, false)
			resp.Diagnostics.Append(diags...)

			for _, community := range communities {
				summaryRouteInput.Community = append(summaryRouteInput.Community, &cato_models.BgpCommunityInput{
					From: scalars.Asn16(strconv.FormatInt(community.From.ValueInt64(), 10)),
					To:   scalars.Asn16(strconv.FormatInt(community.To.ValueInt64(), 10)),
				})
			}

			input.SummaryRoute = append(input.SummaryRoute, &summaryRouteInput)
		}
	}
	if !plan.Tracking.IsNull() {
		var trackingInput cato_models.BgpTrackingInput
		tracking := BgpTrackingInput{}
		diags = plan.Tracking.As(ctx, &tracking, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		trackingInput.Enabled = tracking.Enabled.ValueBool()
		trackingInput.AlertFrequency = cato_models.PolicyRuleTrackingFrequencyEnum(tracking.AlertFrequency.ValueString())
		trackingInput.SubscriptionID = tracking.SubscriptionId.ValueString()
		input.Tracking = &trackingInput
	}

	tflog.Debug(ctx, "Create.SiteAddBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	addBgpPeerPayload, err := r.client.catov2.SiteAddBgpPeer(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addBgpPeerPayload),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteAddBgpPeer",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bgpPeer := addBgpPeerPayload.GetSite().GetAddBgpPeer().GetBgpPeer()
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), bgpPeer.GetID())
}

func (r *bgpPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state BgpPeer
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bgpPeerRefInput := cato_models.BgpPeerRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.ID.ValueString(),
	}
	tflog.Debug(ctx, "Read.SiteBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(bgpPeerRefInput),
	})
	result, err := r.client.catov2.SiteBgpPeer(ctx, bgpPeerRefInput, r.client.AccountId)
	tflog.Debug(ctx, "Read.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteBgpPeer",
			err.Error(),
		)
		return
	}

	if result == nil || result.GetSite() == nil || result.GetSite().GetBgpPeer() == nil {
		tflog.Warn(ctx, "BGP peer resource wasn't found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	bgpPeer := result.GetSite().GetBgpPeer()
	bgpPeerInput := ConvertToBgpPeer(*bgpPeer)

	diags = resp.State.Set(ctx, &bgpPeerInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bgpPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BgpPeer
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.UpdateBgpPeerInput{
		ID: plan.ID.ValueString(),
	}

	if !plan.Name.IsNull() {
		input.Name = plan.Name.ValueStringPointer()
	}
	if !plan.PeerAsn.IsNull() {
		peerAsn := scalars.Asn32(strconv.FormatInt(plan.PeerAsn.ValueInt64(), 10))
		input.PeerAsn = &peerAsn
	}
	if !plan.CatoAsn.IsNull() {
		catoAsn := scalars.Asn16(strconv.FormatInt(plan.CatoAsn.ValueInt64(), 10))
		input.CatoAsn = &catoAsn
	}
	if !plan.PeerIp.IsNull() {
		input.PeerIP = plan.PeerIp.ValueStringPointer()
	}
	if !plan.DefaultAction.IsNull() {
		input.DefaultAction = (*cato_models.BgpDefaultAction)(plan.DefaultAction.ValueStringPointer())
	}
	if !plan.AdvertiseAllRoutes.IsNull() {
		input.AdvertiseAllRoutes = plan.AdvertiseAllRoutes.ValueBoolPointer()
	}
	if !plan.AdvertiseDefaultRoute.IsNull() {
		input.AdvertiseDefaultRoute = plan.AdvertiseDefaultRoute.ValueBoolPointer()
	}
	if !plan.AdvertiseSummaryRoutes.IsNull() {
		input.AdvertiseSummaryRoutes = plan.AdvertiseSummaryRoutes.ValueBoolPointer()
	}
	if !plan.PerformNat.IsNull() {
		input.PerformNat = plan.PerformNat.ValueBoolPointer()
	}
	if !plan.Md5AuthKey.IsNull() {
		input.Md5AuthKey = plan.Md5AuthKey.ValueStringPointer()
	}
	if !plan.Metric.IsNull() {
		input.Metric = plan.Metric.ValueInt64Pointer()
	}
	if !plan.HoldTime.IsNull() {
		input.HoldTime = plan.HoldTime.ValueInt64Pointer()
	}
	if !plan.KeepaliveInterval.IsNull() {
		input.KeepaliveInterval = plan.KeepaliveInterval.ValueInt64Pointer()
	}
	if !plan.BfdEnabled.IsNull() {
		input.BfdEnabled = plan.BfdEnabled.ValueBoolPointer()
		if !plan.BfdSettings.IsNull() {
			var bfdSettingsInput cato_models.BfdSettingsInput
			bfdSettings := BfdSettingsInput{}
			diags = plan.BfdSettings.As(ctx, &bfdSettings, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			bfdSettingsInput.TransmitInterval = bfdSettings.TransmitInterval.ValueInt64()
			bfdSettingsInput.ReceiveInterval = bfdSettings.ReceiveInterval.ValueInt64()
			bfdSettingsInput.Multiplier = bfdSettings.Multiplier.ValueInt64()
			input.BfdSettings = &bfdSettingsInput
		}
	}
	if !plan.SummaryRoute.IsNull() {
		input.SummaryRoute = make([]*cato_models.BgpSummaryRouteInput, 0)
		summaryRoutes := make([]BgpSummaryRouteInput, 0, len(plan.SummaryRoute.Elements()))
		diags = plan.SummaryRoute.ElementsAs(ctx, &summaryRoutes, false)
		resp.Diagnostics.Append(diags...)

		for _, summaryRoute := range summaryRoutes {
			var summaryRouteInput cato_models.BgpSummaryRouteInput
			summaryRouteInput.Route = summaryRoute.Route.ValueString()
			summaryRouteInput.Community = make([]*cato_models.BgpCommunityInput, 0)
			communities := make([]BgpCommunityInput, 0, len(summaryRoute.Community.Elements()))
			diags = summaryRoute.Community.ElementsAs(ctx, &communities, false)
			resp.Diagnostics.Append(diags...)

			for _, community := range communities {
				summaryRouteInput.Community = append(summaryRouteInput.Community, &cato_models.BgpCommunityInput{
					From: scalars.Asn16(strconv.FormatInt(community.From.ValueInt64(), 10)),
					To:   scalars.Asn16(strconv.FormatInt(community.To.ValueInt64(), 10)),
				})
			}

			input.SummaryRoute = append(input.SummaryRoute, &summaryRouteInput)
		}
	}
	if !plan.Tracking.IsNull() {
		var trackingInput cato_models.BgpTrackingInput
		tracking := BgpTrackingInput{}
		diags = plan.Tracking.As(ctx, &tracking, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		trackingInput.Enabled = tracking.Enabled.ValueBool()
		trackingInput.AlertFrequency = cato_models.PolicyRuleTrackingFrequencyEnum(tracking.AlertFrequency.ValueString())
		trackingInput.SubscriptionID = tracking.SubscriptionId.ValueString()
		input.Tracking = &trackingInput
	}

	tflog.Debug(ctx, "Update.SiteUpdateBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	SiteUpdateBgpPeerResponse, err := r.client.catov2.SiteUpdateBgpPeer(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(SiteUpdateBgpPeerResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteUpdateBgpPeer",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bgpPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BgpPeer
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeBgpPeerInput := cato_models.RemoveBgpPeerInput{ID: state.ID.ValueString()}

	tflog.Debug(ctx, "Delete.SiteUpdateBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeBgpPeerInput),
	})
	SiteRemoveBgpPeerResponse, err := r.client.catov2.SiteRemoveBgpPeer(ctx, removeBgpPeerInput, r.client.AccountId)
	tflog.Debug(ctx, "Delete.SiteUpdateBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(SiteRemoveBgpPeerResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteRemoveBgpPeer",
			err.Error(),
		)
		return
	}
}

// BGP Peer Utils Functions

func convertSummaryRoutes(input []*cato_go_sdk.Site_Site_BgpPeer_SummaryRoute) types.List {
	var summaryRoutes []attr.Value
	for _, summaryRoute := range input {
		var communities []attr.Value
		for _, community := range summaryRoute.Community {
			from, _ := strconv.ParseInt(string(community.From), 10, 64)
			to, _ := strconv.ParseInt(string(community.To), 10, 64)
			communityInput, _ := types.ObjectValue(
				map[string]attr.Type{
					"from": types.Int64Type,
					"to":   types.Int64Type,
				},
				map[string]attr.Value{
					"from": types.Int64Value(from),
					"to":   types.Int64Value(to),
				},
			)
			communities = append(communities, communityInput)
		}

		communityList, _ := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"from": types.Int64Type,
				"to":   types.Int64Type,
			},
		}, communities)

		summaryRouteInput, _ := types.ObjectValue(
			map[string]attr.Type{
				"route": types.StringType,
				"community": types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"from": types.Int64Type,
						"to":   types.Int64Type,
					},
				}},
			},
			map[string]attr.Value{
				"route":     types.StringValue(summaryRoute.Route),
				"community": communityList,
			},
		)
		summaryRoutes = append(summaryRoutes, summaryRouteInput)
	}

	listVal, _ := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"route": types.StringType,
			"community": types.ListType{ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"from": types.Int64Type,
					"to":   types.Int64Type,
				},
			}},
		},
	}, summaryRoutes)

	return listVal
}

func convertBfdSettings(input *cato_go_sdk.Site_Site_BgpPeer_BfdSettingsBgpPeer) types.Object {
	if input == nil {
		return types.ObjectNull(map[string]attr.Type{
			"transmit_interval": types.Int64Type,
			"receive_interval":  types.Int64Type,
			"multiplier":        types.Int64Type,
		})
	}

	bfdSettings, _ := types.ObjectValue(
		map[string]attr.Type{
			"transmit_interval": types.Int64Type,
			"receive_interval":  types.Int64Type,
			"multiplier":        types.Int64Type,
		},
		map[string]attr.Value{
			"transmit_interval": types.Int64Value(input.TransmitInterval),
			"receive_interval":  types.Int64Value(input.ReceiveInterval),
			"multiplier":        types.Int64Value(input.Multiplier),
		},
	)
	return bfdSettings
}

func convertTracking(input *cato_go_sdk.Site_Site_BgpPeer_TrackingBgpPeer) types.Object {
	if input == nil {
		return types.ObjectNull(map[string]attr.Type{
			"enabled":         types.BoolType,
			"alert_frequency": types.StringType,
			"subscription_id": types.StringType,
		})
	}

	tracking, _ := types.ObjectValue(
		map[string]attr.Type{
			"enabled":         types.BoolType,
			"alert_frequency": types.StringType,
			"subscription_id": types.StringType,
		},
		map[string]attr.Value{
			"enabled":         types.BoolValue(input.Enabled),
			"alert_frequency": types.StringValue(input.AlertFrequency.String()),
			"subscription_id": types.StringValue(*input.SubscriptionID),
		},
	)
	return tracking
}

func ConvertToBgpPeer(input cato_go_sdk.Site_Site_BgpPeer) BgpPeer {
	peerAsnInt64, _ := strconv.ParseInt(string(input.PeerAsn), 10, 64)
	catoAsnInt64, _ := strconv.ParseInt(string(input.CatoAsn), 10, 64)
	return BgpPeer{
		ID:                     types.StringValue(input.ID),
		SiteId:                 types.StringValue(input.Site.ID),
		Name:                   types.StringValue(input.Name),
		PeerAsn:                types.Int64Value(peerAsnInt64),
		CatoAsn:                types.Int64Value(catoAsnInt64), // Convert Asn16 to int64
		PeerIp:                 types.StringValue(input.PeerIP),
		AdvertiseDefaultRoute:  types.BoolValue(input.AdvertiseDefaultRoute),
		AdvertiseAllRoutes:     types.BoolValue(input.AdvertiseAllRoutes),
		AdvertiseSummaryRoutes: types.BoolValue(input.AdvertiseSummaryRoutes),
		SummaryRoute:           convertSummaryRoutes(input.SummaryRoute),
		DefaultAction:          types.StringValue(string(input.DefaultAction)), // Assuming BgpDefaultAction is string
		PerformNat:             types.BoolValue(input.PerformNat),
		Md5AuthKey:             utils.ConvertOptionalString(input.Md5AuthKey),
		Metric:                 types.Int64Value(input.Metric),
		HoldTime:               types.Int64Value(input.HoldTime),
		KeepaliveInterval:      types.Int64Value(input.KeepaliveInterval),
		BfdEnabled:             types.BoolValue(input.BfdEnabled),
		BfdSettings:            convertBfdSettings(input.BfdSettingsBgpPeer),
		Tracking:               convertTracking(input.TrackingBgpPeer),
	}
}
