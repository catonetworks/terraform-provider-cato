package webhook

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func validateValues(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known, non-empty name is required.")
	}
	if plan.URL.IsNull() || plan.URL.IsUnknown() || strings.TrimSpace(plan.URL.ValueString()) == "" {
		diags.AddError("Invalid URL", "A known, non-empty URL is required.")
	}
	if plan.Body.IsUnknown() {
		diags.AddError("Unknown body", "The body payload template must be known before applying.")
	}
	if plan.Enabled.IsUnknown() || plan.RequestMethod.IsUnknown() {
		diags.AddError("Unknown value", "enabled and request_method must be known before applying.")
	}
	if plan.Header.IsUnknown() || plan.ResponseMapping.IsUnknown() {
		diags.AddError("Unknown value", "header and response_mapping must be known before applying.")
	}
	validateAuth(plan.Auth, diags)
	validateRelatedWebhookRef(ctx, plan.RelatedWebhookRef, diags)
}

func validateAuth(auth types.Object, diags *diag.Diagnostics) {
	if auth.IsNull() || auth.IsUnknown() {
		diags.AddError("Invalid auth", "A known auth block with exactly one of basic, bearer, or custom is required.")
		return
	}
	attrs := auth.Attributes()
	set := 0
	for _, key := range []string{"basic", "bearer", "custom"} {
		value, ok := attrs[key].(types.Object)
		if !ok {
			continue
		}
		if value.IsUnknown() {
			diags.AddError("Unknown auth", "auth."+key+" must be known before applying.")
			continue
		}
		if !value.IsNull() {
			set++
		}
	}
	if set != 1 {
		diags.AddError("Invalid auth", "Exactly one of auth.basic, auth.bearer, or auth.custom is required.")
	}
}

func validateRelatedWebhookRef(ctx context.Context, ref types.Object, diags *diag.Diagnostics) {
	if ref.IsNull() {
		return
	}
	if ref.IsUnknown() {
		diags.AddError("Unknown related_webhook_ref", "related_webhook_ref must be known before applying.")
		return
	}
	var value relatedWebhookRefModel
	if diags.Append(ref.As(ctx, &value, basetypes.ObjectAsOptions{})...); diags.HasError() {
		return
	}
	if value.By.ValueString() != "ID" && value.By.ValueString() != "NAME" {
		diags.AddError("Invalid related_webhook_ref", "related_webhook_ref.by must be ID or NAME.")
	}
	if strings.TrimSpace(value.Input.ValueString()) == "" {
		diags.AddError("Invalid related_webhook_ref", "related_webhook_ref.input must be a known, non-empty value.")
	}
}

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid webhook ID", "A known, non-empty ID is required.")
		return false
	}
	return true
}
