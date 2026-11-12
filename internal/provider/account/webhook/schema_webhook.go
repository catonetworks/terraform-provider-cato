package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages one webhook integration. `body` and `auth` are write-only: the API never " +
			"echoes webhook secrets (or the body payload template, which may embed them) back on read, so " +
			"Terraform retains the configured values across refreshes instead of showing drift. Creating " +
			"this resource does not send a test request. Import using its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource ID", Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"url": schema.StringAttribute{
				Description: "Destination URL the webhook request is sent to.",
				Required:    true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the webhook is enabled.", Optional: true, Computed: true,
				Default: booldefault.StaticBool(true),
			},
			"request_method": schema.StringAttribute{
				Description: "HTTP request method.", Optional: true, Computed: true,
				Default:    stringdefault.StaticString("POST"),
				Validators: []validator.String{stringvalidator.OneOf("POST", "PUT")},
			},
			"body": schema.StringAttribute{
				Description: "Request body payload template. Write-only: never refreshed from the API; " +
					"Terraform always retains the last configured value.",
				Required: true, Sensitive: true,
			},
			"auth": schema.SingleNestedAttribute{
				Description: "Authentication configuration. Exactly one of `basic`, `bearer`, or `custom` " +
					"is required. Write-only: secret subfields are never refreshed from the API.",
				Required: true,
				Attributes: map[string]schema.Attribute{
					"basic": schema.SingleNestedAttribute{
						Description: "HTTP Basic authentication.", Optional: true,
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{Description: "Username", Required: true},
							"password": schema.StringAttribute{Description: "Password", Required: true, Sensitive: true},
						},
					},
					"bearer": schema.SingleNestedAttribute{
						Description: "Bearer token authentication.", Optional: true,
						Attributes: map[string]schema.Attribute{
							"bearer_token": schema.StringAttribute{Description: "Bearer token", Required: true, Sensitive: true},
						},
					},
					"custom": schema.SingleNestedAttribute{
						Description: "Custom header authentication.", Optional: true,
						Attributes: map[string]schema.Attribute{
							"name":  schema.StringAttribute{Description: "Header name", Required: true},
							"value": schema.StringAttribute{Description: "Header value", Required: true, Sensitive: true},
						},
					},
				},
			},
			"header": schema.ListNestedAttribute{
				Description: "Additional HTTP headers sent with the request.", Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"name":  schema.StringAttribute{Description: "Header name", Required: true},
					"value": schema.StringAttribute{Description: "Header value", Required: true},
				}},
			},
			"response_mapping": schema.ListNestedAttribute{
				Description: "Key/value pairs mapped out of the webhook response.", Optional: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"key":   schema.StringAttribute{Description: "Mapping key", Required: true},
					"value": schema.StringAttribute{Description: "Mapping value", Required: true},
				}},
			},
			"related_webhook_ref": schema.SingleNestedAttribute{
				Description: "Reference to a related webhook (e.g. the \"create\" webhook of an " +
					"\"update\" webhook pair), by ID or name. Write-only: the API only returns the " +
					"related webhook's ID on read (never its name), so this attribute is carried forward " +
					"from the prior configuration instead of being refreshed from the API.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"by": schema.StringAttribute{
						Description: "Reference kind.", Required: true,
						Validators: []validator.String{stringvalidator.OneOf("ID", "NAME")},
					},
					"input": schema.StringAttribute{
						Description: "The referenced webhook's ID or name, matching `by`.",
						Required:    true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
					},
				},
			},
		},
	}
}
