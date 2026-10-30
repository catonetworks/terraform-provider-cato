//go:build acctest

package dns_dhcp

import (
	"context"
	"fmt"
	"os"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// Remove only this suite's objects after the destroy plan has captured their IDs.
// Deleting in PreConfig would merely test refresh, bypassing the Delete methods.
func TestAccDNSDHCPDeleteRace(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 required before live fixture creation")
	}
	f := newLiveFixture(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: f.config("updated"), Check: f.check(false)},
			{Config: f.config("updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: f.config("updated"), Destroy: true, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("cato_dns_forwarding_rule.test", plancheck.ResourceActionDestroy),
				plancheck.ExpectResourceAction("cato_dhcp_option.test", plancheck.ResourceActionDestroy),
				plancheck.ExpectResourceAction("cato_dhcp_relay_group.test", plancheck.ResourceActionDestroy),
				plancheck.ExpectResourceAction("cato_dns_site_settings.test", plancheck.ResourceActionDestroy),
				&deleteAfterPlan{fixture: f},
			}}},
		}, CheckDestroy: f.checkDestroy,
	})
}

type deleteAfterPlan struct{ fixture *liveFixture }

func (d *deleteAfterPlan) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	if req.Plan == nil || len(req.Plan.ResourceChanges) == 0 {
		resp.Error = fmt.Errorf("missing destroy plan; owned fixtures retained")
		return
	}
	f := d.fixture
	for _, name := range []string{"dns_forwarding_rule", "dhcp_option", "dhcp_relay_group", "dns_site_settings"} {
		found := false
		for _, change := range req.Plan.ResourceChanges {
			if change == nil || change.Address != "cato_"+name+".test" {
				continue
			}
			if change.Change == nil || len(change.Change.Actions) != 1 || change.Change.Actions[0] != "delete" {
				resp.Error = fmt.Errorf("expected a pure destroy for %s", name)
				return
			}
			before, ok := change.Change.Before.(map[string]interface{})
			if !ok || f.ids[name] == "" || before["id"] != f.ids[name] {
				resp.Error = fmt.Errorf("planned %s identity does not match the owned fixture", name)
				return
			}
			found = true
		}
		if !found {
			resp.Error = fmt.Errorf("missing owned %s destroy", name)
			return
		}
	}
	f.deleteOwnedItems()
	// Destroy must retain account settings as observed immediately before apply.
	f.last = f.snapshot()
}

func TestDeleteAfterPlanRejectsUnowned(t *testing.T) {
	for _, tc := range []struct {
		name string
		plan *tfjson.Plan
	}{
		{name: "no plan"},
		{name: "no resources", plan: &tfjson.Plan{}},
		{name: "wrong identity", plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{{Address: "cato_dns_forwarding_rule.test", Change: &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionDelete}, Before: map[string]interface{}{"id": "unrelated"}}}}}},
		{name: "not a destroy", plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{{Address: "cato_dns_forwarding_rule.test", Change: &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionUpdate}, Before: map[string]interface{}{"id": "owned"}}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &liveFixture{ids: map[string]string{"dns_forwarding_rule": "owned"}}
			check := &deleteAfterPlan{fixture: f}
			var resp plancheck.CheckPlanResponse
			check.CheckPlan(context.Background(), plancheck.CheckPlanRequest{Plan: tc.plan}, &resp)
			require.Error(t, resp.Error, "invalid plans must fail before accessing the SDK")
		})
	}
}
