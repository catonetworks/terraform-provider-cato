package networkrange

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

// requiresReplaceIfChanged forces replacement only when a Required identity
// attribute (e.g. site_id) changes between two KNOWN values. It does NOT force
// replacement when the prior-state value is null/unknown -- that happens during
// brownfield adoption/import-hydration, where the attribute is simply being
// recorded in state for the first time and the object is not actually moving.
// Those cases update in place (SiteUpdateNetworkRange is keyed on the range id)
// instead of destroy-and-recreate.
func requiresReplaceIfChanged(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		resp.RequiresReplace = false
		return
	}
	resp.RequiresReplace = !req.PlanValue.Equal(req.StateValue)
}

// Schema defines the schema for the network range resource
func (r *networkRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: networkRangeDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Range ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway": schema.StringAttribute{
				Description: "Network range gateway (Only relevant for Routed range_type)",
				Optional:    true,
				Computed:    true,
			},
			"interface_id": schema.StringAttribute{
				Description: "Network Interface ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"interface_index": schema.StringAttribute{
				Description:   "Network Interface Index",
				Required:      false,
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{validators.SocketInterfaceIndexValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"internet_only": schema.BoolAttribute{
				Description:   "Internet only network range (Only relevant for Routed range_type)",
				Computed:      true,
				Optional:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
				Computed:    true,
			},
			"mdns_reflector": schema.BoolAttribute{
				Description:   networkRangeMDNSReflectorDescription,
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Network range name",
				Required:    true,
			},
			"range_type": schema.StringAttribute{
				Description:   "Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)",
				Required:      true,
				Validators:    []validator.String{SubnetTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						requiresReplaceIfChanged,
						"Replace only when site_id changes between two known values (not on brownfield hydration of a previously-null site_id).",
						"Replace only when site_id changes between two known values (not on brownfield hydration of a previously-null site_id).",
					),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description:   "Network range translated native IP range (CIDR)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_settings": dhcp.SchemaDhcpSettings(true),
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
				Optional:    true,
			},
		},
	}
}
