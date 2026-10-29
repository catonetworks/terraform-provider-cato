//go:build acctest

package fqdncontainer

import (
	"context"
	"fmt"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccFQDNContainer(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("fqdn_container")
	config := func(stage int) string {
		values := `["example.com","api.example.com"]`
		if stage == 1 {
			values = `["example.org"]`
		}
		if stage == 2 {
			values = `[]`
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_fqdn_container" "test" {
 name=%q
 description=%q
 mode="INLINE"
 values=%s
}`, name, fmt.Sprintf("description-%d", stage), values)
	}
	object := "cato_fqdn_container.test"
	steps := []resource.TestStep{}
	for i := 0; i < 3; i++ {
		steps = append(steps, resource.TestStep{Config: config(i), Check: resource.TestCheckResourceAttrSet(object, "id")}, resource.TestStep{Config: config(i), PlanOnly: true, ExpectNonEmptyPlan: false}, resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true})
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: steps, CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_fqdn_container" || item.Primary == nil {
				continue
			}
			result, err := acc.GetClient(t).ContainerList(context.Background(), acc.CatoAccountID, m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID, Input: item.Primary.ID}}, Types: []m.ContainerType{m.ContainerTypeFqdn}})
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
