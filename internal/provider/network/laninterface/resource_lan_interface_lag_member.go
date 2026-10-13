package laninterface

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &lanInterfaceLagMemberResource{}
	_ resource.ResourceWithConfigure = &lanInterfaceLagMemberResource{}
)

func NewLanInterfaceLagMemberResource() resource.Resource {
	return &lanInterfaceLagMemberResource{}
}

type lanInterfaceLagMemberResource struct {
	client *catoclient.Service
}

const lanLagMemberIDParts = 2
const lanLagMemberDestType = "LAN_LAG_MEMBER"
const lanInterfaceRequiredMessage = "At least one LAN interface must be defined"

func (r *lanInterfaceLagMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lan_interface_lag_member"
}

func (r *lanInterfaceLagMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *lanInterfaceLagMemberResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *lanInterfaceLagMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LagMember
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
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(
		ctx,
		plan.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLagMember.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	plan.ID = types.StringValue(
		siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID +
			":" +
			string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID),
	)

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
	var state LagMember
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
	var plan LagMember
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
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(
		ctx,
		plan.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterfaceLanLag.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	plan.ID = types.StringValue(
		siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID +
			":" +
			string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID),
	)

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
	var state LagMember
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
	siteUpdateSocketInterfaceResponse, err := r.client.Catov2.SiteUpdateSocketInterface(
		ctx,
		state.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
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
			if apiError.GraphqlErrors[0].Message == lanInterfaceRequiredMessage {
				reservedInterface = true
			}
		}
		tflog.Debug(ctx, "Checking for reservedInterface of LAN on socket.", map[string]interface{}{
			"isReservedInterface": utils.InterfaceToJSONString(reservedInterface),
			"InterfaceID":         utils.InterfaceToJSONString(cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString())),
			"SiteID":              utils.InterfaceToJSONString(state.SiteID.ValueString()),
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

func (r *lanInterfaceLagMemberResource) hydrateLanInterfaceLagMemberState(
	ctx context.Context,
	input LagMember,
) (LagMember, error) {
	// Split the ID to extract site_id and interface_id
	parts := strings.Split(input.ID.ValueString(), ":")
	if len(parts) != lanLagMemberIDParts {
		return input, fmt.Errorf(
			"invalid LAN LAG Member interface ID format: expected format 'site_id:interface_index', got: %s",
			input.ID.ValueString(),
		)
	}
	siteID := parts[0]
	interfaceID := parts[1]
	tflog.Debug(ctx, "hydrateLanInterfaceLagMemberState.parseId", map[string]interface{}{
		"input.ID.ValueString()": utils.InterfaceToJSONString(input.ID.ValueString()),
		"siteID":                 utils.InterfaceToJSONString(siteID),
		"interfaceID":            utils.InterfaceToJSONString(interfaceID),
	})

	// Get the site's accountSnapshot to find the LAG master
	siteAccountSnapshotAPIData, err := r.client.Catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	tflog.Debug(ctx, "Read.AccountSnapshot.response for LAG member", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotAPIData),
	})
	if err != nil {
		return input, fmt.Errorf("Catov2 API error getting account snapshot for LAG member: %s", err.Error())
	}

	// Look for LAG master interface in the site's interfaces
	var lagMemberFound bool
	for _, site := range siteAccountSnapshotAPIData.AccountSnapshot.Sites {
		tflog.Debug(ctx, "hydrateLanInterfaceLagMemberState.siteId check", map[string]interface{}{
			"siteID":                     siteID,
			"input.SiteID.ValueString()": input.SiteID.ValueString(),
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
				if *iface.DestType == lanLagMemberDestType && interfaceID == curInterfaceID {
					tflog.Debug(ctx, "Found LAG master interface for LAG member", map[string]interface{}{
						"curInterfaceID":   curInterfaceID,
						"curInterfaceName": iface.Name,
					})
					input.Name = types.StringPointerValue(iface.Name)
					input.SiteID = types.StringPointerValue(&siteID)
					input.InterfaceID = types.StringPointerValue(&interfaceID)
					input.DestType = types.StringValue(lanLagMemberDestType)
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
