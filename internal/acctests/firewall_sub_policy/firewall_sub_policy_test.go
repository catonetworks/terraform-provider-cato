//go:build acctest

package firewall_sub_policy

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccIfSubPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccIfSubPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	name := acc.GetRandName("if_sub_policy")
	subPolicy := "cato_if_sub_policy.test"
	rule := "cato_if_rule.child"
	var originalID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: ifSubPolicyConfig(name, "10.201.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(subPolicy, "id"),
					resource.TestCheckResourceAttrSet(subPolicy, "scope_rule_id"),
					resource.TestCheckResourceAttr(subPolicy, "name", name),
					resource.TestCheckResourceAttr(subPolicy, "scope.source.ip.0", "10.201.0.1"),
					resource.TestCheckResourceAttrSet(rule, "rule.id"),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
					captureResourceAttr(subPolicy, "id", &originalID),
				),
			},
			{
				ResourceName:            subPolicy,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"at"},
			},
			{
				Config: ifSubPolicyConfig(name, "10.201.0.3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(subPolicy, "scope.source.ip.0", "10.201.0.3"),
					checkResourceAttrUnchanged(subPolicy, "id", &originalID),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
				),
			},
			{
				Config: ifSubPolicyConfig(name+"-replacement", "10.201.0.3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(subPolicy, "name", name+"-replacement"),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
					checkResourceAttrChanged(subPolicy, "id", &originalID),
				),
			},
		},
	})
}

func TestAccWfSubPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccWfSubPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	name := acc.GetRandName("wf_sub_policy")
	subPolicy := "cato_wf_sub_policy.test"
	rule := "cato_wf_rule.child"
	var originalID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: wfSubPolicyConfig(name, "10.202.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(subPolicy, "id"),
					resource.TestCheckResourceAttrSet(subPolicy, "scope_rule_id"),
					resource.TestCheckResourceAttr(subPolicy, "name", name),
					resource.TestCheckResourceAttr(subPolicy, "scope.source.ip.0", "10.202.0.1"),
					resource.TestCheckResourceAttrSet(rule, "rule.id"),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
					captureResourceAttr(subPolicy, "id", &originalID),
				),
			},
			{
				ResourceName:            subPolicy,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"at"},
			},
			{
				Config: wfSubPolicyConfig(name, "10.202.0.3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(subPolicy, "scope.source.ip.0", "10.202.0.3"),
					checkResourceAttrUnchanged(subPolicy, "id", &originalID),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
				),
			},
			{
				Config: wfSubPolicyConfig(name+"-replacement", "10.202.0.3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(subPolicy, "name", name+"-replacement"),
					resource.TestCheckResourceAttrPair(rule, "sub_policy_id", subPolicy, "id"),
					checkResourceAttrChanged(subPolicy, "id", &originalID),
				),
			},
		},
	})
}

func captureResourceAttr(resourceName, key string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found", resourceName)
		}
		value := res.Primary.Attributes[key]
		if value == "" {
			return fmt.Errorf("resource %q attribute %q is empty", resourceName, key)
		}
		*target = value
		return nil
	}
}

func checkResourceAttrChanged(resourceName, key string, previous *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found", resourceName)
		}
		value := res.Primary.Attributes[key]
		if value == "" {
			return fmt.Errorf("resource %q attribute %q is empty", resourceName, key)
		}
		if value == *previous {
			return fmt.Errorf("resource %q attribute %q did not change from %q", resourceName, key, value)
		}
		return nil
	}
}

func checkResourceAttrUnchanged(resourceName, key string, previous *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found", resourceName)
		}
		value := res.Primary.Attributes[key]
		if value != *previous {
			return fmt.Errorf("resource %q attribute %q changed from %q to %q", resourceName, key, *previous, value)
		}
		return nil
	}
}

func ifSubPolicyConfig(name, scopeIP string) string {
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_if_sub_policy" "test" {
  name        = %q
  description = "IF sub-policy acceptance test"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled     = true
    source      = { ip = [%q] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}

resource "cato_if_rule" "child" {
  sub_policy_id = cato_if_sub_policy.test.id
  at            = { position = "LAST_IN_POLICY" }

  rule = {
    name        = %q
    enabled     = true
    action      = "ALLOW"
    source      = { ip = ["10.201.0.2"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}
`, name, scopeIP, name+"-child")
}

func wfSubPolicyConfig(name, scopeIP string) string {
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_wf_sub_policy" "test" {
  name        = %q
  description = "WAN sub-policy acceptance test"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled     = true
    direction   = "BOTH"
    source      = { ip = [%q] }
    destination = {}
    application = {}
    tracking    = { event = { enabled = false } }
  }
}

resource "cato_wf_rule" "child" {
  sub_policy_id = cato_wf_sub_policy.test.id
  at            = { position = "LAST_IN_POLICY" }

  rule = {
    name        = %q
    enabled     = true
    action      = "ALLOW"
    direction   = "BOTH"
    source      = { ip = ["10.202.0.2"] }
    destination = {}
    application = {}
    tracking    = { event = { enabled = false } }
  }
}
`, name, scopeIP, name+"-child")
}
