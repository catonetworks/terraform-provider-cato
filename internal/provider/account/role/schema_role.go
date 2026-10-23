package role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		Description: "Manages an account-owned custom RBAC role. Updates replace the complete permission set. " +
			"Import using ACCOUNT_ID:ROLE_ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier in ACCOUNT_ID:ROLE_ID format.",
				Computed:      true,
				PlanModifiers: stable,
			},
			"role_id": schema.StringAttribute{
				Description:   "API role identifier.",
				Computed:      true,
				PlanModifiers: stable,
			},
			"account_id": schema.StringAttribute{
				Description: "Owning account; defaults to the provider account. Changes require replacement.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Custom role name.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Description: "Role description. Removing it clears the remote description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"permissions": schema.SetNestedAttribute{
				Description:  "Complete permission set. Each resource occurs once; [] grants no permissions.",
				Required:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: permissionAttributes()},
			},
			"predefined": schema.BoolAttribute{
				Description: "Whether this role is predefined. Predefined roles cannot be managed.",
				Computed:    true,
			},
			"applicable_account_type": schema.SetAttribute{
				Description: "Destination account types on which the role can be applied.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_used_on_external_access": schema.BoolAttribute{
				Description: "Whether external access currently uses this role.",
				Computed:    true,
			},
		},
	}
}

func permissionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"resource": schema.StringAttribute{
			Description: "Resource identifier from cato_permission_catalog.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"action": schema.StringAttribute{
			Description: "Granted action: VIEW or EDIT.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf("VIEW", "EDIT")},
		},
	}
}
