package admin

import (
	"context"
	"testing"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestReadEmptyManagedRoles(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		roles []*models.AdminRole
	}{
		{name: "omitted"},
		{name: "empty", roles: []*models.AdminRole{}},
		{name: "missing role", roles: []*models.AdminRole{nil, {}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var state Admin
			var diagnostics diag.Diagnostics
			convertReadState(context.Background(), &state, &models.Admin{ManagedRoles: tc.roles}, &diagnostics)
			require.False(t, diagnostics.HasError(), "%v", diagnostics)
			require.True(t, types.SetValueMust(idname.RefObjectType, nil).Equal(state.ManagedRoles), "got %s", state.ManagedRoles)
		})
	}
}
