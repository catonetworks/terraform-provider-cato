package policy

import (
	"context"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSharedPolicyValidation(t *testing.T) {
	for _, value := range []types.String{types.StringNull(), types.StringUnknown(), types.StringValue("")} {
		var d diag.Diagnostics
		require.False(t, ValidID(value, &d))
		require.True(t, d.HasError())
	}
	var d diag.Diagnostics
	require.True(t, ValidID(types.StringValue("opaque"), &d))
	status := m.PolicyMutationStatusSuccess
	require.True(t, CheckStatus(&status, &d))
	require.False(t, CheckStatus(nil, &d))
	d = nil
	ValidateRuleValues(&d, KnownValue{Name: "criteria", Value: types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()}), Required: true})
	require.True(t, d.HasError())
	d = nil
	ValidateKnownValues(&d, KnownValue{Name: "optional", Value: types.StringNull()})
	require.False(t, d.HasError())
	ValidateKnownValues(&d, KnownValue{Name: "required", Value: types.StringNull(), Required: true})
	require.True(t, d.HasError())
}
func TestSharedPositions(t *testing.T) {
	ctx := context.Background()
	var d diag.Diagnostics
	value := types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{"position": types.StringValue("BEFORE_RULE"), "ref": types.StringValue("ref-1")})
	in := BuildRulePosition(ctx, value, &d)
	ValidateRulePosition(in, &d)
	require.False(t, d.HasError())
	require.Equal(t, "ref-1", *in.Ref)
	empty := types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{"position": types.StringValue("LAST_IN_POLICY"), "ref": types.StringValue("")})
	in = BuildRulePosition(ctx, empty, &d)
	require.Nil(t, in.Ref)
	d = nil
	ValidateRulePosition(&m.PolicyRulePositionInput{}, &d)
	require.True(t, d.HasError())
	d = nil
	position := m.PolicyRulePositionEnumBeforeRule
	ValidateRulePosition(&m.PolicyRulePositionInput{Position: &position}, &d)
	require.True(t, d.HasError())
	d = nil
	section := BuildSectionPosition(ctx, types.ObjectValueMust(positionAttrTypes, map[string]attr.Value{"position": types.StringValue("AFTER_SECTION"), "ref": types.StringValue("ref-2")}), &d)
	ValidateSectionPosition(section, &d)
	require.False(t, d.HasError())
	require.Equal(t, "ref-2", *section.Ref)
}

type testRevisionState struct {
	data []byte
	fail bool
}

func (s *testRevisionState) GetKey(context.Context, string) ([]byte, diag.Diagnostics) {
	if s.fail {
		var d diag.Diagnostics
		d.AddError("private read", "failed")
		return nil, d
	}
	return s.data, nil
}
func (s *testRevisionState) SetKey(_ context.Context, key string, data []byte) diag.Diagnostics {
	if key != "pending_revision" {
		panic("unexpected key")
	}
	s.data = data
	return nil
}
func TestSharedPrivateRevisionState(t *testing.T) {
	ctx := context.Background()
	store := &testRevisionState{}
	var d diag.Diagnostics
	RememberRevision(ctx, store, "revision-\"quoted\"", &d)
	require.Equal(t, "revision-\"quoted\"", PendingRevision(ctx, store, &d))
	require.False(t, d.HasError())
	RememberRevision(ctx, store, "", &d)
	require.Empty(t, store.data)
	store.data = []byte("not-json")
	require.Empty(t, PendingRevision(ctx, store, &d))
	require.True(t, d.HasError())
	d = nil
	store.fail = true
	require.Empty(t, PendingRevision(ctx, store, &d))
	require.True(t, d.HasError())
}
