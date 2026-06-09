package dhcp

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

type SdkClienter interface {
	V2() *cato_go_sdk.Client // return SDK Client v2
	AccountID() string
}

// SchemaDhcpSettings returns the schema for the DHCP settings of a network range resource.
func SchemaDhcpSettings(isRangeResource bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type)",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{planmodifiers.DHCPSettingsModifier(isRangeResource)},
		Attributes: map[string]schema.Attribute{
			"dhcp_type": schema.StringAttribute{
				Description: "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
				Required:    true,
				Validators:  []validator.String{validators.DHCPTypeValidator{}},
			},
			"ip_range": schema.StringAttribute{
				Description: "Network range dhcp range (format \"192.168.1.10-192.168.1.20\")",
				Optional:    true,
			},
			"relay_group_id": schema.StringAttribute{
				Description: "Network range dhcp relay group id",
				Optional:    true,
				Computed:    true,
			},
			"relay_group_name": schema.StringAttribute{
				Description: "Network range dhcp relay group name",
				Optional:    true,
				Computed:    true,
			},
			"dhcp_microsegmentation": schema.BoolAttribute{
				Description: "DHCP Microsegmentation. When enabled, the DHCP server will allocate /32 subnet mask. " +
					"Make sure to enable the proper Firewall rules and enable it with caution, " +
					"as it is not supported on all operating systems; monitor the network closely after activation. " +
					"This setting can only be configured when dhcp_type is set to DHCP_RANGE.",
				Optional: true,
				Computed: true,
			},
		},
	}
}

// PrepareDHCPSettings constructs the NetworkDhcpSettingsInput part of API input for SiteUpdateNetworkRange()
// from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func PrepareDHCPSettings(ctx context.Context, client SdkClienter, rangeType cato_models.SubnetType,
	dhcpSettings types.Object, diags *diag.Diagnostics,
) *cato_models.NetworkDhcpSettingsInput {
	var tfDhcpSettings tf.DhcpSettings

	switch rangeType {
	case cato_models.SubnetTypeNative, cato_models.SubnetTypeVlan: // OK, allowed
	default:
		return nil // for other range types, DHCP settings are not relevant, ignore any provided settings
	}

	if !utils.HasValue(dhcpSettings) {
		return nil
	}

	if utils.CheckErr(diags, dhcpSettings.As(ctx, &tfDhcpSettings, basetypes.ObjectAsOptions{})) {
		return nil
	}
	if !utils.HasValue(tfDhcpSettings.DhcpType) {
		return nil
	}

	dhcpType := cato_models.DhcpType(tfDhcpSettings.DhcpType.ValueString())

	input := &cato_models.NetworkDhcpSettingsInput{
		DhcpType: dhcpType,
		IPRange:  parse.KnownStringPointer(tfDhcpSettings.IPRange),
	}

	// if dhcp type is DHCP_RELAY and we don't have relayGroupID (just Name), fetch the relayGroupID
	if dhcpType == cato_models.DhcpTypeDhcpRelay && (!utils.HasValue(tfDhcpSettings.RelayGroupID)) {
		if !utils.HasValue(tfDhcpSettings.RelayGroupName) {
			diags.AddError("Missing DHCP relay group name", "DHCP settings of type DHCP_RELAY require a relay group name to be specified.")
			return nil
		}
		relayGroupID := getDhcpRelayGroupID(ctx, client, tfDhcpSettings.RelayGroupName.ValueString(), diags)
		if diags.HasError() {
			return nil
		}
		input.RelayGroupID = &relayGroupID
	}

	// Microsegmentation is only relevant for DHCP range
	if input.DhcpType == cato_models.DhcpTypeDhcpRange {
		input.DhcpMicrosegmentation = parse.KnownBoolPointer(tfDhcpSettings.DhcpMicrosegmentation)
	}

	return input
}

// getDhcpRelayGroupID looks up the DHCP relay group ID based on the provided relay group name.
func getDhcpRelayGroupID(ctx context.Context, client SdkClienter, relayGroupName string, diags *diag.Diagnostics,
) (groupID string) {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult := lookupDhcpRelayGroupID(ctx, client, diags)
	if diags.HasError() {
		return ""
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
			return item.Entity.GetID()
		}
	}

	// Relay group not found
	diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("DHCP relay group: '%s' not found", relayGroupName))
	return ""
}

// lookupDhcpRelayGroupID looks up the DHCP relay groups
func lookupDhcpRelayGroupID(ctx context.Context, client SdkClienter, diags *diag.Diagnostics,
) *cato_go_sdk.EntityLookup {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult, err := client.V2().EntityLookupMinimal(ctx, client.AccountID(), cato_models.EntityTypeDhcpRelayGroup,
		nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("An error was encountered when looking up DHCP relay group: %v", err))
		return nil
	}
	return dhcpRelayGroupResult
}

// SettingsDefault returns types.Object or tf.DhcpSettings, with type=ACCOUNT_DEFAULT, other fields null.
func SettingsDefault(ctx context.Context, diags *diag.Diagnostics) types.Object {
	tfDhcpSettings := tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeAccountDefault)),
		IPRange:               types.StringNull(),
		RelayGroupID:          types.StringNull(),
		RelayGroupName:        types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}
	dhcpSettingsObj, objDiags := types.ObjectValueFrom(ctx, tf.SiteNativeRangeDhcpResourceAttrTypes, tfDhcpSettings)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(tf.SiteNativeRangeDhcpResourceAttrTypes)
	}
	return dhcpSettingsObj
}

// ParseSettings converts API DHCP settings data to the types.Object or tf.DhcpSettings
func ParseSettings(ctx context.Context, client SdkClienter,
	dhcpSettings *cato_go_sdk.NetworkRange_Site_NetworkRange_DhcpSettings,
	diags *diag.Diagnostics,
) types.Object {
	tfDhcpSettings := tf.DhcpSettings{
		DhcpType:              types.StringValue(string(dhcpSettings.DhcpType)),
		IPRange:               types.StringNull(),
		RelayGroupID:          types.StringNull(),
		RelayGroupName:        types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	switch dhcpSettings.DhcpType {
	case cato_models.DhcpTypeDhcpRelay:
		tfDhcpSettings.RelayGroupID = types.StringPointerValue(dhcpSettings.RelayGroupID)
		if dhcpSettings.RelayGroupID != nil {
			relayGroupName := GetRelayGroupName(ctx, client, *dhcpSettings.RelayGroupID, diags)
			tfDhcpSettings.RelayGroupName = types.StringPointerValue(relayGroupName)
		}

	case cato_models.DhcpTypeDhcpRange:
		tfDhcpSettings.IPRange = types.StringPointerValue(dhcpSettings.IPRange)
		tfDhcpSettings.DhcpMicrosegmentation = types.BoolValue(dhcpSettings.DhcpMicrosegmentation)

	case cato_models.DhcpTypeAccountDefault, cato_models.DhcpTypeDhcpDisabled:
		// nullValues

	default:
		diags.AddError("Unsupported DHCP type", "Unknown DHCP type from API: "+string(dhcpSettings.DhcpType))
		return types.ObjectNull(tf.SiteNativeRangeDhcpResourceAttrTypes)
	}

	dhcpSettingsObj, objDiags := types.ObjectValueFrom(ctx, tf.SiteNativeRangeDhcpResourceAttrTypes, tfDhcpSettings)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(tf.SiteNativeRangeDhcpResourceAttrTypes)
	}
	return dhcpSettingsObj
}

// GetRelayGroupName looks up the DHCP relay group name based on the provided relay group ID.
func GetRelayGroupName(ctx context.Context, client SdkClienter, relayGroupID string, diags *diag.Diagnostics,
) (groupName *string) {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult := LookupRelayGroupID(ctx, client, diags)
	if diags.HasError() {
		return nil
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if item.Entity.GetID() == relayGroupID {
			return item.Entity.GetName()
		}
	}

	// Relay group not found
	diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("DHCP relay group: '%s' not found", relayGroupID))
	return nil
}

// LookupRelayGroupID looks up the DHCP relay groups
func LookupRelayGroupID(ctx context.Context, client SdkClienter, diags *diag.Diagnostics,
) *cato_go_sdk.EntityLookup {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult, err := client.V2().EntityLookupMinimal(ctx, client.AccountID(), cato_models.EntityTypeDhcpRelayGroup,
		nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("An error was encountered when looking up DHCP relay group: %v", err))
		return nil
	}
	return dhcpRelayGroupResult
}

// GetRelayGroupID looks up the DHCP relay group ID based on the provided relay group name.
func GetRelayGroupID(ctx context.Context, client SdkClienter, relayGroupName string, diags *diag.Diagnostics,
) (groupID string) {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult := LookupRelayGroupID(ctx, client, diags)
	if diags.HasError() {
		return ""
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
			return item.Entity.GetID()
		}
	}

	// Relay group not found
	diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("DHCP relay group: '%s' not found", relayGroupName))
	return ""
}
