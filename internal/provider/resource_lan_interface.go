package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
				// Validators: []validator.String{
				// 	stringvalidator.RegexMatches(
				// 		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`,
				// 		"must be a valid IPv4 address",
				// 	),
				// },
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet of the LAN interface in CIDR notation",
				Required:    true,
				// Validators: []validator.String{
				// 	stringvalidator.RegexMatches(
				// 		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}/(?:[0-9]|[1-2][0-9]|3[0-2])$`,
				// 		"must be a valid CIDR notation",
				// 	),
				// },
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Translated NAT subnet configuration",
				Required:    false,
				Optional:    true,
				// Validators: []validator.String{
				// 	stringvalidator.RegexMatches(
				// 		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}/(?:[0-9]|[1-2][0-9]|3[0-2])$`,
				// 		"must be a valid CIDR notation",
				// 	),
				// },
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

func (r *lanInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan LanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := hydrateLanInterfaceAPI(ctx, plan)

	tflog.Debug(ctx, "lan_interface create", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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

	_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, plan.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()), input, r.client.AccountId)
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

	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, state.SiteId.ValueString(), cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()), input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateSocketInterface error",
			err.Error(),
		)
		return
	}
}

func hydrateLanInterfaceAPI(ctx context.Context, plan LanInterface) cato_models.UpdateSocketInterfaceInput {
	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"plan.DestType.String()": utils.InterfaceToJSONString(plan.DestType.ValueString()),
		"cato_models.SocketInterfaceDestType(plan.DestType.String())": cato_models.SocketInterfaceDestType(plan.DestType.ValueString()),
	})

	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(plan.DestType.ValueString()),
		// Name:     plan.Name.ValueStringPointer(),
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
