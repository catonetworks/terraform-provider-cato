package bgppeer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:funlen
func (r *bgpPeerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_bgp_peer`resource contains the configuration parameters necessary to add a BGP peer to a Cato site. " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.site.AddBgpPeerPayload()](https://api.catonetworks.com/documentation/#definition-AddBgpPeerPayload).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the BGP peer.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site Id",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the BGP configuration entity",
				Required:    true,
			},
			"peer_asn": schema.Int64Attribute{
				Description: "The AS number of the peer BGP endpoint.",
				Required:    true,
			},
			"cato_asn": schema.Int64Attribute{
				Description: "The AS number of Cato's BGP endpoint.",
				Required:    true,
			},
			"peer_ip": schema.StringAttribute{
				Description: "The IP address of the BGP peer, this is the configured ip from the " +
					"Site->IPSec Tunnel (Primary or Secondary)->Private IPs->Site",
				Required: true,
			},
			"advertise_default_route": schema.BoolAttribute{
				Description: "Advertise the default route (0.0.0.0/0) if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"advertise_all_routes": schema.BoolAttribute{
				Description: "Advertise all routes if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"advertise_summary_routes": schema.BoolAttribute{
				Description: "Advertise summarized routes if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"summary_route": schema.ListNestedAttribute{
				Description: "Summarized routes to advertise.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"route": schema.StringAttribute{
							Description: "Subnet of the summarized route to be advertised.",
							Required:    true,
						},
						"community": schema.ListNestedAttribute{
							Description: "Community values to associate with the summarized route.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"from": schema.Int64Attribute{
										Description: "Start of the community range.",
										Required:    true,
									},
									"to": schema.Int64Attribute{
										Description: "End of the community range.",
										Required:    true,
									},
								},
							},
							Optional: true,
						},
					},
				},
			},
			"default_action": schema.StringAttribute{
				Description: "Default action for routes not matching filters (ACCEPT or DROP).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ACCEPT", "DROP"),
				},
			},
			"perform_nat": schema.BoolAttribute{
				Description: "Perform NAT if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"md5_auth_key": schema.StringAttribute{
				Description: "MD5 authentication key for secure sessions.",
				Optional:    true,
				Sensitive:   true,
			},
			"metric": schema.Int64Attribute{
				Description: "Route preference metric; lower values are given precedence.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(50),
			},
			"hold_time": schema.Int64Attribute{
				Description: "Time (in seconds) before declaring the peer unreachable.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"keepalive_interval": schema.Int64Attribute{
				Description: "Time (in seconds) between keepalive messages.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
			},
			"bfd_enabled": schema.BoolAttribute{
				Description: "Enable BFD for session failure detection if true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"bfd_settings": schema.SingleNestedAttribute{
				Description: "Required BFD configuration if BFD is enabled.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"transmit_interval": schema.Int64Attribute{
						Description: "Time interval (in milliseconds) between BFD packets sent by this peer.",
						Optional:    true,
					},
					"receive_interval": schema.Int64Attribute{
						Description: "Time interval (in milliseconds) in which this peer expects to receive BFD packets.",
						Optional:    true,
					},
					"multiplier": schema.Int64Attribute{
						Description: "Number of missed BFD packets before considering the session down.",
						Optional:    true,
					},
				},
			},
			"tracking": schema.SingleNestedAttribute{
				Description: "Configuration for tracking the health and status of the BGP peer.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Indicates if tracking is enabled.",
						Optional:    true,
					},
					"alert_frequency": schema.StringAttribute{
						Description: "Frequency of health alerts.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("HOURLY", "DAILY", "WEEKLY", "IMMEDIATE"),
						},
					},
					"subscription_id": schema.StringAttribute{
						Description: "Subscription ID associated with this tracking rule.",
						Optional:    true,
					},
				},
			},
		},
	}
}
