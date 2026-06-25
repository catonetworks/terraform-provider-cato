package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// wanApplicationSchemaAttributes returns schema attributes matching WanApplicationAttrTypes (WAN-style application matcher).
//
//nolint:funlen
func wanApplicationSchemaAttributes() map[string]schema.Attribute {
	nameIDNested := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Object name",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				Description:   "Object ID",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description:  "Predefined applications (logical OR within the set)",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"custom_app": schema.SetNestedAttribute{
			Description:  "Custom applications",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"app_category": schema.SetNestedAttribute{
			Description:  "Application categories",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"custom_category": schema.SetNestedAttribute{
			Description:  "Custom categories",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"sanctioned_apps_category": schema.SetNestedAttribute{
			Description:  "Sanctioned cloud application categories",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
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
			Description:  "Global IP range objects",
			Optional:     true,
			Computed:     true,
			NestedObject: nameIDNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
