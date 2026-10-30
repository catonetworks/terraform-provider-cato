package swprule

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
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

func TestProxyValidation(t *testing.T) {
	ctx := context.Background()
	p := testModel(t)
	for _, port := range []string{"0", "65536", "01", "unknown"} {
		a := p.Rule.Attributes()
		a["port"] = types.StringValue(port)
		var d diag.Diagnostics
		buildCreateObject(ctx, types.ObjectValueMust(ruleAttrTypes, a), &d)
		require.True(t, d.HasError())
	}
	a := p.Rule.Attributes()
	a["should_associate_with_all_sites"] = types.BoolValue(false)
	a["associated_site"] = types.SetValueMust(types.ObjectType{AttrTypes: idname.RefAttrTypes}, []attr.Value{})
	var d diag.Diagnostics
	buildCreateObject(ctx, types.ObjectValueMust(ruleAttrTypes, a), &d)
	require.True(t, d.HasError())
}
func TestProxyErrorsRedactSecrets(t *testing.T) {
	var d diag.Diagnostics
	checkProxyErrors(errors.New("secret-keytab-sentinel"), []*cato.PolicySiteWebProxyAddRule_Policy_SiteWebProxy_AddRule_Errors{}, "failed", &d)
	require.True(t, d.HasError())
	require.NotContains(t, fmt.Sprint(d), "secret-keytab-sentinel")
}
