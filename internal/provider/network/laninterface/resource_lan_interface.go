package laninterface

import (
	"context"
	"encoding/json"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &lanInterfaceResource{}
	_ resource.ResourceWithConfigure = &lanInterfaceResource{}
)

const (
	lanLagMasterDestType        = "LAN_LAG_MASTER"
	lanLagMasterAndVrrpDestType = "LAN_LAG_MASTER_AND_VRRP"
	dhcpNativeRangeMessage      = "DHCP Range should be included in the native range"
)

func NewLanInterfaceResource() resource.Resource {
	return &lanInterfaceResource{}
}

type lanInterfaceResource struct {
	client *catoclient.Service
}

func (r *lanInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lan_interface"
}

func (r *lanInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *lanInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

//nolint:gocyclo,funlen,lll
func (r *lanInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var cfg, plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate LAG configuration
	destType := plan.DestType.ValueString()
	hasLagMinLinks := !plan.LagMinLinks.IsNull() && !plan.LagMinLinks.IsUnknown()

	// Rule 1: If dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
	if (destType == lanLagMasterDestType || destType == lanLagMasterAndVrrpDestType) && !hasLagMinLinks {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("When dest_type is %s, lag_min_links must be specified.", destType),
		)
		return
	}

	// Rule 2: If lag_min_links has a value, dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && destType != lanLagMasterDestType && destType != lanLagMasterAndVrrpDestType {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but dest_type is %s.", destType),
		)
		return
	}

	input := hydrateLanInterfaceAPI(ctx, cfg, plan)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(ctx, plan.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		reservedInterface := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			if apiError.GraphqlErrors[0].Message == dhcpNativeRangeMessage {
				reservedInterface = true
			}
		}
		if reservedInterface {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error()+"\n\nThe interfaceID "+plan.InterfaceID.ValueString()+" on this site type is reserved as a native range managed from the socket_site resource, and is unable to be modified from the cato_lan_interface resource.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
		}
		return
	}
	// Updating interface a second time due to API bug where only the name field does not propagate on first update intermittently.
	_, _ = r.client.Catov2.SiteUpdateSocketInterface(
		ctx,
		plan.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})

	// Resolve numeric interface ID by querying entityLookup
	siteEntity := &cato_models.EntityInput{Type: "site", ID: plan.SiteID.ValueString()}
	zeroInt64 := int64(0)
	queryInterfaceResult, err := r.client.Catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	// Find matching interface to get numeric ID
	tflog.Debug(ctx, "Iterate over network interfaces to match interfaceID '"+plan.InterfaceID.ValueString()+"' to retrieve numeric networkInterfaceID")
	for _, item := range queryInterfaceResult.GetEntityLookup().GetItems() {
		helperFields := item.GetHelperFields()
		interfaceID := cast.ToString(helperFields["interfaceId"])
		if _, err := cast.ToIntE(interfaceID); err == nil {
			interfaceID = fmt.Sprintf("INT_%v", interfaceID)
		}
		if interfaceID == plan.InterfaceID.ValueString() {
			plan.ID = types.StringValue(item.GetEntity().GetID())
			tflog.Debug(ctx, "Network interface matched! Setting plan.ID="+plan.ID.ValueString())
			break
		}
	}

	// Use hydration function to populate all state from API (including translated_subnet and local_ip from siteRange)
	updatedPlan, err := r.hydrateLanInterfaceState(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating interface state",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, updatedPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use hydration function to populate state from API
	updatedState, err := r.hydrateLanInterfaceState(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating interface state",
			err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:gocyclo,lll
func (r *lanInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate LAG configuration
	destType := plan.DestType.ValueString()
	hasLagMinLinks := !plan.LagMinLinks.IsNull() && !plan.LagMinLinks.IsUnknown()

	// Rule 1: If dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
	if (destType == lanLagMasterDestType || destType == lanLagMasterAndVrrpDestType) && !hasLagMinLinks {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("When dest_type is %s, lag_min_links must be specified.", destType),
		)
		return
	}

	// Rule 2: If lag_min_links has a value, dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && destType != lanLagMasterDestType && destType != lanLagMasterAndVrrpDestType {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but dest_type is %s.", destType),
		)
		return
	}

	input := hydrateLanInterfaceAPI(ctx, cfg, plan)
	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})
	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(ctx, plan.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		reservedInterface := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			if apiError.GraphqlErrors[0].Message == dhcpNativeRangeMessage {
				reservedInterface = true
			}
		}
		if reservedInterface {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error()+"\n\nThe interfaceID "+plan.InterfaceID.ValueString()+" on this site type is reserved as a native range managed from the socket_site resource, and is unable to be modified from the cato_lan_interface resource.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
		}
		return
	}

	// Use hydration function to populate state from API
	updatedPlan, err := r.hydrateLanInterfaceState(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating interface state",
			err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, updatedPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:gocyclo,gocritic,lll
func (r *lanInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if LAG_MASTER, lookup other LAG_MEMBER interfaces and disable to successfully delete LAG_MASTER
	if state.DestType.ValueString() == lanLagMasterDestType || state.DestType.ValueString() != lanLagMasterAndVrrpDestType {
		// Get the site's accountSnapshot to find the LAG master
		siteAccountSnapshotAPIData, err := r.client.Catov2.AccountSnapshot(ctx, []string{state.SiteID.ValueString()}, nil, &r.client.AccountId)
		tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteAccountSnapshotAPIData),
		})
		for _, site := range siteAccountSnapshotAPIData.AccountSnapshot.GetSites() {
			siteID := site.GetID()
			if siteID != nil && state.SiteID.ValueString() == *siteID {
				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					if iface.DestType != nil && *iface.DestType == lanLagMemberDestType {
						tflog.Debug(ctx, "Create.AccountSnapshot.response found LAN_LAG_MEMBER", map[string]interface{}{
							"response": utils.InterfaceToJSONString(iface),
						})
						curInterfaceID := iface.ID
						if _, err := cast.ToIntE(curInterfaceID); err == nil {
							curInterfaceID = fmt.Sprintf("INT_%v", curInterfaceID)
						}
						input := cato_models.UpdateSocketInterfaceInput{
							Name:     &curInterfaceID,
							DestType: "INTERFACE_DISABLED",
						}

						tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request LAN_LAG_MEMBER", map[string]interface{}{
							"request": utils.InterfaceToJSONString(input),
						})
						_, err := r.client.Catov2.SiteUpdateSocketInterface(ctx, state.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(curInterfaceID), input, r.client.AccountId)
						if err != nil {
							resp.Diagnostics.AddError(
								"Cato API SiteUpdateSocketInterface error",
								err.Error(),
							)
							return
						}
					}
				}
			}
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API error getting account snapshot for LAG member creation",
				err.Error(),
			)
			return
		}
	}

	// Disabled interface to "remove" an interface
	input := cato_models.UpdateSocketInterfaceInput{
		Name:     state.InterfaceID.ValueStringPointer(),
		DestType: "INTERFACE_DISABLED",
	}

	tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(ctx, state.SiteID.ValueString(), cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})

	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		reservedInterface := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			if apiError.GraphqlErrors[0].Message == "At least one LAN interface must be defined" {
				reservedInterface = true
			}
		}
		tflog.Debug(ctx, "Checking for reservedInterface of LAN on socket, if reservedInterface, gracefully failing deleting resource from state.", map[string]interface{}{
			"isReservedInterface": utils.InterfaceToJSONString(reservedInterface),
			"InterfaceId":         utils.InterfaceToJSONString(cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString())),
			"SiteId":              utils.InterfaceToJSONString(state.SiteID.ValueString()),
		})
		if !reservedInterface {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
	}
}

func hydrateLanInterfaceAPI(ctx context.Context, cfg, plan LanInterface) cato_models.UpdateSocketInterfaceInput {
	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"plan.DestType.String()": utils.InterfaceToJSONString(plan.DestType.ValueString()),
		"plan.Name.String()":     utils.InterfaceToJSONString(plan.Name.ValueStringPointer()),
		"cato_models.SocketInterfaceDestType(plan.DestType.String())": cato_models.SocketInterfaceDestType(plan.DestType.ValueString()),
		"plan": utils.InterfaceToJSONString(plan),
	})

	destType := plan.DestType.ValueString()
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}

	// For LAN_LAG_MEMBER, only send basic interface info - no LAN, LAG, or VRRP config
	if destType != lanLagMemberDestType {
		// Add LAN configuration for non-LAG member interfaces
		input.Lan = &cato_models.SocketInterfaceLanInput{
			Subnet:           plan.Subnet.ValueString(),
			TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
			LocalIP:          plan.LocalIP.ValueString(),
		}

		// Add VRRP configuration if specified and not LAG member
		if !plan.VrrpType.IsNull() {
			input.Vrrp = &cato_models.SocketInterfaceVrrpInput{
				VrrpType: (*cato_models.VrrpType)(plan.VrrpType.ValueStringPointer()),
			}
		}
	}

	// Add LAG configuration only for LAG master types
	if (destType == lanLagMasterDestType || destType == lanLagMasterAndVrrpDestType) &&
		!plan.LagMinLinks.IsNull() &&
		!plan.LagMinLinks.IsUnknown() {
		input.Lag = &cato_models.SocketInterfaceLagInput{
			MinLinks: plan.LagMinLinks.ValueInt64(),
		}
	}

	return input
}

//nolint:gocyclo,funlen,lll
func (r *lanInterfaceResource) hydrateLanInterfaceState(ctx context.Context, state *LanInterface) (LanInterface, error) {
	// Standard interface lookup (for non-LAG members or after LAG master validation)
	tflog.Debug(ctx, "hydrateLanInterfaceState()", map[string]interface{}{
		"state":    utils.InterfaceToJSONString(state),
		"state.ID": utils.InterfaceToJSONString(state.ID.ValueString()),
	})
	zeroInt64 := int64(0)
	thouInt64 := int64(1000)
	queryInterfaceResult, err := r.client.Catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), &thouInt64, &zeroInt64, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		return *state, fmt.Errorf("Catov2 API error: %s", err.Error())
	}
	isPresent := false
	for _, curIint := range queryInterfaceResult.EntityLookup.Items {
		// find the socket site entry we need
		tflog.Warn(ctx, "For.queryInterfaceResult.EntityLookup", map[string]interface{}{
			"curIint": utils.InterfaceToJSONString(curIint),
		})
		if curIint.Entity.ID == state.ID.ValueString() {
			tflog.Warn(ctx, "curIint.Entity.ID==state.ID.ValueString()", map[string]interface{}{
				"curIint.Entity.ID": utils.InterfaceToJSONString(curIint.Entity.ID),
			})
			isPresent = true
			state.ID = types.StringValue(curIint.Entity.ID)
			if siteIDVal, ok := curIint.HelperFields["siteId"]; ok {
				state.SiteID = types.StringValue(cast.ToString(siteIDVal))
			}
			if _, ok := curIint.HelperFields["interfaceID"]; ok {
				if idxInt, err := cast.ToIntE(curIint.HelperFields["interfaceID"]); err == nil {
					state.InterfaceID = types.StringValue(fmt.Sprintf("INT_%d", idxInt))
				} else {
					state.InterfaceID = types.StringValue(cast.ToString(curIint.HelperFields["interfaceID"]))
				}
			}
			if nameVal, ok := curIint.HelperFields["interfaceName"]; ok {
				state.Name = types.StringValue(cast.ToString(nameVal))
			}
			if destTypeVal, ok := curIint.HelperFields["destType"]; ok {
				state.DestType = types.StringValue(cast.ToString(destTypeVal))
			}
			if subnetVal, ok := curIint.HelperFields["subnet"]; ok {
				state.Subnet = types.StringValue(cast.ToString(subnetVal))
			}

			// Populate vrrp_type from networkInterface if available
			if vrrpTypeVal, ok := curIint.HelperFields["vrrpType"]; ok {
				vrrpTypeStr := cast.ToString(vrrpTypeVal)
				if vrrpTypeStr != "" {
					state.VrrpType = types.StringValue(vrrpTypeStr)
				} else {
					state.VrrpType = types.StringNull()
				}
			} else {
				state.VrrpType = types.StringNull()
			}

			// lag_min_links will be populated below if this is a LAG master interface
			break
		}
	}

	tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
		"state.LagMinLinks.IsUnknown()": utils.InterfaceToJSONString(state.LagMinLinks.IsUnknown()),
		"state.LagMinLinks.IsNull()":    utils.InterfaceToJSONString(state.LagMinLinks.IsNull()),
		"state.DestType.ValueString()":  utils.InterfaceToJSONString(state.DestType.ValueString()),
	})
	if (state.DestType.ValueString() == lanLagMasterDestType || state.DestType.ValueString() == lanLagMasterAndVrrpDestType) &&
		(state.LagMinLinks.IsNull() || state.LagMinLinks.IsUnknown()) {
		lagMinLinks := 0
		siteAccountSnapshotAPIData, err := r.client.Catov2.AccountSnapshot(ctx, []string{state.SiteID.ValueString()}, nil, &r.client.AccountId)
		tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteAccountSnapshotAPIData),
		})
		if err != nil {
			return *state, fmt.Errorf("Catov2 API error getting account snapshot for LAG member: %s", err.Error())
		}
		for _, site := range siteAccountSnapshotAPIData.AccountSnapshot.GetSites() {
			siteID := site.GetID()
			if siteID != nil && state.SiteID.ValueString() == *siteID {
				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					if iface.DestType != nil && *iface.DestType == lanLagMemberDestType {
						lagMinLinks++
					}
				}
			}
		}
		state.LagMinLinks = types.Int64Value(int64(lagMinLinks))
	}

	if !isPresent {
		tflog.Warn(ctx, "networkInterface not found, networkInterface resource removed")
		return *state, fmt.Errorf("interface not found")
	}

	// Query siteRange to get translated_subnet and local_ip (gateway)
	// These fields are not available in networkInterface entityLookup, but are in siteRange
	if !state.Subnet.IsNull() && !state.Subnet.IsUnknown() && state.Subnet.ValueString() != "" {
		siteEntity := &cato_models.EntityInput{Type: "site", ID: state.SiteID.ValueString()}
		zeroInt64 := int64(0)
		thouInt64 := int64(1000)
		querySiteRangeResult, err := r.client.Catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), &thouInt64, &zeroInt64, siteEntity, nil, nil, nil, nil, nil)
		tflog.Debug(ctx, "hydrateLanInterfaceState.EntityLookupSiteRange.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(querySiteRangeResult),
		})
		if err == nil {
			// Find the siteRange that matches this interface's subnet
			for _, v := range querySiteRangeResult.GetEntityLookup().GetItems() {
				curSubnet := ""
				if subnetVal, ok := v.GetHelperFields()["subnet"]; ok {
					curSubnet = cast.ToString(subnetVal)
				}

				if curSubnet == state.Subnet.ValueString() {
					// Populate translated_subnet from siteRange
					if translatedSubnetVal, ok := v.HelperFields["translatedSubnet"]; ok {
						translatedSubnetStr := cast.ToString(translatedSubnetVal)
						if translatedSubnetStr != "" && translatedSubnetStr != state.Subnet.ValueString() {
							state.TranslatedSubnet = types.StringValue(translatedSubnetStr)
						} else {
							state.TranslatedSubnet = types.StringNull()
						}
					} else {
						state.TranslatedSubnet = types.StringNull()
					}

					// Populate local_ip from gateway field in siteRange
					// For LAN interfaces, localIp is the same as gateway
					if gatewayVal, ok := v.HelperFields["gateway"]; ok && gatewayVal != nil {
						gatewayStr := cast.ToString(gatewayVal)
						if gatewayStr != "" {
							state.LocalIP = types.StringValue(gatewayStr)
						} else {
							state.LocalIP = types.StringNull()
						}
					} else {
						state.LocalIP = types.StringNull()
					}

					break
				}
			}
		} else {
			tflog.Warn(ctx, "Failed to query siteRange for translated_subnet and local_ip", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	return *state, nil
}
