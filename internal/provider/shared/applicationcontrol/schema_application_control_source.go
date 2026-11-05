//nolint:lll,funlen
package applicationcontrol

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// ApplicationControlSourceSchemaAttributes returns the inner attribute map for Application Control / App Tenant Restriction WAN-like source blocks.
func ApplicationControlSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"country": schema.SetNestedAttribute{
			Description: "Source country matching criteria",
			Optional:    true,
			Computed:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Country"),
			},
		},
		"ip": schema.ListAttribute{
			Description: "IPv4 address list",
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"host": schema.SetNestedAttribute{
			Description: "Hosts and servers defined for your account",
			Optional:    true,
			Computed:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Host"),
			},
		},
		"site": schema.SetNestedAttribute{
			Description: "Site defined for the account",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Site"),
			},
		},
		"subnet": schema.ListAttribute{
			ElementType: types.StringType,
			Description: "Subnets and network ranges defined for the LAN interfaces of a site",
			Optional:    true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"ip_range": schema.ListNestedAttribute{
			Description: "Multiple separate IP addresses or an IP range",
			Optional:    true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{Optional: true},
					"to":   schema.StringAttribute{Optional: true},
				},
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Globally defined IP range, IP and subnet objects",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Global IP range"),
			},
		},
		"network_interface": schema.SetNestedAttribute{
			Description: "Network interface defined for a site",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Network interface"),
			},
		},
		"site_network_subnet": schema.SetNestedAttribute{
			Description: "Site network subnet",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Site network subnet"),
			},
		},
		"floating_subnet": schema.SetNestedAttribute{
			Description: "Floating subnets",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Floating subnet"),
			},
		},
		"user": schema.SetNestedAttribute{
			Description: "Individual users defined for the account",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("User"),
			},
		},
		"users_group": schema.SetNestedAttribute{
			Description: "Group of users",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("User group"),
			},
		},
		"group": schema.SetNestedAttribute{
			Description: "Groups defined for your account",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("Group"),
			},
		},
		"system_group": schema.SetNestedAttribute{
			Description: "Predefined Cato groups",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				idname.SetPlanModifier(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: idname.SchemaNameID("System group"),
			},
		},
	}
}
