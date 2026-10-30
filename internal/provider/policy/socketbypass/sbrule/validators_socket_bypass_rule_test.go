package sbrule

import (
	"context"
	"fmt"
	"testing"

	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

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

func TestInvalidCriterionElements(t *testing.T) {
	t.Parallel()
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprintf("unknown=%t", unknown), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			model := testModel(t)
			fields := model.Rule.Attributes()
			criterion := fields["exception"].(types.Set)
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
			fields["exception"] = types.SetValueMust(elementType, []attr.Value{element})
			model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
			response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
			r := &policyResource{}
			r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, model)}, &response)
			require.True(t, response.Diagnostics.HasError())
			require.Contains(t, fmt.Sprint(response.Diagnostics), "Invalid configuration")
		})
	}
}

func TestVLANBoundsAndPortShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, number := range []int64{0, 1, 4094, 4095} {
		t.Run(fmt.Sprint(number), func(t *testing.T) {
			t.Parallel()
			values := types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(number)})
			var diagnostics diag.Diagnostics
			validateValues(&diagnostics, knownValue{"rule.source.vlan", values, true})
			require.Equal(t, number < 1 || number > 4094, diagnostics.HasError())
		})
	}
	var diagnostics diag.Diagnostics
	input := buildCreateObject(ctx, testModel(t).Rule, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.Equal(t, scalars.Vlan(1), input.Source.Vlan[0])
	require.Equal(t, scalars.Port("443"), input.Service.Custom[0].Port[0])
	require.Equal(t, scalars.Vlan(1), input.Exception[0].Source.Vlan[0])
	output := parseVLANs(ctx, []scalars.Vlan{1, 4094}, &diagnostics)
	require.Equal(t, types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(4094)}), output)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
}
