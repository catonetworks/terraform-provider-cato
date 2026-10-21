package planmodifiers

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func reference(name, id types.String) types.Object {
	return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"name": name, "id": id})
}
func references(items ...attr.Value) types.Set {
	return types.SetValueMust(idname.RefObjectType, items)
}
func objectWithValues(values map[string]attr.Value) types.Object {
	attrTypes := make(map[string]attr.Type, len(values))
	for k, v := range values {
		attrTypes[k] = v.Type(context.Background())
	}
	return types.ObjectValueMust(attrTypes, values)
}
func TestPlanModifierDescriptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for name, m := range map[string]interface {
		Description(context.Context) string
		MarkdownDescription(context.Context) string
	}{
		"active period": ActivePeriodModifier(), "DHCP": DHCPSettingsModifier(true), "empty set": EmptySetDefault(types.StringType),
		"group members": GroupMembersSetModifier(), "ID/name": IDNamePlanModifier(), "ID/name set": IDNameSetPlanModifier(),
		"null set": NullSetModifier(types.StringType), "source/destination": SourceDestObjectModifier(), "volatile index": VolatileInt64(), "WAN exceptions": WanExceptionsSetModifier(),
	} {
		t.Run(name, func(t *testing.T) {
			require.NotEmpty(t, m.Description(ctx))
			require.Equal(t, m.Description(ctx), strings.ReplaceAll(m.MarkdownDescription(ctx), "`", ""))
		})
	}
}
