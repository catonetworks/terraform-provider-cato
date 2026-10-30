//go:build acctest

package allocatedip

import (
	"context"
	"fmt"
	"os"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAllocatedIP(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("allocated_ip")
	pop := testPop(t)
	config := func(updated, clearDescription bool) string {
		description := name
		if updated {
			description += "_updated"
		}
		optional := fmt.Sprintf("description = %q", description)
		if clearDescription {
			optional = ""
		}
		return acc.ProviderCfg() + fmt.Sprintf("resource \"cato_allocated_ip\" \"test\" {\n type = \"SYSTEM\"\n pop_location = {id = %q}\n%s\n}", pop, optional)
	}
	object := "cato_allocated_ip.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_allocated_ip" || item.Primary == nil {
					continue
				}
				id := item.Primary.ID
				result, err := acc.GetClient(t).PopLocationsAllocatedIPList(context.Background(), acc.CatoAccountID, &m.PopLocationAllocatedIPInput{
					Filter: &m.PopLocationAllocatedIPFilterInput{ID: &m.StringFilterInput{Eq: &id}},
				})
				if err != nil {
					return err
				}
				page := result.GetPopLocations().GetAllocatedIPList()
				if page == nil || page.GetAllocatedIP() == nil {
					return fmt.Errorf("incomplete allocated-IP response")
				}
				if len(page.GetAllocatedIP()) != 0 {
					return fmt.Errorf("owned allocation was not released")
				}
			}
			return nil
		}, Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttrSet(object, "ip"), resource.TestCheckResourceAttr(object, "type", "SYSTEM"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.TestCheckResourceAttr(object, "description", name+"_updated")},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "description", "")},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
func testPop(t *testing.T) string {
	t.Helper()
	value := os.Getenv("TFACC_ALLOCATED_IP_POP_ID")
	if value == "" {
		t.Skip("Set TFACC_ALLOCATED_IP_POP_ID to a test PoP with available public IPs and account license capacity")
	}
	return value
}
