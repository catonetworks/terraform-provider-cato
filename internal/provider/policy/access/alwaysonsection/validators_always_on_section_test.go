package alwaysonsection

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestSectionPositionValidator(t *testing.T) {
	t.Parallel()
	v := sectionPositionValidator{}
	require.NotEmpty(t, v.Description(context.Background()))
	require.Equal(t, v.Description(context.Background()), v.MarkdownDescription(context.Background()))
	for _, tc := range []struct {
		name    string
		value   types.Object
		invalid bool
	}{
		{name: "last", value: sectionModel("", "", "LAST_IN_POLICY", "").At},
		{name: "before", value: sectionModel("", "", "BEFORE_SECTION", "other").At},
		{name: "missing ref", value: sectionModel("", "", "BEFORE_SECTION", "").At, invalid: true},
		{name: "null", value: types.ObjectNull(policyrule.PositionAttrTypes)},
		{name: "unknown", value: types.ObjectUnknown(policyrule.PositionAttrTypes)},
		{name: "unknown ref", value: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("BEFORE_SECTION"), "ref": types.StringUnknown(),
		})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var response validator.ObjectResponse
			v.ValidateObject(context.Background(), validator.ObjectRequest{ConfigValue: tc.value, Path: path.Root("at")}, &response)
			require.Equal(t, tc.invalid, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		})
	}
}
