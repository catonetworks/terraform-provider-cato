//nolint:lll
package provider

import (
	"context"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource              = &wanInterfaceResource{}
	_ resource.ResourceWithConfigure = &wanInterfaceResource{}
)

const (
	wanInterfaceIDParts                = 2
	wanInterfaceActiveNaturalOrder     = 1
	wanInterfacePassiveNaturalOrder    = 2
	wanInterfaceLastResortNaturalOrder = 3
)

func NewWanInterfaceResource() resource.Resource {
	return &wanInterfaceResource{}
}

type wanInterfaceResource struct {
	client *catoClientData
}

func (r *wanInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_interface"
}

func (r *wanInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wan_interface` resource contains the configuration parameters necessary to add a wan interface to a socket. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The WAN interface ID, which is a combination of the site ID and the interface ID (e.g., `site_id:interface_id`, 12345:INT_1). This is used to identify the WAN interface resource.",
				Required:    false,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface_id": schema.StringAttribute{
				Description: "The interface ID, which is a unique identifier for the WAN interface (e.g., `INT_1`, `INT_2`, etc.). This is used to identify the specific WAN interface resource.",
				Required:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "WAN interface name",
				Required:    true,
			},
			"upstream_bandwidth": schema.Int64Attribute{
				Description: "WAN interface upstream bandwidth",
				Required:    true,
			},
			"downstream_bandwidth": schema.Int64Attribute{
				Description: "WAN interface downstream bandwidth",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "WAN interface role (https://api.catonetworks.com/documentation/#definition-SocketInterfaceRole)",
				Required:    true,
			},
			"precedence": schema.StringAttribute{
				Description: "WAN interface precedence (https://api.catonetworks.com/documentation/#definition-SocketInterfacePrecedenceEnum)",
				Required:    true,
			},
			// "off_cloud": schema.SingleNestedAttribute{
			// 	Description: "Off Cloud configuration (https://support.catonetworks.com/hc/en-us/articles/4413265642257-Routing-Traffic-to-an-Off-Cloud-Link#heading-1)",
			// 	Required:    true,
			// 	Optional:    false,
			// 	Attributes: map[string]schema.Attribute{
			// 		"enabled": schema.BoolAttribute{
			// 			Description: "Attribute to define off cloud status (enabled or disabled)",
			// 			Required:    true,
			// 			Optional:    false,
			// 		},
			// 		"public_ip": schema.StringAttribute{
			// 			Required:    false,
			// 			Optional:    true,
			// 		},
			// 		"public_port": schema.StringAttribute{
			// 			Required:    false,
			// 			Optional:    true,
			// 		},
			// 	},
			// },
		},
	}
}

func (r *wanInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{State: resp.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

func (r *wanInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: "CATO",
		Name:     plan.Name.ValueStringPointer(),
		Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
			UpstreamBandwidth:   plan.UpstreamBandwidth.ValueInt64Pointer(),
			DownstreamBandwidth: plan.DownstreamBandwidth.ValueInt64Pointer(),
		},
		Wan: &cato_models.SocketInterfaceWanInput{
			Role:       cato_models.SocketInterfaceRole(plan.Role.ValueString()),
			Precedence: cato_models.SocketInterfacePrecedenceEnum(plan.Precedence.ValueString()),
		},
	}

	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	intID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteID.ValueString() + ":" + intID.ValueString())

	// Hydrate the state from the API to get the latest values including precedence
	hydratedState, interfaceExists, hydrateErr := r.hydrateWanInterfaceState(ctx, plan, siteID.ValueString(), intID.ValueString(), true)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating WAN interface state after create",
			hydrateErr.Error(),
		)
		return
	}
	if !interfaceExists {
		resp.Diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found after create", intID.ValueString()),
		)
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Split the ID to extract site_id and interface_id
	parts := strings.Split(state.ID.ValueString(), ":")
	if len(parts) != wanInterfaceIDParts {
		resp.Diagnostics.AddError(
			"Invalid WAN interface ID format",
			"Expected format 'site_id:interface_id', got: "+state.ID.ValueString(),
		)
		return
	}
	siteID := parts[0]
	interfaceID := parts[1]

	// Hydrate the state from the API
	hydratedState, interfaceExists, err := r.hydrateWanInterfaceState(ctx, state, siteID, interfaceID, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating WAN interface state during read",
			err.Error(),
		)
		return
	}

	if !interfaceExists {
		resp.Diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found in site %q", interfaceID, siteID),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: "CATO",
		Name:     plan.Name.ValueStringPointer(),
		Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
			UpstreamBandwidth:   plan.UpstreamBandwidth.ValueInt64Pointer(),
			DownstreamBandwidth: plan.DownstreamBandwidth.ValueInt64Pointer(),
		},
		Wan: &cato_models.SocketInterfaceWanInput{
			Role:       cato_models.SocketInterfaceRole(plan.Role.ValueString()),
			Precedence: cato_models.SocketInterfacePrecedenceEnum(plan.Precedence.ValueString()),
		},
	}

	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	intID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteID.ValueString() + ":" + intID.ValueString())

	// Hydrate the state from the API to get the latest values including precedence
	hydratedState, interfaceExists, hydrateErr := r.hydrateWanInterfaceState(ctx, plan, siteID.ValueString(), intID.ValueString(), true)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating WAN interface state after update",
			hydrateErr.Error(),
		)
		return
	}
	if !interfaceExists {
		resp.Diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found after update", intID.ValueString()),
		)
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteID.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Delete.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		// check if there is only one WAN interface & rewrite the input with default one
		accountSnapshotSite, err := r.client.accountSnapshot(ctx, []string{state.SiteID.ValueString()}, nil, true)
		tflog.Debug(ctx, "Delete.AccountSnapshot.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(accountSnapshotSite),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}

		var c = 0
		for _, item := range accountSnapshotSite.AccountSnapshot.Sites[0].InfoSiteSnapshot.Interfaces {
			if *item.DestType == "CATO" {
				c++
			}
		}

		var input cato_models.UpdateSocketInterfaceInput

		if (c >= 1) && (state.Role == types.StringValue("wan_1")) {
			bandwidth := int64(10)
			input = cato_models.UpdateSocketInterfaceInput{
				DestType: "CATO",
				Name:     state.InterfaceID.ValueStringPointer(),
				Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
					UpstreamBandwidth:   &bandwidth,
					DownstreamBandwidth: &bandwidth,
				},
				Wan: &cato_models.SocketInterfaceWanInput{
					Role:       cato_models.SocketInterfaceRole("wan_1"),
					Precedence: cato_models.SocketInterfacePrecedenceEnum("ACTIVE"),
				},
			}
		} else {
			// Disabled interface to "remove" an interface
			input = cato_models.UpdateSocketInterfaceInput{
				Name:     state.InterfaceID.ValueStringPointer(),
				DestType: "INTERFACE_DISABLED",
			}
		}

		tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(input),
		})
		_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, state.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()), input, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteUpdateSocketInterface error",
				err.Error(),
			)
			return
		}
	}
}

func normalizedInterfaceIDCandidates(rawID string) []string {
	id := strings.TrimSpace(rawID)
	if id == "" {
		return nil
	}

	candidatesMap := map[string]struct{}{
		id: {},
	}

	upper := strings.ToUpper(id)
	if strings.HasPrefix(upper, "INT_") {
		candidatesMap[strings.TrimPrefix(upper, "INT_")] = struct{}{}
	}
	if strings.HasPrefix(upper, "WAN") {
		suffix := strings.TrimPrefix(upper, "WAN")
		if suffix != "" {
			candidatesMap[suffix] = struct{}{}
			candidatesMap["INT_"+suffix] = struct{}{}
		}
	}

	// If the raw ID is numeric (e.g., "1"), include common aliases.
	isNumeric := true
	for _, ch := range id {
		if ch < '0' || ch > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		candidatesMap["INT_"+id] = struct{}{}
		candidatesMap["WAN"+id] = struct{}{}
	}

	candidates := make([]string, 0, len(candidatesMap))
	for candidate := range candidatesMap {
		candidates = append(candidates, strings.ToUpper(candidate))
	}

	return candidates
}

func wanInterfaceIDsMatch(resourceInterfaceID string, snapshotInterfaceID string, deviceInterfaceID string) bool {
	resourceCandidates := normalizedInterfaceIDCandidates(resourceInterfaceID)
	if len(resourceCandidates) == 0 {
		return false
	}

	otherCandidatesMap := map[string]struct{}{}
	for _, candidate := range normalizedInterfaceIDCandidates(snapshotInterfaceID) {
		otherCandidatesMap[candidate] = struct{}{}
	}
	for _, candidate := range normalizedInterfaceIDCandidates(deviceInterfaceID) {
		otherCandidatesMap[candidate] = struct{}{}
	}

	for _, candidate := range resourceCandidates {
		if _, ok := otherCandidatesMap[candidate]; ok {
			return true
		}
	}

	return false
}

func precedenceFromNaturalOrder(naturalOrder *int64) types.String {
	if naturalOrder == nil {
		return types.StringNull()
	}

	switch *naturalOrder {
	case wanInterfaceActiveNaturalOrder:
		return types.StringValue("ACTIVE")
	case wanInterfacePassiveNaturalOrder:
		return types.StringValue("PASSIVE")
	case wanInterfaceLastResortNaturalOrder:
		return types.StringValue("LAST_RESORT")
	default:
		return types.StringNull()
	}
}

func wanRoleFromInterfaceID(interfaceID string) string {
	switch strings.ToUpper(interfaceID) {
	case "WAN1", "INT_1":
		return "wan_1"
	case "WAN2", "INT_2":
		return "wan_2"
	case "WAN3", "INT_3":
		return "wan_3"
	case "WAN4", "INT_4":
		return "wan_4"
	default:
		return ""
	}
}

// hydrateWanInterfaceState fetches the current state of a WAN interface from the API
// and populates the state object with the latest values, including precedence mapping
//
//nolint:gocyclo
func (r *wanInterfaceResource) hydrateWanInterfaceState(
	ctx context.Context,
	state WanInterface,
	siteID string,
	interfaceID string,
	forceRefresh bool,
) (WanInterface, bool, error) {
	// Get accountSnapshot data to check if site exists and has interfaces
	siteAccountSnapshotData, err := r.client.accountSnapshot(ctx, []string{siteID}, nil, forceRefresh)
	tflog.Debug(ctx, "hydrateWanInterfaceState.AccountSnapshot.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotData),
	})
	if err != nil {
		return WanInterface{}, false, fmt.Errorf("catov2 API AccountSnapshot error: %w", err)
	}

	// Check if site exists
	if len(siteAccountSnapshotData.AccountSnapshot.Sites) == 0 {
		return state, false, nil
	}

	site := siteAccountSnapshotData.AccountSnapshot.Sites[0]

	// Find the interface in InfoSiteSnapshot.Interfaces
	foundInterfaceIndex := -1
	for i := range site.InfoSiteSnapshot.Interfaces {
		iface := site.InfoSiteSnapshot.Interfaces[i]
		if "INT_"+iface.ID == interfaceID || iface.ID == interfaceID {
			foundInterfaceIndex = i
			break
		}
	}

	if foundInterfaceIndex == -1 {
		return state, false, nil
	}

	foundInterface := site.InfoSiteSnapshot.Interfaces[foundInterfaceIndex]

	// Populate basic interface information
	if strings.HasPrefix(foundInterface.ID, "WAN") ||
		strings.HasPrefix(foundInterface.ID, "LTE") ||
		strings.HasPrefix(foundInterface.ID, "USB") {
		state.InterfaceID = types.StringValue(foundInterface.ID)
	} else {
		state.InterfaceID = types.StringValue("INT_" + foundInterface.ID)
	}
	state.SiteID = types.StringValue(siteID)
	state.ID = types.StringValue(siteID + ":" + state.InterfaceID.ValueString())
	state.Name = types.StringValue(*foundInterface.Name)
	state.UpstreamBandwidth = types.Int64Value(*foundInterface.UpstreamBandwidth)
	state.DownstreamBandwidth = types.Int64Value(*foundInterface.DownstreamBandwidth)
	if derivedRole := wanRoleFromInterfaceID(foundInterface.ID); derivedRole != "" {
		state.Role = types.StringValue(derivedRole)
	} else {
		state.Role = types.StringNull()
	}

	// Map precedence from naturalOrder in devices.interfaces
	// naturalOrder: 1 = ACTIVE, 2 = PASSIVE, 3 = LAST_RESORT
	if len(site.Devices) > 0 {
		for _, device := range site.Devices {
			for _, deviceIface := range device.Interfaces {
				// Match interface by ID - deviceIface.ID is a pointer to string
				deviceIfaceID := ""
				if deviceIface.ID != nil {
					deviceIfaceID = *deviceIface.ID
				}
				if wanInterfaceIDsMatch(state.InterfaceID.ValueString(), foundInterface.ID, deviceIfaceID) {
					mappedPrecedence := precedenceFromNaturalOrder(deviceIface.NaturalOrder)
					if !mappedPrecedence.IsNull() && !mappedPrecedence.IsUnknown() {
						state.Precedence = mappedPrecedence
					}
					break
				}
			}
		}
	}

	return state, true, nil
}
