package socketsite

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func GetCloudRouterValidator() CloudRouterValidator {
	return CloudRouterValidator{}
}

// CloudRouterValidator validates that Cloud Router mode is configured only for GCP vSocket sites.
type CloudRouterValidator struct{}

func (v CloudRouterValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	var connectionType types.String

	if !utils.HasValue(req.ConfigValue) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, req.Config.GetAttribute(ctx, path.Root("connection_type"), &connectionType)) {
		return
	}

	v.validate(connectionType, req.ConfigValue, &resp.Diagnostics)
}

func (CloudRouterValidator) Description(_ context.Context) string {
	return "is_cloud_router can only be configured for GCP vSocket HA sites"
}

func (v CloudRouterValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (CloudRouterValidator) validate(connectionType types.String, isCloudRouter types.Bool, diags *diag.Diagnostics) {
	if !utils.HasValue(isCloudRouter) || !utils.HasValue(connectionType) {
		return
	}

	if connectionType.ValueString() != string(cato_models.SiteConnectionTypeEnumSocketGCP1500) {
		diags.AddError(
			"Unsupported Cloud Router configuration",
			"is_cloud_router is supported only for GCP vSocket HA sites.",
		)
	}
}
