package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

var (
	_ resource.Resource              = &lanInterfaceResource{}
	_ resource.ResourceWithConfigure = &lanInterfaceResource{}
)

func NewLanInterfaceResource() resource.Resource {
	return &lanInterfaceResource{}
}

type lanInterfaceResource struct {
	client *catoClientData
}

func (r *lanInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lan_interface"
}

func (r *lanInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_lan_interface` resource contains the configuration parameters necessary to add a lan interface to a socket. ([physical socket physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Interface ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "LAN interface name",
				Required:    true,
			},
			"interface_id": schema.StringAttribute{
				Description: "SocketInterface available ids, INT_# stands for 1,2,3...12 supported ids (https://api.catonetworks.com/documentation/#definition-SocketInterfaceIDEnum)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"INT_1", "INT_2", "INT_3", "INT_4", "INT_5", "INT_6", "INT_7", "INT_8", "INT_9", "INT_10", "INT_11", "INT_12", "LAN1", "LAN2",
					),
				},
			},
			"dest_type": schema.StringAttribute{
				Description: "SocketInterface destination type (https://api.catonetworks.com/documentation/#definition-SocketInterfaceDestType)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"INTERFACE_DISABLED", "LAN", "LAN_AND_HA", "LAN_LAG_MASTER", "LAN_LAG_MASTER_AND_VRRP", "VRRP", "VRRP_AND_LAN",
					),
				},
			},
			"local_ip": schema.StringAttribute{
				Description: "Local IP address of the LAN interface",
				Required:    false,
				Optional:    true,
			},
			"lag_min_links": schema.Int64Attribute{
				Description: "Number of interfaces to include in the link aggregagtion, only relevant for LAN_LAG_MASTER, LAN_LAG_MASTER_AND_VRRP",
				Required:    false,
				Optional:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet of the LAN interface in CIDR notation",
				Required:    false,
				Optional:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Translated NAT subnet configuration",
				Required:    false,
				Optional:    true,
			},
			"vrrp_type": schema.StringAttribute{
				Description: "VRRP Type (https://api.catonetworks.com/documentation/#definition-VrrpType)",
				Required:    false,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DIRECT_LINK", "VIA_SWITCH"),
				},
			},
		},
	}
}

func (r *lanInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *lanInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *lanInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate LAG configuration
	destType := plan.DestType.ValueString()
	hasLagMinLinks := !plan.LagMinLinks.IsNull() && !plan.LagMinLinks.IsUnknown()

	// Rule 1: If dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
	if (destType == "LAN_LAG_MASTER" || destType == "LAN_LAG_MASTER_AND_VRRP") && !hasLagMinLinks {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("When dest_type is %s, lag_min_links must be specified.", destType),
		)
		return
	}

	// Rule 2: If lag_min_links has a value, dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && destType != "LAN_LAG_MASTER" && destType != "LAN_LAG_MASTER_AND_VRRP" {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but dest_type is %s.", destType),
		)
		return
	}

	input := hydrateLanInterfaceAPI(ctx, plan)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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
			if apiError.GraphqlErrors[0].Message == "DHCP Range should be included in the native range" {
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
	_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})

	// Not using hydrateLanInterfaceState fucntion looling up interface in create as we need to resolve the numeric interface ID as it is not in update API response
	siteEntity := &cato_models.EntityInput{Type: "site", ID: plan.SiteId.ValueString()}
	zeroInt64 := int64(0)
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	tflog.Warn(ctx, "Interate over network interfaces from entityLookup to match name against defatult interfaceID (name) '"+*plan.InterfaceID.ValueStringPointer()+"' to retrieve numeric networkInterfaceID")
	for _, item := range queryInterfaceResult.GetEntityLookup().GetItems() {
		tflog.Debug(ctx, "queryInterfaceResult.GetEntityLookup().GetItems()", map[string]interface{}{
			"item": utils.InterfaceToJSONString(item),
		})
		entityFields := item.GetEntity()
		helperFields := item.GetHelperFields()
		interfaceName := cast.ToString(helperFields["interfaceName"])
		interfaceId := cast.ToString(helperFields["interfaceId"])
		if _, err := cast.ToIntE(interfaceId); err == nil {
			interfaceId = fmt.Sprintf("INT_%v", interfaceId)
		}
		if plan.InterfaceID.ValueStringPointer() != nil && interfaceId == *plan.InterfaceID.ValueStringPointer() {
			tflog.Warn(ctx, "Network interface name matched! "+interfaceName+", setting plan.ID "+fmt.Sprintf("%v", entityFields.GetID()))
			plan.ID = types.StringValue(entityFields.GetID())
			if siteIdVal, ok := item.HelperFields["siteId"]; ok {
				plan.SiteId = types.StringValue(cast.ToString(siteIdVal))
			}
			if _, ok := item.HelperFields["interfaceId"]; ok {
				if idxInt, err := cast.ToIntE(item.HelperFields["interfaceId"]); err == nil {
					plan.InterfaceID = types.StringValue(fmt.Sprintf("INT_%d", idxInt))
				} else {
					plan.InterfaceID = types.StringValue(cast.ToString(item.HelperFields["interfaceId"]))
				}
			}
			if nameVal, ok := item.HelperFields["interfaceName"]; ok {
				plan.Name = types.StringValue(cast.ToString(nameVal))
			}
			if destTypeVal, ok := item.HelperFields["destType"]; ok {
				plan.DestType = types.StringValue(cast.ToString(destTypeVal))
			}
			if subnetVal, ok := item.HelperFields["subnet"]; ok {
				plan.Subnet = types.StringValue(cast.ToString(subnetVal))
			}
			// translatedSubnet is missing from API
			// localIp is missing from API
			// vrrpType is missing from API
			// lag_min_links is missing from API - preserve existing value from state
			break
		}
	}

	diags = resp.State.Set(ctx, plan)
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

func (r *lanInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate LAG configuration
	destType := plan.DestType.ValueString()
	hasLagMinLinks := !plan.LagMinLinks.IsNull() && !plan.LagMinLinks.IsUnknown()

	// Rule 1: If dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
	if (destType == "LAN_LAG_MASTER" || destType == "LAN_LAG_MASTER_AND_VRRP") && !hasLagMinLinks {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("When dest_type is %s, lag_min_links must be specified.", destType),
		)
		return
	}

	// Rule 2: If lag_min_links has a value, dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && destType != "LAN_LAG_MASTER" && destType != "LAN_LAG_MASTER_AND_VRRP" {
		resp.Diagnostics.AddError(
			"Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but dest_type is %s.", destType),
		)
		return
	}

	input := hydrateLanInterfaceAPI(ctx, plan)
	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})
	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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
			if apiError.GraphqlErrors[0].Message == "DHCP Range should be included in the native range" {
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

func (r *lanInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state LanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if LAG_MASTER, lookup other LAG_MEMBER interfaces and disable to successfully delete LAG_MASTER
	if state.DestType.ValueString() == "LAN_LAG_MASTER" || state.DestType.ValueString() != "LAN_LAG_MASTER_AND_VRRP" {
		// Get the site's accountSnapshot to find the LAG master
		siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{state.SiteId.ValueString()}, nil, &r.client.AccountId)
		tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
		})
		for _, site := range siteAccountSnapshotApiData.AccountSnapshot.GetSites() {
			siteID := site.GetID()
			if siteID != nil && state.SiteId.ValueString() == *siteID {
				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					if iface.DestType != nil && *iface.DestType == "LAN_LAG_MEMBER" {
						tflog.Debug(ctx, "Create.AccountSnapshot.response found LAN_LAG_MEMBER", map[string]interface{}{
							"response": utils.InterfaceToJSONString(iface),
						})
						curInterfaceId := iface.ID
						if _, err := cast.ToIntE(curInterfaceId); err == nil {
							curInterfaceId = fmt.Sprintf("INT_%v", curInterfaceId)
						}
						input := cato_models.UpdateSocketInterfaceInput{
							Name:     &curInterfaceId,
							DestType: "INTERFACE_DISABLED",
						}

						tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request LAN_LAG_MEMBER", map[string]interface{}{
							"request": utils.InterfaceToJSONString(input),
						})
						_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, state.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(curInterfaceId), input, r.client.AccountId)
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
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, state.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()), input, r.client.AccountId)
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
			"InterfaceID":         utils.InterfaceToJSONString(cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString())),
			"SiteId":              utils.InterfaceToJSONString(state.SiteId.ValueString()),
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

func hydrateLanInterfaceAPI(ctx context.Context, plan LanInterface) cato_models.UpdateSocketInterfaceInput {
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
	if destType != "LAN_LAG_MEMBER" {
		// Add LAN configuration for non-LAG member interfaces
		input.Lan = &cato_models.SocketInterfaceLanInput{
			Subnet:           plan.Subnet.ValueString(),
			TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
			LocalIP:          plan.LocalIp.ValueString(),
		}

		// Add VRRP configuration if specified and not LAG member
		if !plan.VrrpType.IsNull() {
			input.Vrrp = &cato_models.SocketInterfaceVrrpInput{
				VrrpType: (*cato_models.VrrpType)(plan.VrrpType.ValueStringPointer()),
			}
		}
	}

	// Add LAG configuration only for LAG master types
	if (destType == "LAN_LAG_MASTER" || destType == "LAN_LAG_MASTER_AND_VRRP") && !plan.LagMinLinks.IsNull() && !plan.LagMinLinks.IsUnknown() {
		input.Lag = &cato_models.SocketInterfaceLagInput{
			MinLinks: plan.LagMinLinks.ValueInt64(),
		}
	}

	return input
}

func (r *lanInterfaceResource) hydrateLanInterfaceState(ctx context.Context, state *LanInterface) (LanInterface, error) {
	// Standard interface lookup (for non-LAG members or after LAG master validation)
	tflog.Debug(ctx, "hydrateLanInterfaceState()", map[string]interface{}{
		"state":    utils.InterfaceToJSONString(state),
		"state.ID": utils.InterfaceToJSONString(state.ID.ValueString()),
	})
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
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
			if siteIdVal, ok := curIint.HelperFields["siteId"]; ok {
				state.SiteId = types.StringValue(cast.ToString(siteIdVal))
			}
			if _, ok := curIint.HelperFields["interfaceId"]; ok {
				if idxInt, err := cast.ToIntE(curIint.HelperFields["interfaceId"]); err == nil {
					state.InterfaceID = types.StringValue(fmt.Sprintf("INT_%d", idxInt))
				} else {
					state.InterfaceID = types.StringValue(cast.ToString(curIint.HelperFields["interfaceId"]))
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
			// translatedSubnet is missing from API
			// localIp is missing from API
			// vrrpType is missing from API
			// lag_min_links is missing from API - preserve existing value from state
			break
		}
	}

	tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
		"state.LagMinLinks.IsUnknown()": utils.InterfaceToJSONString(state.LagMinLinks.IsUnknown()),
		"state.LagMinLinks.IsNull()":    utils.InterfaceToJSONString(state.LagMinLinks.IsNull()),
		"state.DestType.ValueString()":  utils.InterfaceToJSONString(state.DestType.ValueString()),
	})
	if (state.DestType.ValueString() == "LAN_LAG_MASTER" || state.DestType.ValueString() == "LAN_LAG_MASTER_AND_VRRP") && (state.LagMinLinks.IsNull() || state.LagMinLinks.IsUnknown()) {
		lagMinLinks := 0
		siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{state.SiteId.ValueString()}, nil, &r.client.AccountId)
		tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
		})
		if err != nil {
			return *state, fmt.Errorf("Catov2 API error getting account snapshot for LAG member: %s", err.Error())
		}
		for _, site := range siteAccountSnapshotApiData.AccountSnapshot.GetSites() {
			siteID := site.GetID()
			if siteID != nil && state.SiteId.ValueString() == *siteID {
				for _, iface := range site.InfoSiteSnapshot.Interfaces {
					if iface.DestType != nil && *iface.DestType == "LAN_LAG_MEMBER" {
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

	return *state, nil
}
