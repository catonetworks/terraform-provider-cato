package splittunnelrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCreatePreservesOmittedDNSExclusion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	backend := testBackendData(t)
	backend.object["dnsExclusion"] = map[string]any{"domain": []any{}}
	r := testResource(t, backend)
	model := testModel(t)
	fields := model.Rule.Attributes()
	fields["dns_exclusion"] = types.ObjectNull(ruleDNSExclusionAttrTypes)
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, model)}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var exclusion types.Object
	require.False(t, response.State.GetAttribute(ctx, path.Root("rule").AtName("dns_exclusion"), &exclusion).HasError())
	require.True(t, exclusion.IsNull())
	input := backend.variables[1]["splitTunnelAddRuleInput"].(map[string]any)
	require.NotContains(t, input["rule"].(map[string]any), "dnsExclusion")
}

func TestImportReadsRemoteDNSExclusion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := testResource(t, testBackendData(t))
	model := testModel(t)
	fields := model.Rule.Attributes()
	fields["name"] = types.StringNull()
	fields["dns_exclusion"] = types.ObjectNull(ruleDNSExclusionAttrTypes)
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	state := testState(t, model)
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var exclusion types.Object
	require.False(t, response.State.GetAttribute(ctx, path.Root("rule").AtName("dns_exclusion"), &exclusion).HasError())
	require.False(t, exclusion.IsNull())
}

func TestReadDNSDriftRemainsVisible(t *testing.T) {
	ctx := context.Background()
	model := testModel(t)
	fields := model.Rule.Attributes()
	fields["dns_exclusion"] = types.ObjectNull(ruleDNSExclusionAttrTypes)
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	state := testState(t, model)
	r := testResource(t, testBackendData(t))
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var exclusion types.Object
	require.False(t, response.State.GetAttribute(ctx, path.Root("rule").AtName("dns_exclusion"), &exclusion).HasError())
	require.False(t, exclusion.IsNull(), "nonempty external exclusions must remain visible")
}

func TestImportNormalizesEmptyDNSExclusion(t *testing.T) {
	ctx := context.Background()
	backend := testBackendData(t)
	backend.object["dnsExclusion"] = map[string]any{"domain": []any{}}
	model := testModel(t)
	fields := model.Rule.Attributes()
	fields["name"] = types.StringNull()
	fields["dns_exclusion"] = types.ObjectNull(ruleDNSExclusionAttrTypes)
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	state := testState(t, model)
	response := resource.ReadResponse{State: state}
	r := testResource(t, backend)
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var exclusion types.Object
	require.False(t, response.State.GetAttribute(ctx, path.Root("rule").AtName("dns_exclusion"), &exclusion).HasError())
	require.True(t, exclusion.IsNull(), "empty API defaults must match omitted configuration on import")
}
