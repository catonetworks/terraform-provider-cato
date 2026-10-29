//go:build acctest

package cloudinterconnectsite

import (
	"context"
	"fmt"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccCloudInterconnectSite(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("cloud_interconnect_site")
	config := func(updated, clearDescription bool) string {
		suffix, location := "", ""
		if updated {
			suffix = "_updated"
			location = `city = "Manchester"`
		}
		description := fmt.Sprintf("description = %q", name+suffix)
		if clearDescription {
			description = ""
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_cloud_interconnect_site" "test" {
 name = %q
 site_type = "BRANCH"
 %s
 site_location = {
  country_code = "GB"
  timezone = "Europe/London"
  %s
 }
}`, name+suffix, description, location)
	}
	object := "cato_cloud_interconnect_site.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_cloud_interconnect_site" || item.Primary == nil {
					continue
				}
				limit, from := int64(1), int64(0)
				result, err := acc.GetClient(t).EntityLookup(context.Background(), acc.CatoAccountID, m.EntityTypeSite,
					&limit, &from, nil, nil, []string{item.Primary.ID}, nil, nil, nil)
				if err != nil {
					return err
				}
				page := result.GetEntityLookup()
				if page == nil || page.GetItems() == nil {
					return fmt.Errorf("invalid account-scoped configuration lookup")
				}
				if len(page.GetItems()) != 0 {
					return fmt.Errorf("owned Cloud Interconnect site still exists in configuration")
				}
			}
			return nil
		}, Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "site_location.country_code", "GB"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "name", name+"_updated"), resource.TestCheckResourceAttr(object, "site_location.city", "Manchester"))},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "description", "")},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
