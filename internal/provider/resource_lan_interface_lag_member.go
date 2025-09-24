package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

var (
	_ resource.Resource              = &lanInterfaceLagMemberResource{}
	_ resource.ResourceWithConfigure = &lanInterfaceLagMemberResource{}
)

func NewLanInterfaceLagMemberResource() resource.Resource {
	return &lanInterfaceLagMemberResource{}
}

type lanInterfaceLagMemberResource struct {
	client *catoClientData
}

func (r *lanInterfaceLagMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lan_interface_lag_member"
}

func (r *lanInterfaceLagMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `catolan_interface_lab_member` resource contains the configuration parameters necessary to add a lan interface lan member to a socket lag interface (LAN_LAG_MASTER, or LAN_LAG_MASTER_AND_VRRP). ([physical socket physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The LAN LAG Member interface ID, which is a combination of the numeric interface ID and the interface Index (e.g., `site_id:interface_index`, 12345:INT_5). This is used to identify the primary LAN Master interface resource that the lag member is a part of.",
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
				Description: "LAN LAG member interface name",
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
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"LAN_LAG_MEMBER",
					),
				},
				Default: stringdefault.StaticString("LAN_LAG_MEMBER"),
			},
		},
	}
}

func (r *lanInterfaceLagMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *lanInterfaceLagMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *lanInterfaceLagMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LanInterfaceLagMember
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	destType := plan.DestType.ValueString()
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLagMember.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLagMember.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	plan.ID = types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID + ":" + string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID))

	// Use the new hydration function to populate the state
	plan, err = r.hydrateLanInterfaceLagMemberState(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating LAG member state",
			err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanInterfaceLagMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanInterfaceLagMember
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the new hydration function to populate the state
	hydratedState, err := r.hydrateLanInterfaceLagMemberState(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating LAG member state",
			err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanInterfaceLagMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan LanInterfaceLagMember
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	destType := plan.DestType.ValueString()
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLag.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLag.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	plan.ID = types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID + ":" + string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID))

	// Use the new hydration function to populate the state
	plan, err = r.hydrateLanInterfaceLagMemberState(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating LAG member state",
			err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanInterfaceLagMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state LanInterfaceLagMember
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

func (r *lanInterfaceLagMemberResource) hydrateLanInterfaceLagMemberState(ctx context.Context, input LanInterfaceLagMember) (LanInterfaceLagMember, error) {
	// Split the ID to extract site_id and interface_id
	parts := strings.Split(input.ID.ValueString(), ":")
	if len(parts) != 2 {
		return input, fmt.Errorf("Invalid LAN LAG Member interface ID format: Expected format 'site_id:interface_index', got: %s", input.ID.ValueString())
	}
	siteID := parts[0]
	interfaceID := parts[1]
	tflog.Debug(ctx, "hydrateLanInterfaceLagMemberState.parseId", map[string]interface{}{
		"input.ID.ValueString()": utils.InterfaceToJSONString(input.ID.ValueString()),
		"siteID":                 utils.InterfaceToJSONString(siteID),
		"interfaceID":            utils.InterfaceToJSONString(interfaceID),
	})

	// Get the site's accountSnapshot to find the LAG master
	siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	tflog.Debug(ctx, "Read.AccountSnapshot.response for LAG member", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
	})
	if err != nil {
		return input, fmt.Errorf("Catov2 API error getting account snapshot for LAG member: %s", err.Error())
	}

	// Look for LAG master interface in the site's interfaces
	var lagMemberFound bool
	for _, site := range siteAccountSnapshotApiData.AccountSnapshot.Sites {
		tflog.Debug(ctx, "hydrateLanInterfaceLagMemberState.siteId check", map[string]interface{}{
			"siteID":                     siteID,
			"input.SiteId.ValueString()": input.SiteId.ValueString(),
		})
		if *site.ID == siteID {
			for _, iface := range site.InfoSiteSnapshot.Interfaces {
				curInterfaceID := iface.ID
				if idxInt, err := cast.ToIntE(curInterfaceID); err == nil {
					curInterfaceID = fmt.Sprintf("INT_%d", idxInt)
				}
				tflog.Debug(ctx, "Lookinug for LAG master interface for LAG member", map[string]interface{}{
					"curInterfaceID":   curInterfaceID,
					"curInterfaceName": iface.Name,
					"interfaceID":      interfaceID,
					"*iface.DestType":  *iface.DestType,
				})
				if *iface.DestType == "LAN_LAG_MEMBER" && interfaceID == curInterfaceID {
					tflog.Debug(ctx, "Found LAG master interface for LAG member", map[string]interface{}{
						"curInterfaceID":   curInterfaceID,
						"curInterfaceName": iface.Name,
					})
					input.Name = types.StringPointerValue(iface.Name)
					input.SiteId = types.StringPointerValue(&siteID)
					input.InterfaceID = types.StringPointerValue(&interfaceID)
					input.DestType = types.StringValue("LAN_LAG_MEMBER")
					lagMemberFound = true
					break
				}
			}
			break
		}
	}

	if !lagMemberFound {
		tflog.Warn(ctx, "LAG member not found, interface may have been removed", map[string]interface{}{
			"siteID":      siteID,
			"interfaceId": interfaceID,
		})
		return input, fmt.Errorf("LAG member not found")
	}

	return input, nil
}
