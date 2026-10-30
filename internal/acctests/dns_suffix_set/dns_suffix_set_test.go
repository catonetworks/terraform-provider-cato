//go:build acctest

package dnssuffixset

import (
	"context"
	"fmt"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccDNSSuffixSet(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("dns_suffix_set")
	config := func(updated, clearMembers bool) string {
		suffix, values := "", `["first.example.invalid", "second.example.invalid"]`
		if updated {
			suffix = "_updated"
			values = `["second.example.invalid", "first.example.invalid"]`
		}
		if clearMembers {
			values = "[]"
		}
		return acc.ProviderCfg() + fmt.Sprintf("resource \"cato_dns_suffix_set\" \"test\" {\n name = %q\n suffix = %s\n}", name+suffix, values)
	}
	object := "cato_dns_suffix_set.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_dns_suffix_set" || item.Primary == nil {
					continue
				}
				result, err := acc.GetClient(t).NetworkConfigDNSSuffixSet(context.Background(), acc.CatoAccountID,
					m.NetworkConfigDNSSuffixSetRefInput{By: m.ObjectRefByID, Input: item.Primary.ID})
				if err != nil {
					return err
				}
				if result == nil || result.GetNetworkConfig() == nil {
					return fmt.Errorf("invalid DNS set response")
				}
				if result.GetNetworkConfig().GetDNS().GetSuffixSet() != nil {
					return fmt.Errorf("owned DNS suffix set still exists")
				}
			}
			return nil
		}, Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "suffix.0", "first.example.invalid"), resource.TestCheckResourceAttr(object, "suffix.1", "second.example.invalid"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "name", name+"_updated"), resource.TestCheckResourceAttr(object, "suffix.0", "second.example.invalid"), resource.TestCheckResourceAttr(object, "suffix.1", "first.example.invalid"))},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "suffix.#", "0")},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(false, false), Check: resource.TestCheckResourceAttr(object, "suffix.#", "2")},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
