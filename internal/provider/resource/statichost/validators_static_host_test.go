package statichost

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSiteReferenceReplacementDuringAdoption(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		state, plan types.String
		wantReplace bool
	}{
		{name: "adopt", state: types.StringNull(), plan: types.StringValue("site")},
		{name: "unknown state", state: types.StringUnknown(), plan: types.StringValue("site")},
		{name: "switch to id", state: types.StringValue("site"), plan: types.StringNull()},
		{name: "unresolved plan", state: types.StringValue("site"), plan: types.StringUnknown()},
		{name: "same site", state: types.StringValue("site"), plan: types.StringValue("site")},
		{name: "different site", state: types.StringValue("site"), plan: types.StringValue("other"), wantReplace: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{StateValue: tc.state, PlanValue: tc.plan}
			var id, name stringplanmodifier.RequiresReplaceIfFuncResponse
			siteIDRequiresReplaceIf(context.Background(), req, &id)
			siteNameRequiresReplaceIf(context.Background(), req, &name)
			require.Equal(t, tc.wantReplace, id.RequiresReplace)
			require.Equal(t, tc.wantReplace, name.RequiresReplace)
		})
	}
}
