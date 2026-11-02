//go:build acctest

package rpfrule

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccRemotePortFwdRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	acc.CheckCMAVars(t)()
	fixtureIP, fixturePort := selectExternalIP(t)
	_, anchorPort := selectExternalIP(t, fixturePort)
	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "FIRST_IN_POLICY"
			}
			return "LAST_IN_POLICY"
		}(), "${external_ip}", fixtureIP, "${external_port}", fixturePort).Replace(`resource "cato_remote_port_forward_rule" "test" {
at = { position = "${position}" }
rule = {
 description = "${description}"
 enabled = false
 external_ip = { id = "${external_ip}" }
 external_port_range = { from = "${external_port}", to = "${external_port}" }
 forward_icmp = false
 internal_ip = "192.0.2.241"
 internal_port_range = { from = "443", to = "443" }
 name = "acctest-remote_port_forward_rule-${name}"
 remote_ips = { global_ip_range = [], ip = ["192.0.2.100"], ip_range = [], subnet = [] }
 restriction_type = "ALLOW_LIST"
 tracking = { enabled = false, frequency = "HOURLY", mailing_list = [], subscription_group = [], webhook = [] }
}
}
`)
	}
	name := acc.GetRandName("remote_port_forward_rule")
	singleRuleConfig := config
	anchorName := name + "-anchor"
	config = func(name, description string) string {
		anchor := strings.Replace(singleRuleConfig(anchorName, "initial"), `"test" {`, `"anchor" {`, 1)
		anchor = strings.ReplaceAll(anchor, fmt.Sprintf("%q", fixturePort), fmt.Sprintf("%q", anchorPort))
		target := strings.Replace(singleRuleConfig(name, description), `"test" {`, `"test" {
 depends_on = [cato_remote_port_forward_rule.anchor]`, 1)
		anchor = strings.TrimPrefix(anchor, acc.ProviderCfg())
		return anchor + "\n" + target
	}
	resourceName := "cato_remote_port_forward_rule.test"

	published := func(state *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		revisionType := cato_models.PolicyRevisionTypePublic
		policy, err := acc.GetClient(t).RemotePortFwdPolicy(ctx, acc.CatoAccountID,
			&cato_models.RemotePortFwdPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
		if err != nil {
			return err
		}

		ids := []string{}
		for _, item := range policy.GetPolicy().GetRemotePortFwd().GetPolicy().GetRules() {
			ids = append(ids, item.GetRule().GetID())
		}
		return acc.CheckPolicyPlacement(state, resourceName, "rule.id", ids)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).RemotePortFwdPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.RemotePortFwdPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetRemotePortFwd().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_remote_port_forward_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "rule.id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-remote_port_forward_rule-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
		}, resourceName,

			// Reuse an account-owned allocated IP without changing its allocation. The new
			// rule stays disabled, uses a free external port, and is the only object deleted.
			"rule.id", published),
	})
}

func selectExternalIP(t *testing.T, excluded ...string) (allocationID, externalPort string) {
	t.Helper()
	client := acc.GetClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	result, err := client.PopLocationsAllocatedIPList(ctx, acc.CatoAccountID, nil)
	if err != nil {
		t.Fatalf("read allocated IP prerequisites: %v", err)
	}
	id := ""
	for _, item := range result.GetPopLocations().GetAllocatedIPList().GetAllocatedIP() {
		if item.GetID() != "" && item.GetAccount().GetID() == acc.CatoAccountID {
			id = item.GetID()
			break
		}
	}
	if id == "" {
		t.Skip("an account-owned allocated public IP is required for Remote Port Forwarding")
	}
	revisionType := cato_models.PolicyRevisionTypePublic
	policy, err := client.RemotePortFwdPolicy(ctx, acc.CatoAccountID, &cato_models.RemotePortFwdPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
	if err != nil {
		t.Fatalf("read published policy prerequisites: %v", err)
	}
	return id, selectFreePort(t, policy.GetPolicy().GetRemotePortFwd().GetPolicy().GetRules(), excluded...)
}

func selectFreePort(t *testing.T, rules []*cato.RemotePortFwdPolicy_Policy_RemotePortFwd_Policy_Rules, excluded ...string) string {
	t.Helper()
	for port := 54321; port <= 65535; port++ {
		used := slices.Contains(excluded, strconv.Itoa(port))
		for _, entry := range rules {
			interval := entry.GetRule().GetExternalPortRange()
			if interval.GetFrom() == nil || interval.GetTo() == nil {
				t.Fatal("incomplete external port range in published policy")
			}
			first, e := strconv.Atoi(string(*interval.GetFrom()))
			if e != nil {
				t.Fatal(e)
			}
			last, e := strconv.Atoi(string(*interval.GetTo()))
			if e != nil {
				t.Fatal(e)
			}
			if first <= port && port <= last {
				used = true
				break
			}
		}
		if !used {
			return strconv.Itoa(port)
		}
	}
	t.Fatal("no unused external test port available")
	return ""
}
