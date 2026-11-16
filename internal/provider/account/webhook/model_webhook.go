package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	URL     types.String `tfsdk:"url"`
	Enabled types.Bool   `tfsdk:"enabled"`
	// RequestMethod, Header and ResponseMapping are ordinary (non-secret) fields:
	// Read refreshes them from the API like any other attribute.
	RequestMethod   types.String `tfsdk:"request_method"`
	Header          types.List   `tfsdk:"header"`
	ResponseMapping types.List   `tfsdk:"response_mapping"`
	// Body and Auth are write-only: the API never echoes webhook secrets (or the
	// body payload template, which may embed them) back on read, so Read carries
	// these two forward from the prior state/plan instead of refreshing them from
	// the API. See ai_context.md for the full rationale.
	Body types.String `tfsdk:"body"`
	Auth types.Object `tfsdk:"auth"`
	// RelatedWebhookRef is also carried forward on Read: the API only returns a
	// plain `relatedWebhookId` scalar, never the referenced webhook's name, so a
	// name-only reference cannot be round-tripped from a Read response.
	RelatedWebhookRef types.Object `tfsdk:"related_webhook_ref"`
}

type headerEntry struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// HeaderAttrTypes and HeaderObjectType describe one `header` list element.
var HeaderAttrTypes = map[string]attr.Type{"name": types.StringType, "value": types.StringType}
var HeaderObjectType = types.ObjectType{AttrTypes: HeaderAttrTypes}

type responseMapEntry struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// ResponseMapAttrTypes and ResponseMapObjectType describe one `response_mapping` list element.
var ResponseMapAttrTypes = map[string]attr.Type{"key": types.StringType, "value": types.StringType}
var ResponseMapObjectType = types.ObjectType{AttrTypes: ResponseMapAttrTypes}

type basicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

var BasicAuthAttrTypes = map[string]attr.Type{"username": types.StringType, "password": types.StringType}

type bearerAuthModel struct {
	BearerToken types.String `tfsdk:"bearer_token"`
}

var BearerAuthAttrTypes = map[string]attr.Type{"bearer_token": types.StringType}

type customAuthModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var CustomAuthAttrTypes = map[string]attr.Type{"name": types.StringType, "value": types.StringType}

// AuthAttrTypes describes the `auth` nested object: exactly one of basic/bearer/custom
// must be set (enforced by validateValues, not by the type system).
var AuthAttrTypes = map[string]attr.Type{
	"basic":  types.ObjectType{AttrTypes: BasicAuthAttrTypes},
	"bearer": types.ObjectType{AttrTypes: BearerAuthAttrTypes},
	"custom": types.ObjectType{AttrTypes: CustomAuthAttrTypes},
}

type relatedWebhookRefModel struct {
	By    types.String `tfsdk:"by"`
	Input types.String `tfsdk:"input"`
}

// RelatedWebhookRefAttrTypes mirrors WebhookIntegrationRefInput (`by` + `input`) directly,
// rather than the shared idname.Ref (`id` + `name`) pattern: the API's relatedWebhookId
// read field is a plain scalar with no name, so there is nothing to resolve/round-trip,
// and this shape matches what downstream Terraform configs pass in 1:1.
var RelatedWebhookRefAttrTypes = map[string]attr.Type{"by": types.StringType, "input": types.StringType}
