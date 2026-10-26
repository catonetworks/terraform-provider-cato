package roles

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestQueryFilters(t *testing.T) {
	id := types.ObjectValueMust(stringFilterAttrTypes, map[string]attr.Value{"eq": types.StringValue("r"), "neq": types.StringNull(), "in": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("r")}), "nin": types.SetNull(types.StringType)})
	filter := types.ObjectValueMust(filterAttrTypes, map[string]attr.Value{"id": id, "name": types.ObjectNull(stringFilterAttrTypes), "predefined": types.ObjectValueMust(booleanFilterAttrTypes, map[string]attr.Value{"eq": types.BoolValue(false), "neq": types.BoolNull()})})
	input, d := queryInput(context.Background(), Model{Filter: filter, SortDirection: types.StringValue("DESC")})
	require.False(t, d.HasError())
	require.Equal(t, "r", *input.Filter.ID.Eq)
	require.Equal(t, []string{"r"}, input.Filter.ID.In)
	require.False(t, *input.Filter.Predefined.Eq)
	require.Equal(t, "DESC", string(input.Sort.Name.Direction))
	_, d = queryInput(context.Background(), Model{Filter: types.ObjectUnknown(filterAttrTypes)})
	require.True(t, d.HasError())
	input, d = queryInput(context.Background(), Model{Filter: types.ObjectNull(filterAttrTypes), SortDirection: types.StringNull()})
	require.False(t, d.HasError())
	require.Nil(t, input.Filter)
	require.Equal(t, "ASC", string(input.Sort.Name.Direction))
	_, d = queryInput(context.Background(), Model{Filter: types.ObjectNull(filterAttrTypes), SortDirection: types.StringValue("bad")})
	require.True(t, d.HasError())
}
