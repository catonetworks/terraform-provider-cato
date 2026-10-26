package bgppeer

import (
	"context"
	"strconv"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

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
		SiteID:                 types.StringValue(input.Site.ID),
		Name:                   types.StringValue(input.Name),
		PeerAsn:                types.Int64Value(peerAsnInt64),
		CatoAsn:                types.Int64Value(catoAsnInt64), // Convert Asn16 to int64
		PeerIP:                 types.StringValue(input.PeerIP),
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

// convertCreateInput converts between the Terraform model and SDK values.
//
//nolint:funlen
func convertCreateInput(
	ctx context.Context,
	plan BgpPeer,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.AddBgpPeerInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultInput = cato_models.AddBgpPeerInput{
		Name:          plan.Name.ValueString(),
		PeerAsn:       scalars.Asn32(plan.PeerAsn.String()),
		CatoAsn:       scalars.Asn16(plan.CatoAsn.String()),
		PeerIP:        plan.PeerIP.ValueString(),
		DefaultAction: cato_models.BgpDefaultAction(plan.DefaultAction.ValueString()),
		Site: &cato_models.SiteRefInput{
			By:    cato_models.ObjectRefByID,
			Input: plan.SiteID.ValueString(),
		},
	}

	// Optional fields
	if !plan.AdvertiseAllRoutes.IsNull() {
		resultInput.AdvertiseAllRoutes = plan.AdvertiseAllRoutes.ValueBool()
	}
	if !plan.AdvertiseDefaultRoute.IsNull() {
		resultInput.AdvertiseDefaultRoute = plan.AdvertiseDefaultRoute.ValueBool()
	}
	if !plan.AdvertiseSummaryRoutes.IsNull() {
		resultInput.AdvertiseSummaryRoutes = plan.AdvertiseSummaryRoutes.ValueBool()
	}
	if !plan.PerformNat.IsNull() {
		resultInput.PerformNat = plan.PerformNat.ValueBool()
	}
	if !plan.Md5AuthKey.IsNull() {
		resultInput.Md5AuthKey = plan.Md5AuthKey.ValueStringPointer()
	}
	if !plan.Metric.IsNull() {
		resultInput.Metric = plan.Metric.ValueInt64()
	}
	if !plan.HoldTime.IsNull() {
		resultInput.HoldTime = plan.HoldTime.ValueInt64()
	}
	if !plan.KeepaliveInterval.IsNull() {
		resultInput.KeepaliveInterval = plan.KeepaliveInterval.ValueInt64()
	}
	if !plan.BfdEnabled.IsNull() {
		resultInput.BfdEnabled = plan.BfdEnabled.ValueBool()
		if !plan.BfdSettings.IsNull() {
			var bfdSettingsInput cato_models.BfdSettingsInput
			bfdSettings := BfdSettingsInput{}
			diags = plan.BfdSettings.As(ctx, &bfdSettings, basetypes.ObjectAsOptions{})
			diagnostics.Append(diags...)
			bfdSettingsInput.TransmitInterval = bfdSettings.TransmitInterval.ValueInt64()
			bfdSettingsInput.ReceiveInterval = bfdSettings.ReceiveInterval.ValueInt64()
			bfdSettingsInput.Multiplier = bfdSettings.Multiplier.ValueInt64()
			resultInput.BfdSettings = &bfdSettingsInput
		}
	}
	if !plan.SummaryRoute.IsNull() {
		resultInput.SummaryRoute = make([]*cato_models.BgpSummaryRouteInput, 0)
		summaryRoutes := make([]BgpSummaryRouteInput, 0, len(plan.SummaryRoute.Elements()))
		diags = plan.SummaryRoute.ElementsAs(ctx, &summaryRoutes, false)
		diagnostics.Append(diags...)

		for _, summaryRoute := range summaryRoutes {
			var summaryRouteInput cato_models.BgpSummaryRouteInput
			summaryRouteInput.Route = summaryRoute.Route.ValueString()
			summaryRouteInput.Community = make([]*cato_models.BgpCommunityInput, 0)
			communities := make([]BgpCommunityInput, 0, len(summaryRoute.Community.Elements()))
			diags = summaryRoute.Community.ElementsAs(ctx, &communities, false)
			diagnostics.Append(diags...)

			for _, community := range communities {
				summaryRouteInput.Community = append(summaryRouteInput.Community, &cato_models.BgpCommunityInput{
					From: scalars.Asn16(strconv.FormatInt(community.From.ValueInt64(), 10)),
					To:   scalars.Asn16(strconv.FormatInt(community.To.ValueInt64(), 10)),
				})
			}

			resultInput.SummaryRoute = append(resultInput.SummaryRoute, &summaryRouteInput)
		}
	}
	if !plan.Tracking.IsNull() {
		var trackingInput cato_models.BgpTrackingInput
		tracking := BgpTrackingInput{}
		diags = plan.Tracking.As(ctx, &tracking, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		trackingInput.Enabled = tracking.Enabled.ValueBool()
		trackingInput.AlertFrequency = cato_models.PolicyRuleTrackingFrequencyEnum(tracking.AlertFrequency.ValueString())
		trackingInput.SubscriptionID = tracking.SubscriptionID.ValueString()
		resultInput.Tracking = &trackingInput
	}
	return resultInput
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(state BgpPeer) (resultBgpPeerRefInput cato_models.BgpPeerRefInput) {
	resultBgpPeerRefInput = cato_models.BgpPeerRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.ID.ValueString(),
	}
	return resultBgpPeerRefInput
}

// convertUpdateInput converts between the Terraform model and SDK values.
//
//nolint:gocyclo,funlen
func convertUpdateInput(
	ctx context.Context,
	plan BgpPeer,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.UpdateBgpPeerInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultInput = cato_models.UpdateBgpPeerInput{
		ID: plan.ID.ValueString(),
	}

	if !plan.Name.IsNull() {
		resultInput.Name = plan.Name.ValueStringPointer()
	}
	if !plan.PeerAsn.IsNull() {
		peerAsn := scalars.Asn32(strconv.FormatInt(plan.PeerAsn.ValueInt64(), 10))
		resultInput.PeerAsn = &peerAsn
	}
	if !plan.CatoAsn.IsNull() {
		catoAsn := scalars.Asn16(strconv.FormatInt(plan.CatoAsn.ValueInt64(), 10))
		resultInput.CatoAsn = &catoAsn
	}
	if !plan.PeerIP.IsNull() {
		resultInput.PeerIP = plan.PeerIP.ValueStringPointer()
	}
	if !plan.DefaultAction.IsNull() {
		resultInput.DefaultAction = (*cato_models.BgpDefaultAction)(plan.DefaultAction.ValueStringPointer())
	}
	if !plan.AdvertiseAllRoutes.IsNull() {
		resultInput.AdvertiseAllRoutes = plan.AdvertiseAllRoutes.ValueBoolPointer()
	}
	if !plan.AdvertiseDefaultRoute.IsNull() {
		resultInput.AdvertiseDefaultRoute = plan.AdvertiseDefaultRoute.ValueBoolPointer()
	}
	if !plan.AdvertiseSummaryRoutes.IsNull() {
		resultInput.AdvertiseSummaryRoutes = plan.AdvertiseSummaryRoutes.ValueBoolPointer()
	}
	if !plan.PerformNat.IsNull() {
		resultInput.PerformNat = plan.PerformNat.ValueBoolPointer()
	}
	if !plan.Md5AuthKey.IsNull() {
		resultInput.Md5AuthKey = plan.Md5AuthKey.ValueStringPointer()
	}
	if !plan.Metric.IsNull() {
		resultInput.Metric = plan.Metric.ValueInt64Pointer()
	}
	if !plan.HoldTime.IsNull() {
		resultInput.HoldTime = plan.HoldTime.ValueInt64Pointer()
	}
	if !plan.KeepaliveInterval.IsNull() {
		resultInput.KeepaliveInterval = plan.KeepaliveInterval.ValueInt64Pointer()
	}
	if !plan.BfdEnabled.IsNull() {
		resultInput.BfdEnabled = plan.BfdEnabled.ValueBoolPointer()
		if !plan.BfdSettings.IsNull() {
			var bfdSettingsInput cato_models.BfdSettingsInput
			bfdSettings := BfdSettingsInput{}
			diags = plan.BfdSettings.As(ctx, &bfdSettings, basetypes.ObjectAsOptions{})
			diagnostics.Append(diags...)
			bfdSettingsInput.TransmitInterval = bfdSettings.TransmitInterval.ValueInt64()
			bfdSettingsInput.ReceiveInterval = bfdSettings.ReceiveInterval.ValueInt64()
			bfdSettingsInput.Multiplier = bfdSettings.Multiplier.ValueInt64()
			resultInput.BfdSettings = &bfdSettingsInput
		}
	}
	if !plan.SummaryRoute.IsNull() {
		resultInput.SummaryRoute = make([]*cato_models.BgpSummaryRouteInput, 0)
		summaryRoutes := make([]BgpSummaryRouteInput, 0, len(plan.SummaryRoute.Elements()))
		diags = plan.SummaryRoute.ElementsAs(ctx, &summaryRoutes, false)
		diagnostics.Append(diags...)

		for _, summaryRoute := range summaryRoutes {
			var summaryRouteInput cato_models.BgpSummaryRouteInput
			summaryRouteInput.Route = summaryRoute.Route.ValueString()
			summaryRouteInput.Community = make([]*cato_models.BgpCommunityInput, 0)
			communities := make([]BgpCommunityInput, 0, len(summaryRoute.Community.Elements()))
			diags = summaryRoute.Community.ElementsAs(ctx, &communities, false)
			diagnostics.Append(diags...)

			for _, community := range communities {
				summaryRouteInput.Community = append(summaryRouteInput.Community, &cato_models.BgpCommunityInput{
					From: scalars.Asn16(strconv.FormatInt(community.From.ValueInt64(), 10)),
					To:   scalars.Asn16(strconv.FormatInt(community.To.ValueInt64(), 10)),
				})
			}

			resultInput.SummaryRoute = append(resultInput.SummaryRoute, &summaryRouteInput)
		}
	}
	if !plan.Tracking.IsNull() {
		var trackingInput cato_models.BgpTrackingInput
		tracking := BgpTrackingInput{}
		diags = plan.Tracking.As(ctx, &tracking, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		trackingInput.Enabled = tracking.Enabled.ValueBool()
		trackingInput.AlertFrequency = cato_models.PolicyRuleTrackingFrequencyEnum(tracking.AlertFrequency.ValueString())
		trackingInput.SubscriptionID = tracking.SubscriptionID.ValueString()
		resultInput.Tracking = &trackingInput
	}
	return resultInput
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state BgpPeer) (resultRemoveBgpPeerInput cato_models.RemoveBgpPeerInput) {
	resultRemoveBgpPeerInput = cato_models.RemoveBgpPeerInput{ID: state.ID.ValueString()}
	return resultRemoveBgpPeerInput
}
