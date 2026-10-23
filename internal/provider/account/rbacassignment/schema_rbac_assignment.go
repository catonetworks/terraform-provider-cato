package rbacassignment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *assignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replacement := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages one RBAC assignment through the undocumented public assignment APIs. " +
			"Updates replace all roles and entity scope. Legacy ADMIN assignments cannot be managed. " +
			"Import using ACCOUNT_ID:ASSIGNMENT_ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier in ACCOUNT_ID:ASSIGNMENT_ID format.",
				Computed:      true,
				PlanModifiers: stable,
			},
			"assignment_id": schema.StringAttribute{
				Description:   "API assignment identifier.",
				Computed:      true,
				PlanModifiers: stable,
			},
			"account_id": schema.StringAttribute{
				Description: "Owning account, defaulting to the provider account. Changes require replacement.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description:   "User principal. Specify exactly one user_id or group_id. Changes require replacement.",
				Optional:      true,
				PlanModifiers: replacement,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRoot("group_id")),
				},
			},
			"group_id": schema.StringAttribute{
				Description:   "Users-group principal. Changes require replacement.",
				Optional:      true,
				PlanModifiers: replacement,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"role_ids": schema.SetAttribute{
				Description: "Complete nonempty set of assigned role IDs.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"allowed_entities": schema.SetNestedAttribute{
				Description:  "Complete entity scope. Use [] for no grants. SYSTEMGROUP entries represent account-wide grants.",
				Required:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: entityAttributes()},
			},
			"description": schema.StringAttribute{
				Description: "Assignment description. Removing it clears the description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"allow_viewing_unobfuscated_data": schema.BoolAttribute{
				Description: "Optional stored preference. Omitting it preserves the preference on update. " +
					"When obfuscation is DISABLED, refresh preserves configuration; import cannot recover the hidden preference.",
				Optional: true,
			},
			"allow_all_sites": schema.StringAttribute{
				Description: "Effective all-sites action derived from SYSTEMGROUP grants.",
				Computed:    true,
			},
			"allow_all_users": schema.StringAttribute{
				Description: "Effective all-users action derived from SYSTEMGROUP grants.",
				Computed:    true,
			},
			"assignment_type": schema.StringAttribute{
				Description: "Managed principal type: ADMIN_USER or GROUP.",
				Computed:    true,
			},
			"data_obfuscation_status": schema.StringAttribute{
				Description: "Effective data obfuscation status: DISABLED, OBFUSCATED or UNOBFUSCATED.",
				Computed:    true,
			},
		},
	}
}

func entityAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Entity identifier.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": schema.StringAttribute{
			Description: "Public entity type.",
			Required:    true,
			Validators: []validator.String{stringvalidator.OneOf(
				"SITE", "USERSGROUP", "GROUP", "SYSTEMGROUP", "WAN_FIREWALL_POLICY", "INTERNET_FIREWALL_POLICY",
			)},
		},
		"action": schema.StringAttribute{
			Description: "VIEW or EDIT access.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf("VIEW", "EDIT")},
		},
	}
}
