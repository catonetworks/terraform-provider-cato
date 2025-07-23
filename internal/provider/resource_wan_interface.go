package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &wanInterfaceResource{}
	_ resource.ResourceWithConfigure = &wanInterfaceResource{}
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
				Description: "WAN interface upstream bandwitdh",
				Required:    true,
			},
			"downstream_bandwidth": schema.Int64Attribute{
				Description: "WAN interface downstream bandwitdh",
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

func (r *wanInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
			Role:       (cato_models.SocketInterfaceRole)(plan.Role.ValueString()),
			Precedence: (cato_models.SocketInterfacePrecedenceEnum)(plan.Precedence.ValueString()),
		},
	}

	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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

	intId := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteId := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteId.ValueString() + ":" + intId.ValueString())

	diags = resp.State.Set(ctx, plan)
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
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid WAN interface ID format",
			"Expected format 'site_id:interface_id', got: "+state.ID.ValueString(),
		)
		return
	}
	siteId := parts[0]
	interfaceId := parts[1]

	// Get accoutnSnapshot data to check if site exists and has interfaces
	siteAccountSnapshotData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteId}, nil, &r.client.AccountId)
	tflog.Warn(ctx, "Read.AccountSnapshot/siteAccountSnapshotData.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	if len(siteAccountSnapshotData.AccountSnapshot.Sites) == 0 {
		resp.Diagnostics.AddError(
			"Site not found",
			"Site with ID "+siteId+" not found in the account snapshot.",
		)
		return
	}

	site := siteAccountSnapshotData.AccountSnapshot.Sites[0]
	if len(site.InfoSiteSnapshot.Interfaces) == 0 {
		resp.Diagnostics.AddError(
			"No WAN interfaces found",
			"Site with ID "+siteId+" has no WAN interfaces.",
		)
		return
	}

	// Find the interface with the specified interface ID
	isPresent := false
	for _, iface := range site.InfoSiteSnapshot.Interfaces {
		if "INT_"+iface.ID == interfaceId || iface.ID == interfaceId {
			isPresent = true
			if strings.HasPrefix(iface.ID, "WAN") || strings.HasPrefix(iface.ID, "LTE") || strings.HasPrefix(iface.ID, "USB") {
				state.InterfaceID = types.StringValue(iface.ID)
			} else {
				state.InterfaceID = types.StringValue("INT_" + iface.ID)
			}
			state.SiteId = types.StringValue(siteId)
			state.Name = types.StringValue(*iface.Name)
			state.UpstreamBandwidth = types.Int64Value(*iface.UpstreamBandwidth)
			state.DownstreamBandwidth = types.Int64Value(*iface.DownstreamBandwidth)
			state.Role = types.StringValue(string(*iface.WanRoleInterfaceInfo))
			// state.Precedence = types.StringValue(string(iface.Precedence))
			break
		}
	}
	if !isPresent {
		resp.Diagnostics.AddError(
			"Interface not found",
			"Interface with ID "+interfaceId+" not found in site "+siteId+".",
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
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
			Role:       (cato_models.SocketInterfaceRole)(plan.Role.ValueString()),
			Precedence: (cato_models.SocketInterfacePrecedenceEnum)(plan.Precedence.ValueString()),
		},
	}

	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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

	intId := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteId := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteId.ValueString() + ":" + intId.ValueString())

	diags = resp.State.Set(ctx, plan)
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

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteId.ValueString()}, nil, nil, nil)
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
		accountSnapshotSite, err := r.client.catov2.AccountSnapshot(ctx, []string{state.SiteId.ValueString()}, nil, &r.client.AccountId)
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

		input := cato_models.UpdateSocketInterfaceInput{}

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
					Role:       (cato_models.SocketInterfaceRole)("wan_1"),
					Precedence: (cato_models.SocketInterfacePrecedenceEnum)("ACTIVE"),
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
		_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, state.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()), input, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteUpdateSocketInterface error",
				err.Error(),
			)
			return
		}

	}

}
