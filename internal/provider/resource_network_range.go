package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &networkRangeResource{}
	_ resource.ResourceWithConfigure   = &networkRangeResource{}
	_ resource.ResourceWithImportState = &networkRangeResource{}
	_ resource.ResourceWithModifyPlan  = &networkRangeResource{}
)

const (
	networkRangeDescription = "The `cato_network_range` resource contains the configuration parameters necessary to " +
		"add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket]" +
		"(https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). " +
		"Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()]" +
		"(https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange)."
	networkRangeMDNSReflectorDescription = "Site native range mDNS reflector. When enabled, the Socket functions as an " +
		"mDNS gateway, it relays mDNS requests and response between all enabled subnets."
	networkRangeDHCPMicrosegmentationDescription = "DHCP Microsegmentation. When enabled, the DHCP server will allocate " +
		"/32 subnet mask. Make sure to enable the proper Firewall rules and enable it with caution, as it is not supported " +
		"on all operating systems; monitor the network closely after activation. This setting can only be configured when " +
		"dhcp_type is set to DHCP_RANGE."
	networkRangeDHCPDisabledError = "When dhcp_type is DHCP_DISABLED, dhcp_ip_range, dhcp_relay_group_id, and " +
		"dhcp_relay_group_name must be null, unset, or empty strings."
	networkRangeDHCPRangeError = "When dhcp_type is DHCP_RANGE, dhcp_ip_range must be provided (not null, unset, or " +
		"empty string), and dhcp_relay_group_id and dhcp_relay_group_name must be null, unset, or empty strings."
)

func NewNetworkRangeResource() resource.Resource {
	return &networkRangeResource{}
}

type networkRangeResource struct {
	client             *catoClientData
	networkRangeClient NetworkRangeClient
}

type NetworkRangeClient interface {
	SiteAddNetworkRange(ctx context.Context, lanSocketInterfaceID string, addNetworkRangeInput cato_models.AddNetworkRangeInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato.SiteAddNetworkRange, error)
	SiteUpdateNetworkRange(ctx context.Context, networkRangeID string, updateNetworkRangeInput cato_models.UpdateNetworkRangeInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato.SiteUpdateNetworkRange, error)
	SiteRemoveNetworkRange(ctx context.Context, networkRangeID string, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.SiteRemoveNetworkRange, error)
	NetworkRange(ctx context.Context, accountID string, networkRangeID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkRange, error)
	EntityLookup(ctx context.Context, accountID string, typeArg cato_models.EntityType, limit *int64, from *int64,
		parent *cato_models.EntityInput, search *string, entityIDs []string, sort []*cato_models.SortInput,
		filters []*cato_models.LookupFilterInput, helperFields []string, interceptors ...clientv2.RequestInterceptor) (*cato.EntityLookup, error)
}

func (r *networkRangeResource) getNetworkRangeClient() NetworkRangeClient {
	if r.networkRangeClient != nil {
		return r.networkRangeClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
}

func (r *networkRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_range"
}

// Schema defines the schema for the network range resource
func (r *networkRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: networkRangeDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Range ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway": schema.StringAttribute{
				Description: "Network range gateway (Only releveant for Routed range_type)",
				Optional:    true,
				Computed:    true,
			},
			"interface_id": schema.StringAttribute{
				Description: "Network Interface ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"interface_index": schema.StringAttribute{
				Description:   "Network Interface Index",
				Required:      false,
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{validators.SocketInterfaceIndexValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"internet_only": schema.BoolAttribute{
				Description:   "Internet only network range (Only releveant for Routed range_type)",
				Computed:      true,
				Optional:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
				Computed:    true,
			},
			"mdns_reflector": schema.BoolAttribute{
				Description:   networkRangeMDNSReflectorDescription,
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Network range name",
				Required:    true,
			},
			"range_type": schema.StringAttribute{
				Description:   "Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)",
				Required:      true,
				Validators:    []validator.String{validators.SubnetTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"site_id": schema.StringAttribute{
				Description:   "Site ID",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description:   "Network range translated native IP range (CIDR)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_settings": dhcp.SchemaDhcpSettings(true),
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
				Optional:    true,
			},
		},
	}
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

// ModifyPlan ensures that exactlu one of interface_id or interface_index is set,
// and if one changes in the config, the other one is markerd as unknown.
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

	// Load state early so the validator can distinguish user-set values from values that
	// Terraform Core propagates from prior state for Optional+Computed attributes.
	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Validate config - ensure there is exactly one interface_index or interface_id.
	// Pass the prior state so the validator can skip state-propagated values.
	nrValidator := validators.NetworkRangeValidator{}
	nrValidator.ValidateNetworkRange(ctx, cfg, state, &resp.Diagnostics)
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
		plan.Gateway = defaultPlanValue(cfg.Gateway, state.Gateway, stateDefined)
	}

	// Local IP is only relevant for Direct, Native and VLAN range types
	plan.LocalIP = types.StringNull()
	if rangeType != cato_models.SubnetTypeRouted {
		plan.LocalIP = defaultPlanValue(cfg.LocalIP, state.LocalIP, stateDefined)
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

	// if interfaceIndex is configured and is different from the state, mark interfaceID as unknown
	if !cfg.InterfaceIndex.IsNull() {
		plan.InterfaceIndex = cfg.InterfaceIndex
		plan.InterfaceID = types.StringUnknown()
		// if the configured interfaceIndex is the same as in the state, use known ID value (if available)
		if stateDefined && utils.HasValue(state.InterfaceIndex) &&
			state.InterfaceIndex.ValueString() == cfg.InterfaceIndex.ValueString() {
			plan.InterfaceID = state.InterfaceID
		}
	}
	// if interfaceID is configured and is different from the state, mark interfaceIndex as unknown
	if !cfg.InterfaceID.IsNull() {
		plan.InterfaceID = cfg.InterfaceID
		plan.InterfaceIndex = types.StringUnknown()
		// if the configured interfaceID is the same as in the state, use known Index value (if available)
		if stateDefined && utils.HasValue(state.InterfaceID) &&
			state.InterfaceID.ValueString() == cfg.InterfaceID.ValueString() {
			plan.InterfaceIndex = state.InterfaceIndex
		}
	}
}

// defaultPlanValue returns the appropriate plan value: cfg -> state -> unknown
func defaultPlanValue(cfgValue, stateValue types.String, stateDefined bool) (planValue types.String) {
	newValue := types.StringUnknown()
	if utils.HasValue(cfgValue) {
		newValue = cfgValue
	} else if stateDefined && utils.HasValue(stateValue) {
		newValue = stateValue
	}
	return newValue
}

func (r *networkRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Hydrate the state from the API
	var state tf.NetworkRange
	state.ID = types.StringValue(req.ID)

	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, nil, &state, req.ID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !rangeExists {
		resp.Diagnostics.AddError(
			"Network Range Not Found",
			fmt.Sprintf("Network range with ID %q not found during import", req.ID),
		)
		return
	}

	// Set the hydrated state
	diags := resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Create the network range resource
func (r *networkRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var cfg, plan *tf.NetworkRange

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API input from the plan
	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())
	networkRangeInput := &cato_models.AddNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, r.client, rangeType, plan.DhcpSettings, &resp.Diagnostics),
		Gateway:          parse.KnownStringPointer(plan.Gateway),
		InternetOnly:     parse.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          parse.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    parse.KnownBoolPointer(plan.MdnsReflector),
		Name:             plan.Name.ValueString(),
		RangeType:        cato_models.SubnetType(plan.RangeType.ValueString()),
		Subnet:           plan.Subnet.ValueString(),
		TranslatedSubnet: translatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             parse.KnownInt64Pointer(plan.Vlan),
	}
	// TODO: check if !isHA { // for HA scenario, local IP is not allowed to be modified

	if resp.Diagnostics.HasError() {
		return
	}

	// Get interface ID, If only interface_index is provided, fetch the ID from the API
	interfaceID := plan.InterfaceID.ValueString()
	if !utils.HasValue(plan.InterfaceID) {
		interfaceID = r.getInterfaceID(ctx, plan.SiteID.ValueString(), plan.InterfaceIndex.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create the network range via API
	networkRange, err := r.getNetworkRangeClient().SiteAddNetworkRange(ctx, interfaceID, *networkRangeInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API SiteAddNetworkRange Error", err.Error())
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, cfg, plan,
		networkRange.Site.AddNetworkRange.NetworkRangeID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !rangeExists {
		tflog.Warn(ctx, "network range not found, network range resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read the network range resource
func (r *networkRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *tf.NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, nil, state, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update the network range resource
func (r *networkRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan *tf.NetworkRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())
	input := cato_models.UpdateNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, r.client, rangeType, plan.DhcpSettings, &resp.Diagnostics),
		Gateway:          parse.KnownStringPointer(plan.Gateway),
		InternetOnly:     parse.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          parse.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    parse.KnownBoolPointer(plan.MdnsReflector),
		Name:             parse.KnownStringPointer(plan.Name),
		RangeType:        (*cato_models.SubnetType)(parse.KnownStringPointer(plan.RangeType)),
		Subnet:           parse.KnownStringPointer(plan.Subnet),
		TranslatedSubnet: translatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             parse.KnownInt64Pointer(plan.Vlan),
	}

	_, err := r.getNetworkRangeClient().SiteUpdateNetworkRange(ctx, plan.ID.ValueString(), input, r.client.AccountId)
	if err != nil {
		var apiError cato.RespErrors
		interfaceNotPresent := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphQLErrors) > 0 {
			msg := apiError.GraphQLErrors[0].Message
			if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
				interfaceNotPresent = true
			}
		}
		if !interfaceNotPresent {
			resp.Diagnostics.AddError("Catov2 API error", err.Error())
			return
		}
		// If the network range is not present, delete the resource and recreate it
		tflog.Warn(ctx, "Network range not found during update, recreating resource")
		// Remove the resource from state
		resp.State.RemoveResource(ctx)
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, cfg, plan, plan.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete the network range resource
func (r *networkRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tf.NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if interface is already removed and fail gracefully
	//	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
	_, err := r.getNetworkRangeClient().SiteRemoveNetworkRange(ctx, state.ID.ValueString(), r.client.AccountId)
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		interfaceNotPresent := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			msg := apiError.GraphqlErrors[0].Message
			if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
				interfaceNotPresent = true
			}
		}
		if !interfaceNotPresent {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
	}
}

// hydrateNetworkRangeState populates the NetworkRange state with data from API responses
func (r *networkRangeResource) hydrateNetworkRangeState(ctx context.Context, cfg, state *tf.NetworkRange,
	networkRangeID string, diags *diag.Diagnostics,
) (netRange tf.NetworkRange, rangeExists bool) {
	// fetch network range details from API, handle 'not found'
	responseRange := r.fetchNetworkRange(ctx, networkRangeID, diags)
	if responseRange == nil {
		return tf.NetworkRange{}, false
	}

	// fetch interface details from API - either by interface_id or interface_index depending on what's available in config
	interfaceIndex, interfaceID := state.InterfaceIndex.ValueString(), state.InterfaceID.ValueString()
	if cfg != nil {
		if utils.HasValue(cfg.InterfaceID) {
			interfaceIndex = r.getInterfaceIndex(ctx, cfg.SiteID.ValueString(), interfaceID, diags)
		} else {
			interfaceID = r.getInterfaceID(ctx, cfg.SiteID.ValueString(), interfaceIndex, diags)
		}
		if diags.HasError() {
			return tf.NetworkRange{}, true
		}
	}

	// DHCP settings
	isDhcpSettingsDefault := r.checkDhcpSettingsDefault(ctx, cfg, state, diags)
	dhcpSettingsObj := dhcp.SettingsDefault(ctx, diags)
	if responseRange.DhcpSettings != nil && !isDhcpSettingsDefault {
		dhcpSettingsObj = dhcp.ParseSettings(ctx, r.client, responseRange.DhcpSettings, diags)
	}
	if diags.HasError() {
		return tf.NetworkRange{}, true
	}

	newState := tf.NetworkRange{
		ID:               types.StringValue(networkRangeID),
		DhcpSettings:     dhcpSettingsObj,
		Gateway:          types.StringPointerValue(responseRange.Gateway),
		InterfaceID:      types.StringValue(interfaceID),
		InterfaceIndex:   types.StringValue(interfaceIndex),
		InternetOnly:     types.BoolValue(responseRange.InternetOnly),
		MdnsReflector:    types.BoolValue(responseRange.MdnsReflector),
		LocalIP:          types.StringPointerValue(responseRange.LocalIP), // TODO: HA
		Name:             types.StringValue(responseRange.Name),
		RangeType:        types.StringValue(string(responseRange.RangeType)),
		SiteID:           state.SiteID,
		Subnet:           types.StringValue(responseRange.Subnet),
		TranslatedSubnet: types.StringPointerValue(responseRange.TranslatedSubnet),
		Vlan:             types.Int64PointerValue(responseRange.Vlan),
	}

	if responseRange.RangeType != cato_models.SubnetTypeVlan {
		newState.Vlan = types.Int64Null()
	}
	if state.MdnsReflector.IsNull() {
		newState.MdnsReflector = types.BoolNull()
	}
	if !state.Gateway.IsUnknown() {
		newState.Gateway = state.Gateway
	}

	return newState, true
}

// fetchNetworkRange retrieves the network range details from the API, and returns nil if the network range is not found,
// on error the diags gets updated.
func (r *networkRangeResource) fetchNetworkRange(ctx context.Context, networkRangeID string, diags *diag.Diagnostics,
) (responseRange *cato.NetworkRange_Site_NetworkRange) {
	const notFoundMsg = "Invalid network range id: "

	queryRangeResult, err := r.getNetworkRangeClient().NetworkRange(ctx, r.client.AccountId, networkRangeID)
	if err != nil {
		// Check if error is not found error, if so return (nil, false, nil) to indicate resource should be removed from state without error
		if gqlError, ok := errors.AsType[*clientv2.ErrorResponse](err); ok {
			if (gqlError.GqlErrors != nil) && (len(*gqlError.GqlErrors) > 0) && strings.Contains((*gqlError.GqlErrors)[0].Message, notFoundMsg) {
				return nil
			}
		}
		diags.AddError(
			"Catov2 NetworkRange API error",
			fmt.Sprintf("error fetching network range details for range ID '%s': %v", networkRangeID, err),
		)
		return nil
	}
	if queryRangeResult == nil || queryRangeResult.Site.NetworkRange == nil {
		return nil
	}
	return queryRangeResult.Site.NetworkRange
}

// getSiteIdFromNetworkRange retrieves the site_id and interface info for a network range using entityLookup
func (r *networkRangeResource) getSiteIDFromNetworkRange(ctx context.Context, networkRangeID string,
) (siteID string, interfaceName string, err error) {
	result, err := r.getNetworkRangeClient().EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"),
		nil, nil, nil, nil, []string{networkRangeID}, nil, nil, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup network range site: %w", err)
	}

	if result == nil || len(result.EntityLookup.GetItems()) == 0 {
		return "", "", fmt.Errorf("network range %s not found in entityLookup", networkRangeID)
	}

	item := result.EntityLookup.GetItems()[0]
	helperFields := item.GetHelperFields()
	if helperFields == nil {
		return "", "", fmt.Errorf("no helperFields returned for network range %s", networkRangeID)
	}

	// Extract siteID from helperFields
	siteID = cast.ToString(helperFields["siteId"])
	if siteID == "" {
		return "", "", fmt.Errorf("siteId not found in helperFields for network range %s", networkRangeID)
	}

	// Extract interfaceName from helperFields
	interfaceName = cast.ToString(helperFields["interfaceName"])

	return siteID, interfaceName, nil
}

// getInterfaceIndex looks up the network interface Index for a given site ID and interface ID
func (r *networkRangeResource) getInterfaceIndex(ctx context.Context, siteID, interfaceID string, diags *diag.Diagnostics,
) (interfaceIndex string) {
	site := &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	networkInterfaceResponse, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityTypeNetworkInterface,
		ptr(int64(0)), nil, site, nil, []string{interfaceID}, nil, nil, nil)
	if err != nil {
		diags.AddError("Error retrieving network interface "+interfaceID, err.Error())
		return ""
	}
	for _, item := range networkInterfaceResponse.GetEntityLookup().GetItems() {
		if item.GetEntity().GetID() == interfaceID {
			helperFields := item.GetHelperFields()
			curIfaceIndex := cast.ToString(helperFields["interfaceId"])
			if numberRE.MatchString(curIfaceIndex) {
				curIfaceIndex = "INT_" + curIfaceIndex
			}
			return curIfaceIndex
		}
	}
	diags.AddError("Error retrieving network interface",
		"network interface with id '"+interfaceID+"' not found in site '"+siteID+"'")
	return ""
}

// getInterfaceID looks up the network interface ID for a given site ID and interface index
func (r *networkRangeResource) getInterfaceID(ctx context.Context, siteID, interfaceIndex string, diags *diag.Diagnostics,
) (interfaceID string) {
	site := &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	networkInterfaceResponse, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityTypeNetworkInterface, ptr(int64(0)), nil, site, nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Error retrieving network interface "+interfaceIndex, err.Error())
		return ""
	}
	for _, item := range networkInterfaceResponse.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		curIfaceIndex := cast.ToString(helperFields["interfaceId"])
		if numberRE.MatchString(curIfaceIndex) {
			curIfaceIndex = "INT_" + curIfaceIndex
		}
		if curIfaceIndex == interfaceIndex {
			return item.GetEntity().GetID()
		}
	}
	diags.AddError("Error retrieving network interface",
		"network interface with index '"+interfaceIndex+"' not found in site '"+siteID+"'")
	return ""
}

// checkDhcpSettingsDefault checks DHCP settings config or state,
// If config is provided,
// return true if it is not defined or if dhcp_type is set to ACCOUNT_DEFAULT, false otherwise
// If config is not provided (nil), check the state with the same logic to determine if DHCP settings are default
//
// Reason: API does not return ACCOUNT_DEFAULT, but some other value based on CMA account configuration
func (r *networkRangeResource) checkDhcpSettingsDefault(ctx context.Context, cfg, state *tf.NetworkRange,
	diags *diag.Diagnostics,
) (isDhcpSettingsDefault bool) {
	// cfg defined -> Create/Update flow; check the config value
	if cfg != nil {
		if !utils.HasValue(cfg.DhcpSettings) {
			return false
		}
		var cfgDhcpSettings tf.DhcpSettings
		if utils.CheckErr(diags, cfg.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{})) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if state == nil || !utils.HasValue(state.DhcpSettings) {
		return false
	}

	var stateDhcpSettings tf.DhcpSettings
	if utils.CheckErr(diags, state.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
		return true
	}
	return false
}
