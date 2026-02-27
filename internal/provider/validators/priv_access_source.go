package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type PrivAccPolicySourceValidator struct{}

// ValidateObject for the "source" ensures that there is either a user or a group specified.
func (v PrivAccPolicySourceValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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
		if attrValue := source[srcKind]; attrValue != nil {
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
	}
	addError() // No users or groups specified
}

func (v PrivAccPolicySourceValidator) Description(ctx context.Context) string {
	return "PrivatAccessPolicy source must specify at least one user or group"
}
func (v PrivAccPolicySourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
