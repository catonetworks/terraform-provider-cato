package idname

import (
	"context"
	"encoding/json"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func referenceValue(id, name types.String) types.Object {
	return types.ObjectValueMust(RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
}

func TestPrepareIDName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		value     types.Object
		want      *cato_models.UserRefInput
		wantError bool
	}{
		{
			name: "ID preferred", value: referenceValue(types.StringValue("user-1"), types.StringValue("Alice")),
			want: &cato_models.UserRefInput{By: cato_models.ObjectRefByID, Input: "user-1"},
		},
		{
			name: "unknown ID", value: referenceValue(types.StringUnknown(), types.StringValue("Alice")),
			want: &cato_models.UserRefInput{By: cato_models.ObjectRefByName, Input: "Alice"},
		},
		{
			name: "null ID", value: referenceValue(types.StringNull(), types.StringValue("Alice")),
			want: &cato_models.UserRefInput{By: cato_models.ObjectRefByName, Input: "Alice"},
		},
		{
			name: "empty ID", value: referenceValue(types.StringValue(""), types.StringValue("Alice")),
			want: &cato_models.UserRefInput{By: cato_models.ObjectRefByName, Input: "Alice"},
		},
		{
			name: "unknown name", value: referenceValue(types.StringValue("user-1"), types.StringUnknown()),
			want: &cato_models.UserRefInput{By: cato_models.ObjectRefByID, Input: "user-1"},
		},
		{name: "absent optional reference", value: types.ObjectNull(RefAttrTypes)},
		{name: "unknown object", value: types.ObjectUnknown(RefAttrTypes), wantError: true},
		{name: "missing reference", value: referenceValue(types.StringNull(), types.StringNull()), wantError: true},
		{name: "empty reference", value: referenceValue(types.StringValue(""), types.StringValue("")), wantError: true},
		{name: "unknown reference", value: referenceValue(types.StringUnknown(), types.StringUnknown()), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			input := PrepareIDName[cato_models.UserRefInput](context.Background(), tc.value, &diags)
			require.Equal(t, tc.wantError, diags.HasError(), "%v", diags)
			require.Equal(t, tc.want, input)
			if tc.wantError {
				require.Contains(t, diags.Errors()[0].Detail(), "known, non-empty ID or name")
			}
		})
	}
}

func TestPrepareIDNameSet(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		value     types.Set
		want      []*cato_models.UserRefInput
		wantError bool
	}{
		{name: "null", value: types.SetNull(RefObjectType), want: []*cato_models.UserRefInput{}},
		{name: "unknown", value: types.SetUnknown(RefObjectType), want: []*cato_models.UserRefInput{}},
		{name: "empty", value: types.SetValueMust(RefObjectType, nil), want: []*cato_models.UserRefInput{}},
		{
			name: "ID and name references", value: types.SetValueMust(RefObjectType, []attr.Value{
				referenceValue(types.StringValue("user-1"), types.StringNull()),
				referenceValue(types.StringNull(), types.StringValue("Alice")),
			}),
			want: []*cato_models.UserRefInput{
				{By: cato_models.ObjectRefByID, Input: "user-1"}, {By: cato_models.ObjectRefByName, Input: "Alice"},
			},
		},
		{
			name: "invalid member aborts conversion", value: types.SetValueMust(RefObjectType, []attr.Value{
				referenceValue(types.StringValue("user-1"), types.StringNull()),
				referenceValue(types.StringUnknown(), types.StringUnknown()),
			}), want: []*cato_models.UserRefInput{}, wantError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			input := PrepareIDNameSet[cato_models.UserRefInput](context.Background(), tc.value, &diags)
			require.Equal(t, tc.wantError, diags.HasError(), "%v", diags)
			require.NotNil(t, input)
			require.Equal(t, tc.want, input)
			if len(input) == 0 {
				body, err := json.Marshal(input)
				require.NoError(t, err)
				require.Equal(t, "[]", string(body))
			}
		})
	}
}
