package parule

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

// PrivAccPolicyActionValidator validates that the provided string is a valid Private Access Policy Action
type PrivAccPolicyActionValidator struct{}

func (v PrivAccPolicyActionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	action := cato_models.PrivateAccessPolicyActionEnum(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid action (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllPrivateAccessPolicyActionEnum))
		return
	}
}
func (v PrivAccPolicyActionValidator) Description(_ context.Context) string {
	return fmt.Sprintf("PrivatAccessPolicy action must be one of: %v", cato_models.AllPrivateAccessPolicyActionEnum)
}
func (v PrivAccPolicyActionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// PrivAccPolicyConnOriginValidator validates that the provided set of strings are valid Private Access Policy Connection Origins
type PrivAccPolicyConnOriginValidator struct{}

func (v PrivAccPolicyConnOriginValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	var diags diag.Diagnostics
	if req.ConfigValue.IsUnknown() {
		return
	}

	origins := convert.PrepareStrings[cato_models.ConnectionOriginsEnum](ctx, req.ConfigValue, &diags)
	if diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	for _, origin := range origins {
		if !origin.IsValid() {
			resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid connection origin (%s: %s)\n - valid options: %+v",
				req.Path.String(), origin, cato_models.AllConnectionOriginsEnum))
			return
		}
	}
}

func (v PrivAccPolicyConnOriginValidator) Description(_ context.Context) string {
	return fmt.Sprintf("PrivatAccessPolicy connection_origins must be one of: %v", cato_models.AllConnectionOriginsEnum)
}
func (v PrivAccPolicyConnOriginValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

type PrivAccPolicySourceValidator struct{}

// ValidateObject for the "source" ensures that there is either a user or a group specified.
func (v PrivAccPolicySourceValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	addError := func(msg ...string) {
		if len(msg) == 0 {
			msg = []string{"at least one user or group must be specified"}
		}
		resp.Diagnostics.AddError("Field validation error", "invalid private_acces_policy source: "+msg[0])
	}
	checkError := func(e error) bool {
		if e == nil {
			return false
		}
		addError(e.Error())
		return true
	}

	source := req.ConfigValue.Attributes()
	if source == nil {
		addError()
		return
	}
	for _, srcKind := range []string{"users", "user_groups"} {
		attrValue := source[srcKind]
		if attrValue == nil {
			addError()
			continue
		}

		var items []tftypes.Value
		tfvalue, err := attrValue.ToTerraformValue(context.Background())
		if checkError(err) {
			return
		}
		if checkError(tfvalue.As(&items)) {
			return
		}
		if len(items) > 0 {
			return // Good, users or groups are specified
		}
	}
	addError() // No users or groups specified
}

func (v PrivAccPolicySourceValidator) Description(_ context.Context) string {
	return "PrivatAccessPolicy source must specify at least one user or group"
}
func (v PrivAccPolicySourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
