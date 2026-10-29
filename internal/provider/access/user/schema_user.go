package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a local user. Directory-managed users cannot be imported or changed. " +
			"Remote-access eligibility is not granted by this resource.",
		Attributes: schemaUser(),
	}
}
func schemaUser() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"department":   schemaOptionalText("Department", false),
		"display_name": schema.StringAttribute{Description: "Display name returned by the API", Computed: true},
		"email":        schemaRequiredText("Email", true),
		"enabled": schema.BoolAttribute{
			Description: "Whether the local user is enabled. Defaults to true.", Optional: true, Computed: true,
			Default: booldefault.StaticBool(true),
		},
		"first_name": schemaRequiredText("First name", false),
		"id": schema.StringAttribute{
			Description: "Resource ID", Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"job_title":        schemaOptionalText("Job title", false),
		"last_name":        schemaRequiredText("Last name", false),
		"phone_number":     schemaOptionalText("Phone number", true),
		"user_import_type": schema.StringAttribute{Description: "User source; only local users can be managed", Computed: true},
		"user_status":      schema.StringAttribute{Description: "Current user status", Computed: true},
	}
}
func schemaOptionalText(description string, phone bool) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description, Optional: true, Computed: true,
		PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		Validators:    []validator.String{userTextValidator{phone: phone}},
	}
}
func schemaRequiredText(description string, email bool) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description, Required: true,
		Validators: []validator.String{userTextValidator{required: true, email: email}},
	}
}
