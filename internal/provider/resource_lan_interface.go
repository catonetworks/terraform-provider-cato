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
						"INTERFACE_DISABLED", "LAN", "LAN_AND_HA", "LAN_LAG_MASTER", "LAN_LAG_MASTER_AND_VRRP", "LAN_LAG_MEMBER", "VRRP", "VRRP_AND_LAN",
					),
				},
			},
			"local_ip": schema.StringAttribute{
				Description: "Local IP address of the LAN interface",
				Required:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet of the LAN interface in CIDR notation",
				Required:    true,
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
	// Fail if LAN1, INT_5, INT_3
	var plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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

	entityInput := &cato_models.EntityInput{}
	entityInput.Type = cato_models.EntityTypeSite
	entityInput.ID = plan.SiteId.ValueString()
	siteResponse, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, entityInput, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteResponse),
		"len()":    len(siteResponse.GetEntityLookup().GetItems()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	tflog.Warn(ctx, "Interate over network interfaces from entityLookup to match name against defatult interfaceID (name) '"+*plan.InterfaceID.ValueStringPointer()+"' to retrieve numeric networkInterfaceID")
	for _, item := range siteResponse.GetEntityLookup().GetItems() {
		tflog.Debug(ctx, "siteResponse.GetEntityLookup().GetItems()", map[string]interface{}{
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
			break
		}
	}

	// // Set name to the correct value
	// tflog.Warn(ctx, "Set interface name to the correct value")
	// input.Name = plan.Name.ValueStringPointer()

	// tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
	// 	"request": utils.InterfaceToJSONString(input),
	// })
	// siteUpdateSocketInterfaceResponse, err = r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	// tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
	// 	"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	// })

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
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

func (r *lanInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if interface exists, else remove resource
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	// queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
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
		}
	}

	if !isPresent {
		tflog.Warn(ctx, "networkInterface not found, networkInterface resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
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
	// Setting the plan.ID to the previous state value, as this can not be retrieved reliably via API.
	planID := plan.ID
	plan.ID = planID
	diags = resp.State.Set(ctx, plan)
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

	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(plan.DestType.ValueString()),
		Name:     plan.Name.ValueStringPointer(),
		Lan: &cato_models.SocketInterfaceLanInput{
			Subnet:           plan.Subnet.ValueString(),
			TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
			LocalIP:          plan.LocalIp.ValueString(),
		},
	}
	if !plan.VrrpType.IsNull() {
		input.Vrrp = &cato_models.SocketInterfaceVrrpInput{
			VrrpType: (*cato_models.VrrpType)(plan.VrrpType.ValueStringPointer()),
		}
	}
	return input
}
