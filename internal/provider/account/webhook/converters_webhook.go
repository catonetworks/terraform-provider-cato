package webhook

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.CreateWebhookInput {
	validateValues(ctx, plan, diags)
	auth := buildCreateAuth(ctx, plan.Auth, diags)
	header := buildHeaderInput(ctx, plan.Header, diags)
	responseMapping := buildResponseMappingInput(ctx, plan.ResponseMapping, diags)
	relatedRef := buildRelatedWebhookRef(ctx, plan.RelatedWebhookRef, diags)
	enabled := plan.Enabled.ValueBool()
	requestMethod := m.WebhookRequestMethod(plan.RequestMethod.ValueString())
	if diags.HasError() {
		return m.CreateWebhookInput{}
	}
	return m.CreateWebhookInput{
		Name:              plan.Name.ValueString(),
		URL:               plan.URL.ValueString(),
		Body:              plan.Body.ValueString(),
		Enabled:           &enabled,
		RequestMethod:     &requestMethod,
		Header:            header,
		ResponseMapping:   responseMapping,
		Auth:              auth,
		RelatedWebhookRef: relatedRef,
	}
}

func buildUpdate(ctx context.Context, plan resourceModel, id string, diags *diag.Diagnostics) m.UpdateWebhookInput {
	in := buildCreate(ctx, plan, diags)
	ref := webhookRef(id)
	if diags.HasError() {
		return m.UpdateWebhookInput{Webhook: &ref}
	}
	return m.UpdateWebhookInput{
		Webhook:           &ref,
		Name:              &in.Name,
		URL:               &in.URL,
		Body:              &in.Body,
		Enabled:           in.Enabled,
		RequestMethod:     in.RequestMethod,
		Header:            in.Header,
		ResponseMapping:   in.ResponseMapping,
		Auth:              buildUpdateAuth(in.Auth),
		RelatedWebhookRef: in.RelatedWebhookRef,
	}
}

func buildUpdateAuth(in *m.CreateWebhookAuthInput) *m.UpdateWebhookAuthInput {
	if in == nil {
		return nil
	}
	out := &m.UpdateWebhookAuthInput{}
	if in.Basic != nil {
		out.Basic = &m.UpdateBasicAuthInput{Username: &in.Basic.Username, Password: &in.Basic.Password}
	}
	if in.Bearer != nil {
		out.Bearer = &m.UpdateBearerTokenAuthInput{BearerToken: &in.Bearer.BearerToken}
	}
	if in.Custom != nil {
		out.Custom = &m.UpdateCustomHeaderAuthInput{Name: &in.Custom.Name, Value: &in.Custom.Value}
	}
	return out
}

func buildCreateAuth(ctx context.Context, auth types.Object, diags *diag.Diagnostics) *m.CreateWebhookAuthInput {
	if auth.IsNull() || auth.IsUnknown() {
		return nil
	}
	attrs := auth.Attributes()
	out := &m.CreateWebhookAuthInput{}
	if basic, ok := attrs["basic"].(types.Object); ok && utils.HasValue(basic) {
		var value basicAuthModel
		if !utils.CheckErr(diags, basic.As(ctx, &value, basetypes.ObjectAsOptions{})) {
			out.Basic = &m.CreateBasicAuthInput{Username: value.Username.ValueString(), Password: value.Password.ValueString()}
		}
	}
	if bearer, ok := attrs["bearer"].(types.Object); ok && utils.HasValue(bearer) {
		var value bearerAuthModel
		if !utils.CheckErr(diags, bearer.As(ctx, &value, basetypes.ObjectAsOptions{})) {
			out.Bearer = &m.CreateBearerTokenAuthInput{BearerToken: value.BearerToken.ValueString()}
		}
	}
	if custom, ok := attrs["custom"].(types.Object); ok && utils.HasValue(custom) {
		var value customAuthModel
		if !utils.CheckErr(diags, custom.As(ctx, &value, basetypes.ObjectAsOptions{})) {
			out.Custom = &m.CreateCustomHeaderAuthInput{Name: value.Name.ValueString(), Value: value.Value.ValueString()}
		}
	}
	return out
}

func buildRelatedWebhookRef(ctx context.Context, ref types.Object, diags *diag.Diagnostics) *m.WebhookIntegrationRefInput {
	if !utils.HasValue(ref) {
		return nil
	}
	var value relatedWebhookRefModel
	if utils.CheckErr(diags, ref.As(ctx, &value, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &m.WebhookIntegrationRefInput{By: m.ObjectRefBy(value.By.ValueString()), Input: value.Input.ValueString()}
}

func buildHeaderInput(ctx context.Context, list types.List, diags *diag.Diagnostics) []*m.HTTPHeaderNameValueInput {
	if !utils.HasValue(list) {
		return nil
	}
	var items []headerEntry
	if utils.CheckErr(diags, list.ElementsAs(ctx, &items, false)) {
		return nil
	}
	out := make([]*m.HTTPHeaderNameValueInput, 0, len(items))
	for _, item := range items {
		out = append(out, &m.HTTPHeaderNameValueInput{Name: item.Name.ValueString(), Value: item.Value.ValueString()})
	}
	return out
}

func buildResponseMappingInput(ctx context.Context, list types.List, diags *diag.Diagnostics) []*m.WebhookResponseMapEntryInput {
	if !utils.HasValue(list) {
		return nil
	}
	var items []responseMapEntry
	if utils.CheckErr(diags, list.ElementsAs(ctx, &items, false)) {
		return nil
	}
	out := make([]*m.WebhookResponseMapEntryInput, 0, len(items))
	for _, item := range items {
		out = append(out, &m.WebhookResponseMapEntryInput{Key: item.Key.ValueString(), Value: item.Value.ValueString()})
	}
	return out
}

// identityState is used after Create/Update: the mutation responses deliberately return
// only `webhook { id name }` (see client_webhook.go), so the rest of state is taken
// verbatim from the plan -- which already holds every field's known, authoritative value
// (Required/Optional attributes always reflect config, never a server-computed value, for
// this resource). The next Read re-syncs the non-secret fields from the API as usual.
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	if plan.Header.IsUnknown() {
		plan.Header = types.ListNull(HeaderObjectType)
	}
	if plan.ResponseMapping.IsUnknown() {
		plan.ResponseMapping = types.ListNull(ResponseMapObjectType)
	}
	if plan.RelatedWebhookRef.IsUnknown() {
		plan.RelatedWebhookRef = types.ObjectNull(RelatedWebhookRefAttrTypes)
	}
	return plan
}

// parseWebhook builds the post-Read resourceModel. Body, Auth and RelatedWebhookRef are
// carried forward from `prior` (the existing state/plan) rather than the API response --
// see model_webhook.go and client_webhook.go for why those three are write-only/non-
// round-trippable. Every other field is refreshed from the API as normal.
func parseWebhook(_ context.Context, item *m.Webhook, prior resourceModel, diags *diag.Diagnostics) resourceModel {
	header := make([]attr.Value, 0, len(item.Header))
	for _, h := range item.Header {
		if h == nil {
			continue
		}
		obj, d := types.ObjectValue(HeaderAttrTypes, map[string]attr.Value{
			"name": types.StringValue(h.Name), "value": types.StringValue(h.Value),
		})
		diags.Append(d...)
		header = append(header, obj)
	}
	headerList, d := types.ListValue(HeaderObjectType, header)
	diags.Append(d...)
	if len(item.Header) == 0 {
		headerList = types.ListNull(HeaderObjectType)
	}

	responseMapping := make([]attr.Value, 0, len(item.ResponseMapping))
	for _, rm := range item.ResponseMapping {
		if rm == nil {
			continue
		}
		obj, d := types.ObjectValue(ResponseMapAttrTypes, map[string]attr.Value{
			"key": types.StringValue(rm.Key), "value": types.StringValue(rm.Value),
		})
		diags.Append(d...)
		responseMapping = append(responseMapping, obj)
	}
	responseMappingList, d := types.ListValue(ResponseMapObjectType, responseMapping)
	diags.Append(d...)
	if len(item.ResponseMapping) == 0 {
		responseMappingList = types.ListNull(ResponseMapObjectType)
	}

	return resourceModel{
		ID:                types.StringValue(item.ID),
		Name:              types.StringValue(item.Name),
		URL:               types.StringValue(item.URL),
		Enabled:           types.BoolValue(item.Enabled),
		RequestMethod:     types.StringValue(string(item.RequestMethod)),
		Header:            headerList,
		ResponseMapping:   responseMappingList,
		Body:              prior.Body,
		Auth:              prior.Auth,
		RelatedWebhookRef: prior.RelatedWebhookRef,
	}
}
