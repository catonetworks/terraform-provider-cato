package terminalserverrule

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

func TestInvalidCriterionElements(t *testing.T) {
	t.Parallel()
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprintf("unknown=%t", unknown), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			model := testModel(t)
			fields := model.Rule.Attributes()
			criterion := fields["exclude_traffic"].(types.Set)
			elementType := criterion.ElementType(ctx)
			var element attr.Value
			if elementType.Equal(types.StringType) {
				element = types.StringNull()
				if unknown {
					element = types.StringUnknown()
				}
			} else {
				objectType := elementType.(types.ObjectType)
				element = types.ObjectNull(objectType.AttrTypes)
				if unknown {
					element = types.ObjectUnknown(objectType.AttrTypes)
				}
			}
			fields["exclude_traffic"] = types.SetValueMust(elementType, []attr.Value{element})
			model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
			response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
			r := &policyResource{}
			r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, model)}, &response)
			require.True(t, response.Diagnostics.HasError())
			require.Contains(t, fmt.Sprint(response.Diagnostics), "Invalid configuration")
		})
	}
}

func TestPositionOmitsEmptyReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	positionValue := types.ObjectValueMust(atAttrTypes, map[string]attr.Value{
		"position": types.StringValue("LAST_IN_POLICY"), "ref": types.StringValue(""),
	})
	var diagnostics diag.Diagnostics
	position := policyshared.BuildRulePosition(ctx, positionValue, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.NotNil(t, position)
	require.Nil(t, position.Ref)
}
