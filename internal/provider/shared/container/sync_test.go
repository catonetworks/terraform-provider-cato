package container

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeMembers(t *testing.T) {
	valid := func(s string) bool { return s == "a" || s == "b" }
	for _, text := range []string{"a\nb", ""} {
		var d diag.Diagnostics
		got := DecodeMembers(base64.StdEncoding.EncodeToString([]byte(text)), valid, &d)
		require.False(t, d.HasError())
		require.NotNil(t, got)
	}
	for _, text := range []string{"a\na", "invalid", "\xff"} {
		var d diag.Diagnostics
		require.Nil(t, DecodeMembers(base64.StdEncoding.EncodeToString([]byte(text)), valid, &d))
		require.True(t, d.HasError())
	}
}
func TestURLValidation(t *testing.T) {
	require.True(t, ValidURL("https://example.com/list.csv"))
	for _, url := range []string{"http://example.com", "https://user:pass@example.com", "https:///x", "https://example.com/#fragment"} {
		require.False(t, ValidURL(url))
	}
	require.True(t, ValidInterval(1))
	require.False(t, ValidInterval(0))
	require.False(t, ValidInterval(2147483648))
	var d diag.Diagnostics
	require.Nil(t, BuildSync(types.StringNull(), types.ObjectNull(SyncAttributeTypes), &d))
	require.True(t, d.HasError())
}
func TestNormalizeComputedReferences(t *testing.T) {
	value := types.ObjectValueMust(map[string]attr.Type{"ref": types.StringType}, map[string]attr.Value{"ref": types.StringUnknown()})
	result := NormalizeObject(value)
	require.False(t, result.Attributes()["ref"].IsUnknown())
	require.Equal(t, "", result.Attributes()["ref"].(types.String).ValueString())
	var d diag.Diagnostics
	set := ReferenceSet(types.SetUnknown(ReferenceType), &d)
	require.Empty(t, set)
	require.True(t, d.HasError())
	d = nil
	ParseSync(context.Background(), nil, &d)
	require.True(t, d.HasError())
}
