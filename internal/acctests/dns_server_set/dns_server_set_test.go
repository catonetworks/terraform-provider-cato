//go:build acctest

package dnsserverset

import (
	"context"
	"fmt"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccDNSServerSet(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("dns_server_set")
	config := func(updated, clearMembers bool) string {
		suffix, values := "", `["192.0.2.1", "192.0.2.2"]`
		if updated {
			suffix = "_updated"
			values = `["192.0.2.2", "192.0.2.1"]`
		}
		if clearMembers {
			values = "[]"
		}
		return acc.ProviderCfg() + fmt.Sprintf("resource \"cato_dns_server_set\" \"test\" {\n name = %q\n server = %s\n}", name+suffix, values)
	}
	object := "cato_dns_server_set.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_dns_server_set" || item.Primary == nil {
					continue
				}
				result, err := acc.GetClient(t).NetworkConfigDNSServerSet(context.Background(), acc.CatoAccountID,
					m.NetworkConfigDNSServerSetRefInput{By: m.ObjectRefByID, Input: item.Primary.ID})
				if err != nil {
					return err
				}
				if result == nil || result.GetNetworkConfig() == nil {
					return fmt.Errorf("invalid DNS set response")
				}
				if result.GetNetworkConfig().GetDNS().GetServerSet() != nil {
					return fmt.Errorf("owned DNS server set still exists")
				}
			}
			return nil
		}, Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "server.0", "192.0.2.1"), resource.TestCheckResourceAttr(object, "server.1", "192.0.2.2"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "name", name+"_updated"), resource.TestCheckResourceAttr(object, "server.0", "192.0.2.2"), resource.TestCheckResourceAttr(object, "server.1", "192.0.2.1"))},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "server.#", "0")},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(false, false), Check: resource.TestCheckResourceAttr(object, "server.#", "2")},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
