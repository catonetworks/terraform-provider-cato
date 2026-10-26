package admin

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:funlen
func (r *adminResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_admin` resource contains the configuration parameters necessary to manage a Cato Networks admin user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique ID for the admin user resource, consisting of ACCOUNT_ID:ADMIN_ID. Example: `123456:98765`",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_id": schema.StringAttribute{
				Description: "Admin user ID",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "Account ID to which the admin belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the account",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`),
						"Must be a valid email address",
					),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "Account first name",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Account last name",
				Required:    true,
			},
			"password_never_expires": schema.BoolAttribute{
				Description: "Whether the password never expires",
				Optional:    true,
				Computed:    true,
			},
			"mfa_enabled": schema.BoolAttribute{
				Description: "Whether MFA is enabled for this admin, always true for admin users",
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"managed_roles": schema.SetNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Role ID",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Role name",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"reseller_roles": schema.SetNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Role ID",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Role name",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"allowed_accounts": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
							Description: "List of allowed account IDs",
						},
					},
				},
			},
		},
	}
}
