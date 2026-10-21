package networkrange

import (
	"context"
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkrange"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func GetNetworkRangeValidator() NetworkRangeValidator {
	return NetworkRangeValidator{}
}

// NetworkRangeValidator validates the NetworkRange settings
type NetworkRangeValidator struct{}

func (v NetworkRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var networkRange tf.NetworkRange

	if !utils.HasValue(req.ConfigValue) {
		return
	}

	// get network range
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &networkRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if v.isUnknown(networkRange) {
		return
	}
	v.ValidateNetworkRange(ctx, &networkRange, &resp.Diagnostics)
}

func (v NetworkRangeValidator) ValidateNetworkRange(ctx context.Context, networkRange *tf.NetworkRange, diags *diag.Diagnostics) {
	v.validateNetworkRange(ctx, networkRange, nil, diags)
}

// ValidateNetworkRangeWithPriorState validates ModifyPlan input. Terraform can populate Optional+Computed
// attributes from prior state before resource ModifyPlan runs, so only values that differ from
// prior state are treated as user-explicit for mutually-exclusive interface fields.
func (v NetworkRangeValidator) ValidateNetworkRangeWithPriorState(
	ctx context.Context, networkRange *tf.NetworkRange, priorState *tf.NetworkRange, diags *diag.Diagnostics,
) {
	v.validateNetworkRange(ctx, networkRange, priorState, diags)
}

func (v NetworkRangeValidator) validateNetworkRange(
	ctx context.Context, networkRange *tf.NetworkRange, priorState *tf.NetworkRange, diags *diag.Diagnostics,
) {
	// Validate that local_ip is within network_network_range
	if networkrange.CheckLocalIP(diags, networkRange.LocalIP, networkRange.Subnet) != nil {
		return
	}

	// Validate DHCP settings
	if validators.DHCPChecker.CheckWithPriorState(ctx, diags, networkRange.DhcpSettings, priorStateDhcpSettings(priorState)) != nil {
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously
	idExplicit := validators.StringFieldIsExplicit(networkRange.InterfaceID, priorStateInterfaceID(priorState))
	indexExplicit := validators.StringFieldIsExplicit(networkRange.InterfaceIndex, priorStateInterfaceIndex(priorState))
	if idExplicit && indexExplicit {
		diags.AddError("Invalid network range Configuration",
			fmt.Sprintf("interface_id '%s' and interface_index '%s' cannot be set simultaneously.",
				networkRange.InterfaceID.ValueString(), networkRange.InterfaceIndex.ValueString()))
		return
	}

	// Validate that InternetOnly and MdnsReflector cannot be set simultaneously
	if utils.HasValue(networkRange.InternetOnly) && utils.HasValue(networkRange.MdnsReflector) &&
		networkRange.InternetOnly.ValueBool() && networkRange.MdnsReflector.ValueBool() {
		diags.AddError(
			"Invalid network range Configuration",
			"InternetOnly and MdnsReflector cannot be set simultaneously",
		)
		return
	}

	// Validate attributes based on rangeType
	if v.checkRangeTypeAttributes(diags, networkRange) != nil {
		return
	}
}

func priorStateInterfaceID(state *tf.NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceID
}

func priorStateInterfaceIndex(state *tf.NetworkRange) types.String {
	if state == nil {
		return types.StringNull()
	}
	return state.InterfaceIndex
}

func priorStateDhcpSettings(state *tf.NetworkRange) types.Object {
	if state == nil {
		return types.ObjectNull(tf.DhcpSettingsAttrTypes)
	}
	return state.DhcpSettings
}

func (v NetworkRangeValidator) checkRangeTypeAttributes(diags *diag.Diagnostics, networkRange *tf.NetworkRange) error {
	rangeType := cato_models.SubnetType(networkRange.RangeType.ValueString())

	isAllowed := func(diags *diag.Diagnostics, rangeType cato_models.SubnetType, attributeName string,
		attributeValue utils.HasValuer, allowedTypes []cato_models.SubnetType) bool {
		if !utils.HasValue(attributeValue) {
			return true
		}
		if slices.Contains(allowedTypes, rangeType) {
			return true
		}
		diags.AddError("Invalid network range Configuration",
			fmt.Sprintf("'%s' is not allowed for rangeType '%s'", attributeName, rangeType))
		return false
	}

	if !isAllowed(diags, rangeType, "local_ip", networkRange.LocalIP, []cato_models.SubnetType{
		cato_models.SubnetTypeDirect,
		cato_models.SubnetTypeNative,
		cato_models.SubnetTypeSecondaryNative,
		cato_models.SubnetTypeVlan}) {
		return apperror.ErrConfig
	}

	if !isAllowed(diags, rangeType, "dhcp_settings", networkRange.DhcpSettings, []cato_models.SubnetType{
		cato_models.SubnetTypeNative,
		cato_models.SubnetTypeSecondaryNative,
		cato_models.SubnetTypeVlan}) {
		return apperror.ErrConfig
	}

	if !isAllowed(diags, rangeType, "gateway", networkRange.Gateway, []cato_models.SubnetType{
		cato_models.SubnetTypeRouted}) {
		return apperror.ErrConfig
	}

	if !isAllowed(diags, rangeType, "vlan", networkRange.Vlan, []cato_models.SubnetType{
		cato_models.SubnetTypeVlan}) {
		return apperror.ErrConfig
	}

	// Validate that mDNS reflector is not set to true when rangeType is "Routed"
	if rangeType == cato_models.SubnetTypeRouted &&
		utils.HasValue(networkRange.MdnsReflector) && networkRange.MdnsReflector.ValueBool() {
		diags.AddError("Invalid network range configuration", "mDNS cannot be enabled when rangeType is 'Routed'")
		return apperror.ErrConfig
	}

	// Validate that vlan is set for "VLAN" ranges
	if !utils.HasValue(networkRange.Vlan) && rangeType == cato_models.SubnetTypeVlan {
		diags.AddError("Invalid network range configuration", "vlan number must be set for VLAN range type")
		return apperror.ErrConfig
	}

	return nil
}

func (v NetworkRangeValidator) Description(_ context.Context) string {
	return "network range settings validation" // TODO: provide more detailed description
}

func (v NetworkRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v NetworkRangeValidator) isUnknown(networkRange tf.NetworkRange) bool {
	return networkRange.DhcpSettings.IsUnknown() ||
		networkRange.Gateway.IsUnknown() ||
		networkRange.InterfaceID.IsUnknown() ||
		networkRange.InterfaceIndex.IsUnknown() ||
		networkRange.InternetOnly.IsUnknown() ||
		networkRange.MdnsReflector.IsUnknown() ||
		networkRange.LocalIP.IsUnknown() ||
		networkRange.Name.IsUnknown() ||
		networkRange.RangeType.IsUnknown() ||
		networkRange.SiteID.IsUnknown() ||
		networkRange.Subnet.IsUnknown() ||
		networkRange.TranslatedSubnet.IsUnknown() ||
		networkRange.Vlan.IsUnknown()
}

// ModifyPlan ensures that exactly one of interface_id or interface_index is set,
// and if one changes in the config, the other one is marked as unknown.
func (r *networkRangeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, cfg *tf.NetworkRange
	state := &tf.NetworkRange{} // avoid nil pointer dereference
	stateDefined := !req.State.Raw.IsNull()

	if req.Plan.Raw.IsNull() { // resource destruction
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Validate config - ensure there is exactly one interface_index or interface_id.
	nrValidator := NetworkRangeValidator{}
	nrValidator.ValidateNetworkRangeWithPriorState(ctx, cfg, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	// mdns reflector is only relevant for native, direct and vlan ranges
	if rangeType == cato_models.SubnetTypeRouted {
		plan.MdnsReflector = types.BoolNull()
		if utils.HasValue(cfg.MdnsReflector) {
			resp.Diagnostics.AddError("Invalid network range configuration",
				"mdns_reflector cannot be used when rangeType is 'Routed'")
			return
		}
	}

	// set interfaceIndex and interfaceID
	r.planInterfaceIDIndex(cfg, plan, state, stateDefined)

	// Gateway is only relevant for Routed range type
	plan.Gateway = types.StringNull()
	if rangeType == cato_models.SubnetTypeRouted {
		plan.Gateway = planmodifiers.DefaultPlanValue(cfg.Gateway, state.Gateway, stateDefined)
	}

	// Local IP is only relevant for Direct, Native and VLAN range types
	plan.LocalIP = types.StringNull()
	if rangeType != cato_models.SubnetTypeRouted {
		plan.LocalIP = planmodifiers.DefaultPlanValue(cfg.LocalIP, state.LocalIP, stateDefined)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// planInterfaceIDIndex - plan mutex interface fields
// - Ensures that exactly one of interface_id or interface_index is set in the plan.
// - If one changes in the config, the other one is marked as unknown.
// - If neither changes in the config, the known value from state is used in the plan (if available)
func (r *networkRangeResource) planInterfaceIDIndex(cfg, plan, state *tf.NetworkRange, stateDefined bool) {
	if !stateDefined {
		return
	}

	indexExplicit := validators.StringFieldIsExplicit(cfg.InterfaceIndex, state.InterfaceIndex)
	idExplicit := validators.StringFieldIsExplicit(cfg.InterfaceID, state.InterfaceID)

	// if interfaceIndex is configured and is different from the state, mark interfaceID as unknown
	if indexExplicit {
		plan.InterfaceIndex = cfg.InterfaceIndex
		plan.InterfaceID = types.StringUnknown()
		return
	}
	// if interfaceID is configured and is different from the state, mark interfaceIndex as unknown
	if idExplicit {
		plan.InterfaceID = cfg.InterfaceID
		plan.InterfaceIndex = types.StringUnknown()
		return
	}

	if utils.HasValue(cfg.InterfaceIndex) {
		plan.InterfaceIndex = cfg.InterfaceIndex
		plan.InterfaceID = state.InterfaceID
	}
	if utils.HasValue(cfg.InterfaceID) {
		plan.InterfaceID = cfg.InterfaceID
		plan.InterfaceIndex = state.InterfaceIndex
	}
}

// SubnetTypeValidator validates that the provided string is a valid subnet (range) type
type (
	SubnetTypeValidator struct{}
)

func (v SubnetTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	subnetType := cato_models.SubnetType(value)
	if !subnetType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid subnet type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllSubnetType))
		return
	}
}

func (v SubnetTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Subnet type must be one of: %v", cato_models.AllSubnetType)
}
func (v SubnetTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
