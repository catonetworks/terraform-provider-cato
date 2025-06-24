package provider

import (
	"context"
	"encoding/json"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &networkRangeResource{}
	_ resource.ResourceWithConfigure   = &networkRangeResource{}
	_ resource.ResourceWithImportState = &networkRangeResource{}
)

func NewNetworkRangeResource() resource.Resource {
	return &networkRangeResource{}
}

type networkRangeResource struct {
	client *catoClientData
}

func (r *networkRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_range"
}

func (r *networkRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_network_range` resource contains the configuration parameters necessary to add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()](https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange).",
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
			},
			"interface_id": schema.StringAttribute{
				Description: "Network Interface ID",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internet_only": schema.BoolAttribute{
				Description: "Internet only network range (Only releveant for Routed range_type)",
				Computed:    true,
				Required:    false,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Network range name",
				Required:    true,
			},
			"range_type": schema.StringAttribute{
				Description: "Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Network range translated native IP range (CIDR)",
				Optional:    true,
			},
			"dhcp_settings": schema.SingleNestedAttribute{
				Description: "Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type)",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"dhcp_type": schema.StringAttribute{
						Description: "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
						Required:    true,
					},
					"ip_range": schema.StringAttribute{
						Description: "Network range dhcp range (format \"192.168.1.10-192.168.1.20\")",
						Optional:    true,
					},
					"relay_group_id": schema.StringAttribute{
						Description: "Network range dhcp relay group id",
						Optional:    true,
					},
				},
			},
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
				Optional:    true,
			},
		},
	}
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *networkRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("Id"), req, resp)
}

func (r *networkRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan NetworkRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := cato_models.AddNetworkRangeInput{
		Name:             plan.Name.ValueString(),
		RangeType:        (cato_models.SubnetType)(plan.RangeType.ValueString()),
		Subnet:           plan.Subnet.ValueString(),
		LocalIP:          plan.LocalIp.ValueStringPointer(),
		TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
		Gateway:          plan.Gateway.ValueStringPointer(),
		Vlan:             plan.Vlan.ValueInt64Pointer(),
		InternetOnly:     plan.InternetOnly.ValueBoolPointer(),
	}

	internetOnly := false // default
	if !plan.InternetOnly.IsNull() {
		internetOnly = plan.InternetOnly.ValueBool()
	}
	input.InternetOnly = &internetOnly

	// get planned DHCP settings Object value, or set default value if null (for VLAN Type)
	var DhcpSettings DhcpSettings
	if plan.RangeType == types.StringValue("VLAN") {
		if plan.DhcpSettings.IsNull() {
			DhcpSettings.DhcpType = types.StringValue("DHCP_DISABLED")
		} else {
			diags = plan.DhcpSettings.As(ctx, &DhcpSettings, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		input.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{
			DhcpType:     (cato_models.DhcpType)(DhcpSettings.DhcpType.ValueString()),
			IPRange:      DhcpSettings.IpRange.ValueStringPointer(),
			RelayGroupID: DhcpSettings.RelayGroupId.ValueStringPointer(),
		}
	}

	// retrieving native-network range ID to update native range
	entityParent := cato_models.EntityInput{
		ID:   plan.SiteId.ValueString(),
		Type: "site",
	}

	lanInterface := cato_go_sdk.EntityLookup_EntityLookup_Items_Entity{}
	if plan.InterfaceId.IsNull() || plan.InterfaceId.IsUnknown() {
		networkInterface, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, &entityParent, nil, nil, nil, nil, nil)
		tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(networkInterface),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}

		for _, item := range networkInterface.EntityLookup.GetItems() {
			splitName := strings.Split(*item.Entity.Name, " \\ ")
			if splitName[1] == "LAN 01" {
				lanInterface = item.Entity
			}
		}
	} else {
		lanInterface.ID = plan.InterfaceId.ValueString()
	}

	tflog.Debug(ctx, "network range create", map[string]interface{}{
		"input":          utils.InterfaceToJSONString(input),
		"lanInterfaceID": lanInterface.ID,
	})

	tflog.Debug(ctx, "Create.SiteAddNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	networkRange, err := r.client.catov2.SiteAddNetworkRange(ctx, lanInterface.ID, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddNetworkRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(networkRange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API SiteAddNetworkRange error",
			err.Error(),
		)
		return
	}

	plan.InterfaceId = types.StringValue(lanInterface.ID)
	plan.Id = types.StringValue(networkRange.Site.AddNetworkRange.NetworkRangeID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteId.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	for _, v := range querySiteResult.EntityLookup.Items {
		if v.Entity.ID == state.SiteId.ValueString() {
			resp.State.SetAttribute(
				ctx,
				path.Root("section").AtName("id"),
				v.Entity.ID,
			)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan NetworkRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := cato_models.UpdateNetworkRangeInput{
		Name:             plan.Name.ValueStringPointer(),
		RangeType:        (*cato_models.SubnetType)(plan.RangeType.ValueStringPointer()),
		Subnet:           plan.Subnet.ValueStringPointer(),
		LocalIP:          plan.LocalIp.ValueStringPointer(),
		TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
		Gateway:          plan.Gateway.ValueStringPointer(),
		Vlan:             plan.Vlan.ValueInt64Pointer(),
		InternetOnly:     plan.InternetOnly.ValueBoolPointer(),
	}

	internetOnly := false // default
	if !plan.InternetOnly.IsNull() {
		internetOnly = plan.InternetOnly.ValueBool()
	}
	input.InternetOnly = &internetOnly

	// get planned DHCP settings Object value, or set default value if null (for VLAN Type)
	var DhcpSettings DhcpSettings
	if plan.RangeType == types.StringValue("VLAN") {
		if plan.DhcpSettings.IsNull() {
			DhcpSettings.DhcpType = types.StringValue("DHCP_DISABLED")
		} else {
			diags = plan.DhcpSettings.As(ctx, &DhcpSettings, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		input.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{
			DhcpType:     (cato_models.DhcpType)(DhcpSettings.DhcpType.ValueString()),
			IPRange:      DhcpSettings.IpRange.ValueStringPointer(),
			RelayGroupID: DhcpSettings.RelayGroupId.ValueStringPointer(),
		}
	}

	tflog.Debug(ctx, "network range update", map[string]interface{}{
		"input":          utils.InterfaceToJSONString(input),
		"lanInterfaceID": plan.Id.ValueString(),
	})

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateNetworkRangeResponse, err := r.client.catov2.SiteUpdateNetworkRange(ctx, plan.Id.ValueString(), input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateNetworkRangeResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error",
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

func (r *networkRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if interface is already removed and fail gracefully
	//	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
	_, err := r.client.catov2.SiteRemoveNetworkRange(ctx, state.Id.ValueString(), r.client.AccountId)
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
