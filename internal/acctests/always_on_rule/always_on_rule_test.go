//go:build acctest

package always_on_rule

import (
	"context"
	"os"
	"testing"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/access/alwaysonrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// TestAccAlwaysOnRuleCreate exercises Create directly until the remaining lifecycle methods exist.
// Cleanup uses the SDK to remove only the created draft rule, without publishing policy changes.
func TestAccAlwaysOnRuleCreate(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Set TF_ACC=1 to run against the real API")
	}
	acc.SkipByEnv(t)
	acc.CheckCMAVars(t)()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client := acc.GetClient(t)
	r := alwaysonrule.NewAlwaysOnRuleResource()
	configure := resource.ConfigureResponse{}
	r.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: &catoclient.Service{
		AccountId: acc.CatoAccountID, Catov2: catoclient.NewProviderSDKClient(client),
	}}, &configure)
	require.False(t, configure.Diagnostics.HasError(), "%v", configure.Diagnostics)
	schema := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	require.False(t, schema.Diagnostics.HasError(), "%v", schema.Diagnostics)
	name := acc.GetRandName("always_on_rule")
	plan := tfsdk.Plan{Schema: schema.Schema}
	diags := plan.Set(ctx, alwaysOnModel(t, name))
	require.False(t, diags.HasError(), "%v", diags)
	response := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var state alwaysonrule.AlwaysOnRule
	require.False(t, response.State.Get(ctx, &state).HasError())
	var rule alwaysonrule.AlwaysOnRuleData
	require.False(t, state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
	id := rule.ID.ValueString()
	require.NotEmpty(t, id)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		result, err := client.PolicyZtnaAlwaysOnRemoveRule(cleanupCtx, acc.CatoAccountID, nil,
			cato_models.ZtnaAlwaysOnRemoveRuleInput{ID: id})
		require.NoError(t, err)
		payload := result.GetPolicy().GetZtnaAlwaysOn().GetRemoveRule()
		require.Equal(t, cato_models.PolicyMutationStatusSuccess, *payload.GetStatus())
		require.Empty(t, payload.GetErrors())
	})
	require.Equal(t, name, rule.Name.ValueString())
	require.False(t, rule.Enabled.ValueBool())
	require.False(t, rule.Section.IsUnknown())
	// Confirm the generated ID exists in the draft policy returned by the read API.
	policy, err := client.PolicyZtnaAlwaysOnPolicy(ctx, acc.CatoAccountID, nil)
	require.NoError(t, err)
	for _, entry := range policy.GetPolicy().GetZtnaAlwaysOn().GetPolicy().GetRules() {
		if entry.GetRule().GetID() == id {
			require.Equal(t, name, entry.GetRule().GetName())
			return
		}
	}
	t.Fatalf("created rule %q was not returned in the draft policy", id)
}

func alwaysOnModel(t *testing.T, name string) alwaysonrule.AlwaysOnRule {
	t.Helper()
	ctx := context.Background()
	duration, diags := types.ObjectValueFrom(ctx, alwaysonrule.AlwaysOnRuleBypassDurationAttrTypes,
		alwaysonrule.AlwaysOnRuleBypassDuration{Time: types.Int64Value(30), Unit: types.StringValue("MINUTES")})
	require.False(t, diags.HasError())
	source, diags := types.ObjectValueFrom(ctx, alwaysonrule.AlwaysOnRuleSourceAttrTypes, alwaysonrule.AlwaysOnRuleSource{
		User: types.SetValueMust(idname.RefObjectType, nil), UsersGroup: types.SetValueMust(idname.RefObjectType, nil),
	})
	require.False(t, diags.HasError())
	rule, diags := types.ObjectValueFrom(ctx, alwaysonrule.AlwaysOnRuleDataAttrTypes, alwaysonrule.AlwaysOnRuleData{
		Action: types.StringValue("ENFORCE"), AllowFailOpen: types.BoolValue(false), AllowUserBypass: types.BoolValue(false),
		AntiTamperMode: types.StringValue("OFF"), BypassDuration: duration, Description: types.StringValue("Create acceptance test"),
		DevicePostureProfile: types.SetValueMust(idname.RefObjectType, nil), Enabled: types.BoolValue(false),
		ID: types.StringUnknown(), Index: types.Int64Unknown(), Name: types.StringValue(name),
		Platform: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("WINDOWS")}),
		Section:  types.ObjectUnknown(alwaysonrule.AlwaysOnRuleSectionAttrTypes), Source: source,
	})
	require.False(t, diags.HasError())
	return alwaysonrule.AlwaysOnRule{At: types.ObjectNull(policyrule.PositionAttrTypes), Rule: rule}
}
