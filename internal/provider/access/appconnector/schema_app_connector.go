package appconnector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func (r *appConnectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_app_connector` resource contains the configuration parameters necessary to manage an app connector.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Description: "Optional description of the ZTNA App Connector (max 250 characters)",
				Optional:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "Name of the ZTNA App Connector group. It will be created if it does not exist.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique ID of the ZTNA App Connector",
				Computed:    true,
			},
			"location": r.schemaLocation(),
			"name": schema.StringAttribute{
				Description: "The unique name of the ZTNA App Connector",
				Required:    true,
			},
			"pooled_bandwidth_allocation": schema.SetNestedAttribute{
				Description:   "App connector pooled bandwidth license allocations",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(BandwidthAllocationObjectType)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bandwidth": schema.Int64Attribute{
							Description: "The bandwidth (in Mbps) to allocate from the pooled license.",
							Required:    true,
						},
						"allocation_id": schema.StringAttribute{
							Description: "The pooled bandwidth license allocation ID",
							Computed:    true,
						},
						"license_id": schema.StringAttribute{
							Description: "The pooled bandwidth license to allocate from",
							Required:    true,
						},
					},
				},
			},
			"preferred_pop_location": r.schemaPreferredPopLocation(),
			"private_apps": schema.SetNestedAttribute{
				Description: "List of private applications",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: idname.SchemaNameID("Private app"),
				},
			},
			"serial_number": schema.StringAttribute{
				Description: "Unique serial number of the ZTNA App Connector",
				Computed:    true,
			},
			"socket_id": schema.StringAttribute{
				Description: "Connector type (virtual, physical)",
				Computed:    true,
			},
			"socket_model": schema.StringAttribute{
				Description: "Socket model of the ZTNA App Connector",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Connector type (virtual, physical)",
				Required:    true,
				Validators:  []validator.String{validators.AppConTypeValidator{}},
			},
		},
	}
}

func (r *appConnectorResource) schemaLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "App connector location",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "Street, number",
				Optional:    true,
			},
			"city_name": schema.StringAttribute{
				Description: "City name",
				Required:    true,
			},
			"country_code": schema.StringAttribute{
				Description: "Country code",
				Required:    true,
			},
			"state_code": schema.StringAttribute{
				Description: "State code",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone",
				Required:    true,
			},
		},
	}
}

func (r *appConnectorResource) schemaPreferredPopLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Preferred PoP locations settings",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"automatic": schema.BoolAttribute{
				Description: "Automatic PoP location",
				Required:    true,
			},
			"preferred_only": schema.BoolAttribute{
				Description: "Restrict connector attachment exclusively to the configured pop locations",
				Required:    true,
			},
			"primary": schema.SingleNestedAttribute{
				Description:   "Physical location of the primary Pop",
				Optional:      true,
				Attributes:    idname.SchemaNameID("Primary location"),
				PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
			},
			"secondary": schema.SingleNestedAttribute{
				Description:   "Physical location of the secondary Pop",
				Optional:      true,
				Attributes:    idname.SchemaNameID("Secondary location"),
				PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
			},
		},
	}
}
