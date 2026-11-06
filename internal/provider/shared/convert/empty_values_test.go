package convert

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPreserveEmptyValues(t *testing.T) {
	listType := types.ListType{ElemType: types.StringType}
	objectTypes := map[string]attr.Type{"items": listType}
	emptyList := types.ListValueMust(types.StringType, nil)
	populatedList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("remote")})
	emptyObject := types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": emptyList})
	populatedObject := types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": populatedList})
	tests := []struct {
		name                     string
		current, prior, expected attr.Value
	}{
		{"omitted list remains null", types.ListNull(types.StringType), types.ListNull(types.StringType), types.ListNull(types.StringType)},
		{"empty list", types.ListNull(types.StringType), emptyList, emptyList},
		{"empty set", types.SetNull(types.StringType), types.SetValueMust(types.StringType, nil), types.SetValueMust(types.StringType, nil)},
		{"empty nested object", types.ObjectNull(objectTypes), emptyObject, emptyObject},
		{"all null object", types.ObjectNull(objectTypes), types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": types.ListNull(types.StringType)}), types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": types.ListNull(types.StringType)})},
		{"removed list stays removed", types.ListNull(types.StringType), populatedList, types.ListNull(types.StringType)},
		{"removed object stays removed", types.ObjectNull(objectTypes), populatedObject, types.ObjectNull(objectTypes)},
		{"unknown list is not copied", types.ListNull(types.StringType), types.ListUnknown(types.StringType), types.ListNull(types.StringType)},
		{"unknown child is not copied", types.ObjectNull(objectTypes), types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": types.ListUnknown(types.StringType)}), types.ObjectNull(objectTypes)},
		{"remote list wins", populatedList, emptyList, populatedList},
		{"remote nested list wins", populatedObject, emptyObject, populatedObject},
		{"nested empty list", types.ObjectValueMust(objectTypes, map[string]attr.Value{"items": types.ListNull(types.StringType)}), emptyObject, emptyObject},
		{"removed scalar stays removed", types.StringNull(), types.StringValue("old"), types.StringNull()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := map[string]attr.Type{"value": tt.current.Type(context.Background())}
			current := types.ObjectValueMust(attrs, map[string]attr.Value{"value": tt.current})
			prior := types.ObjectValueMust(attrs, map[string]attr.Value{"value": tt.prior})
			got, diags := PreserveEmptyValues(context.Background(), attrs, current, prior)
			require.False(t, diags.HasError(), "%v", diags)
			require.True(t, tt.expected.Equal(got.Attributes()["value"]), "expected %s, got %s", tt.expected, got.Attributes()["value"])
		})
	}
}
