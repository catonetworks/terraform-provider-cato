package networkrange

import (
	"context"
	json "encoding/json"
	"fmt"
	"regexp"
	"strings"

	cato "github.com/catonetworks/cato-go-sdk"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &networkRangeResource{}
	_ resource.ResourceWithConfigure   = &networkRangeResource{}
	_ resource.ResourceWithImportState = &networkRangeResource{}
	_ resource.ResourceWithModifyPlan  = &networkRangeResource{}
)

var numberRE = regexp.MustCompile(`^\d+$`)

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
	client             *catoclient.Service
	networkRangeClient NetworkRangeClient

	apiClient ResourceClient
}

func (r *networkRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_range"
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *networkRangeResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
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
	networkRangeInput := hydrateAddNetworkRangeAPI(ctx, r.dhcpClient(), cfg, plan, &resp.Diagnostics)
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
	networkRange, err := r.resourceAPI().SiteAddNetworkRange(ctx, interfaceID, networkRangeInput, r.client.AccountId)
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

	input := hydrateUpdateNetworkRangeAPI(ctx, r.dhcpClient(), cfg, plan, &resp.Diagnostics)

	_, err := r.resourceAPI().SiteUpdateNetworkRange(ctx, plan.ID.ValueString(), input, r.client.AccountId)
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
	_, err := r.resourceAPI().SiteRemoveNetworkRange(ctx, state.ID.ValueString(), r.client.AccountId)
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
		dhcpSettingsObj = dhcp.ParseSettings(ctx, r.dhcpClient(), responseRange.DhcpSettings, diags)
	}
	if diags.HasError() {
		return tf.NetworkRange{}, true
	}

	newState := convertHydrateNetworkRangeStateValues(state, networkRangeID, responseRange, interfaceIndex, interfaceID, dhcpSettingsObj)

	return newState, true
}
