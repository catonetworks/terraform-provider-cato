package swprule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:funlen
func (r *siteWebProxyRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato site web proxy policy. " +
			"Underlying GraphQL is marked @beta; behavior and fields may change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy, section, or another rule",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(
								"AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION",
								"LAST_IN_POLICY", "LAST_IN_SECTION",
							),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference rule or section ID when required by position",
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Rule definition",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Rule ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Rule name",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "Rule description",
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether the rule is enabled",
						Required:    true,
					},
					"fqdn": schema.StringAttribute{
						Description: "Proxy FQDN",
						Required:    true,
					},
					"port": schema.StringAttribute{
						Description: "Proxy port",
						Required:    true,
					},
					"allowed_destination_protocol": schema.StringAttribute{
						Description: "Allowed destination protocol (HTTPS or BOTH)",
						Optional:    true,
					},
					"should_associate_with_all_sites": schema.BoolAttribute{
						Description: "Associate the proxy with all sites",
						Optional:    true,
						Computed:    true,
					},
					"associated_site": schema.SetNestedAttribute{
						Description: "Sites the proxy is associated with",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
								"name": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					"authentication_config": schema.SingleNestedAttribute{
						Description: "Authentication configuration",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"method": schema.StringAttribute{
								Description: "Authentication method (KERBEROS or NONE)",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("KERBEROS", "NONE"),
								},
							},
							"kerberos_config": schema.SingleNestedAttribute{
								Description: "Kerberos configuration",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"is_enabled": schema.BoolAttribute{
										Description: "Whether Kerberos authentication is enabled",
										Required:    true,
									},
									"encrypted_keytab": schema.StringAttribute{
										Description: "Encrypted keytab (from mutation.webProxyConfig.encryptKeytabFile). Write-only - never returned by the API.",
										Optional:    true,
										Sensitive:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
