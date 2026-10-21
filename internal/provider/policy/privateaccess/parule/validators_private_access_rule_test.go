package parule

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPrivateAccessActionValidation(t *testing.T) {
	t.Parallel()
	v := PrivAccPolicyActionValidator{}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid", types.StringValue(string(cato_models.AllPrivateAccessPolicyActionEnum[0])), false},
		{"invalid", types.StringValue("INVALID"), true},
		{"null", types.StringNull(), true},
		{"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("value"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

func TestPrivateAccessOriginValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		value     types.Set
		wantError bool
	}{
		{"valid", types.SetValueMust(types.StringType, []attr.Value{types.StringValue(string(cato_models.AllConnectionOriginsEnum[0]))}), false},
		{"invalid", types.SetValueMust(types.StringType, []attr.Value{types.StringValue("INVALID")}), true},
		{"null", types.SetNull(types.StringType), false},
		{"unknown", types.SetUnknown(types.StringType), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.SetResponse{}
			PrivAccPolicyConnOriginValidator{}.ValidateSet(context.Background(), validator.SetRequest{ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}
func TestPrivateAccessSourceValidation(t *testing.T) {
	t.Parallel()
	refType := types.ListType{ElemType: types.StringType}
	attrTypes := map[string]attr.Type{"users": refType, "user_groups": refType}
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	nonempty := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("user-1")})
	source := func(users, groups types.List) types.Object {
		return types.ObjectValueMust(attrTypes, map[string]attr.Value{"users": users, "user_groups": groups})
	}
	for _, tc := range []struct {
		name      string
		value     types.Object
		wantError bool
	}{
		{"users", source(nonempty, empty), false},
		{"groups", source(empty, nonempty), false},
		{"empty", source(empty, empty), true},
		{"null", types.ObjectNull(attrTypes), true},
		{"unknown", types.ObjectUnknown(attrTypes), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.ObjectResponse{}
			PrivAccPolicySourceValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}
