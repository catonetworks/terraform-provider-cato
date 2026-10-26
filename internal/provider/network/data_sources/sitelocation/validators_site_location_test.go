package sitelocation

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSiteLocationFilterValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		value     types.List
		wantError bool
	}{
		{"null", types.ListNull(types.ObjectType{AttrTypes: SLDFilterAttrTypes}), true},
		{"unknown", types.ListUnknown(types.ObjectType{AttrTypes: SLDFilterAttrTypes}), false},
		{"empty", types.ListValueMust(types.ObjectType{AttrTypes: SLDFilterAttrTypes}, []attr.Value{}), true},
		{"filter", types.ListValueMust(types.ObjectType{AttrTypes: SLDFilterAttrTypes}, []attr.Value{
			types.ObjectValueMust(SLDFilterAttrTypes, map[string]attr.Value{"field": types.StringValue("city"), "search": types.StringValue("Prague"), "operation": types.StringValue("exact")}),
		}), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.ListResponse{}
			sldFilterValidator{}.ValidateList(context.Background(), validator.ListRequest{ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError())
		})
	}
}
