package group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_group` resource contains the configuration parameters necessary to manage a group. " +
			"Groups can contain various member types including sites, hosts, network ranges, and more. Documentation for the " +
			"underlying API used in this resource can be found at [mutation.groups.createGroup()]" +
			"(https://api.catonetworks.com/documentation/#mutation-groups.createGroup).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Group name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Group description",
				Optional:    true,
			},
			"members": schema.SetNestedAttribute{
				Description: "List of group members. Each member has 'name', 'id', and 'type' fields. " +
					"You can specify either 'name' or 'id' when creating/updating; both will be populated in state. " +
					"At least one member is required.",
				Required: true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Member name (computed if not specified)",
							Optional:    true,
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Member ID (computed if not specified)",
							Optional:    true,
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Member type: FLOATING_SUBNET, GLOBAL_IP_RANGE, HOST, NETWORK_INTERFACE, SITE, " +
								"SITE_NETWORK_SUBNET", // USER, USERS_GROUP, SYSTEM_GROUP
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"FLOATING_SUBNET",
									"GLOBAL_IP_RANGE",
									"HOST",
									"NETWORK_INTERFACE",
									"SITE",
									"SITE_NETWORK_SUBNET",
									// "USER",
									// "USERS_GROUP",
									// "SYSTEM_GROUP",
								),
							},
						},
					},
				},
			},
		},
	}
}
