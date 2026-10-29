//go:build acctest

package ipaddressrangecontainer

import (
	"context"
	"fmt"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccIPAddressRangeContainer(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("ip_address_range_container")
	config := func(stage int) string {
		values := `[{from="192.0.2.1",to="192.0.2.10"},{from="198.51.100.1",to="198.51.100.1"}]`
		if stage == 1 {
			values = `[{from="203.0.113.1",to="203.0.113.10"}]`
		}
		if stage == 2 {
			values = `[]`
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_ip_address_range_container" "test" {
 name=%q
 description=%q
 mode="INLINE"
 values=%s
}`, name, fmt.Sprintf("description-%d", stage), values)
	}
	object := "cato_ip_address_range_container.test"
	steps := []resource.TestStep{}
	for i := 0; i < 3; i++ {
		steps = append(steps, resource.TestStep{Config: config(i), Check: resource.TestCheckResourceAttrSet(object, "id")}, resource.TestStep{Config: config(i), PlanOnly: true, ExpectNonEmptyPlan: false}, resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true})
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: steps, CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_ip_address_range_container" || item.Primary == nil {
				continue
			}
			result, err := acc.GetClient(t).ContainerList(context.Background(), acc.CatoAccountID, m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID, Input: item.Primary.ID}}, Types: []m.ContainerType{m.ContainerTypeIPRange}})
			if err != nil {
				return err
			}
			if result.GetContainer().GetList() == nil || result.GetContainer().GetList().Containers == nil || len(result.GetContainer().GetList().Containers) != 0 {
				return fmt.Errorf("container absence not confirmed")
			}
		}
		return nil
	}})
}
