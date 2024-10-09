package provider

import (
	"context"
	"strings"

	cato_go_sdk "github.com/cato-networks/cato-go-sdk"
	cato_models "github.com/cato-networks/cato-go-sdk/models"
	"github.com/cato-networks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &networkRangeResource{}
	_ resource.ResourceWithConfigure = &networkRangeResource{}
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
		Description: "The `cato-oss_network_range` resource contains the configuration parameters necessary to add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()](https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Range ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface_id": schema.StringAttribute{
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
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Network range translated native IP range (CIDR)",
				Optional:    true,
			},
			"gateway": schema.StringAttribute{
				Description: "Network range gateway (Only releveant for Routed range_type)",
				Optional:    true,
			},
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
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
		},
	}
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
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
	}

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

	networkInterface, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, &entityParent, nil, nil, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	lanInterface := cato_go_sdk.EntityLookup_EntityLookup_Items_Entity{}
	for _, item := range networkInterface.EntityLookup.GetItems() {
		splitName := strings.Split(*item.Entity.Name, " \\ ")
		if splitName[1] == "LAN 01" {
			lanInterface = item.Entity
		}
	}

	tflog.Debug(ctx, "network range create", map[string]interface{}{
		"input":          utils.InterfaceToJSONString(input),
		"lanInterfaceID": lanInterface.ID,
	})

	networkRange, err := r.client.catov2.SiteAddNetworkRange(ctx, lanInterface.ID, input, r.client.AccountId)
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
	}

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

	_, err := r.client.catov2.SiteUpdateNetworkRange(ctx, plan.Id.ValueString(), input, r.client.AccountId)
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

	// check if site exist before removing
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteId.ValueString()}, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {

		_, err = r.client.catov2.SiteRemoveNetworkRange(ctx, state.Id.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteUpdateSocketInterface error",
				err.Error(),
			)
			return
		}
	}

}
