//go:build acctest

package customapplication

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccCustomApplication(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("custom_app")
	acc.CheckCMAVars(t)()
	categoryID := categoryFixture(t)
	config := func(stage int) string {
		suffix, description, destination, criteria := "", fmt.Sprintf("description = %q", name), "", `{ protocol = "TCP", port = ["443"], port_range = [] }`
		if stage >= 1 {
			suffix = "_updated"
			destination = `destination = {
 domain = ["example.com"]
 fqdn = ["api.example.com"]
 destination_ip = { ip = ["192.0.2.1"], subnet = ["198.51.100.0/24"], ip_range = [{from="203.0.113.1",to="203.0.113.10"}] }
 }`
			criteria = fmt.Sprintf(`{protocol="UDP",port=["53"],port_range=[{from="1000",to="2000"}],%s}`, destination)
		}
		category := ""
		if categoryID != "" {
			category = fmt.Sprintf("category = [{id = %q}]", categoryID)
		}
		if stage >= 2 {
			category = ""
			description = ""
			criteria = `{protocol="TCP",port=["0"],port_range=[{from="0",to="8100"}]}`
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_custom_application" "test" {
 name = %q
 %s
 %s
 criteria = [%s]
}`, name+suffix, description, category, criteria)
	}
	object := "cato_custom_application.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_custom_application" || item.Primary == nil {
				continue
			}
			_, err := acc.GetClient(t).CustomAppDataCustomApplication(context.Background(), acc.CatoAccountID, m.CustomApplicationRefInput{By: m.ObjectRefByID, Input: item.Primary.ID})
			var api *clientv2.ErrorResponse
			if !errors.As(err, &api) || api.GqlErrors == nil || len(*api.GqlErrors) == 0 {
				return fmt.Errorf("application absence not confirmed")
			}
			for _, e := range *api.GqlErrors {
				if e == nil || e.Message != "Custom application not found by ID: "+item.Primary.ID {
					return fmt.Errorf("application absence not confirmed")
				}
			}
		}
		return nil
	}, Steps: []resource.TestStep{
		{Config: config(0), Check: resource.TestCheckResourceAttrSet(object, "id")},
		{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
		{Config: config(1), Check: resource.TestCheckResourceAttr(object, "name", name+"_updated")},
		{Config: config(1), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
		{Config: config(2), Check: resource.TestCheckResourceAttr(object, "description", "")},
		{Config: config(2), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
	}})
}

// Read only: reuse a static category reference, without changing existing applications.
func categoryFixture(t *testing.T) string {
	result, err := acc.GetClient(t).CustomAppDataCustomApplicationList(context.Background(), acc.CatoAccountID,
		m.CustomApplicationListInput{Paging: &m.PagingInput{From: 0, Limit: 20}})
	if err != nil {
		t.Log("No live category fixture available; populated-category clearing remains unit-tested")
		return ""
	}
	for _, item := range result.GetCustomAppData().GetCustomApplicationList().GetItems() {
		for _, category := range item.GetCategory() {
			if category != nil && category.GetID() != "" {
				t.Log("Using a read-only static category reference fixture")
				return category.GetID()
			}
		}
	}
	t.Log("No populated-category fixture available; populated-category clearing remains unit-tested")
	return ""
}
