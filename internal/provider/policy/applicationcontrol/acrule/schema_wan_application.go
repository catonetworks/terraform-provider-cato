package acrule

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// wanApplicationSchemaAttributes returns schema attributes matching WanApplicationAttrTypes (WAN-style application matcher).
func wanApplicationSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description:   "Predefined applications (logical OR within the set)",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"custom_app": schema.SetNestedAttribute{
			Description:   "Custom applications",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom application")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"app_category": schema.SetNestedAttribute{
			Description:   "Application categories",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("App category")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"custom_category": schema.SetNestedAttribute{
			Description:   "Custom categories",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom category")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"sanctioned_apps_category": schema.SetNestedAttribute{
			Description:   "Sanctioned cloud application categories",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Sanctioned apps category")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"domain": schema.ListAttribute{
			Description: "Second-level domains to match",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"fqdn": schema.ListAttribute{
			Description: "FQDNs to match",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"ip": schema.ListAttribute{
			Description: "IPv4 addresses",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"subnet": schema.ListAttribute{
			Description: "Subnets in CIDR notation",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"ip_range": schema.ListNestedAttribute{
			Description: "Inclusive IP ranges",
			Optional:    true,
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{Optional: true, Computed: true},
					"to":   schema.StringAttribute{Optional: true, Computed: true},
				},
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description:   "Global IP range objects",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global IP range")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
		},
		"application_type": schema.ListAttribute{
			Description: "Match applications by type instead of by reference (APPLICATION, CLOUD_APPLICATION, SERVICE). " +
				"Combine with application_criteria on the rule to match by attributes (e.g. all CLOUD_APPLICATIONs supporting PCI DSS).",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
