package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Connection origins validator
type PrivAccPolicyConnOriginValidator struct{}

func (v PrivAccPolicyConnOriginValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	var diags diag.Diagnostics
	if req.ConfigValue.IsUnknown() {
		return
	}

	origins := parse.PrepareStrings[cato_models.PrivateAccessPolicyOriginEnum](ctx, req.ConfigValue, &diags, "rule.connection_origins")
	if diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	for _, origin := range origins {
		if !origin.IsValid() {
			resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid connection origin (%s: %s)\n - valid options: %+v", req.Path.String(),
				origin, cato_models.AllPrivateAccessPolicyOriginEnum))
			return
		}
	}
}

func (v PrivAccPolicyConnOriginValidator) Description(_ context.Context) string {
	return fmt.Sprintf("PrivatAccessPolicy connection_origins must be one of: %v", cato_models.AllPrivateAccessPolicyOriginEnum)
}
func (v PrivAccPolicyConnOriginValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
