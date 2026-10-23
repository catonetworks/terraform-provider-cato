package alwaysonrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestSchemaMatchesModelAttributeTypes(t *testing.T) {
	t.Parallel()
	var response resource.SchemaResponse
	(&alwaysOnRuleResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	require.True(t, AlwaysOnRuleObjectType.Equal(response.Schema.Type()),
		"Always On schema and model must agree on all nested attribute types")
}
