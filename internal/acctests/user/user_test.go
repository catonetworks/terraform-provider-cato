//go:build acctest

package user

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccUser(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("local_user")
	config := func(updated, clearOptional bool) string {
		first, department, job, phone := "Terraform", "Engineering", "Developer", "+15555550123"
		if updated {
			first = "Updated"
			department = "Testing"
			job = "QA"
			phone = "+15555550124"
		}
		optional := fmt.Sprintf("department = %q\njob_title = %q\nphone_number = %q", department, job, phone)
		if clearOptional {
			optional = ""
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_user" "test" {
 email = %q
 first_name = %q
 last_name = "Acceptance"
 enabled = false
 %s
}`, strings.ReplaceAll(name, "_", "-")+"@example.invalid", first, optional)
	}
	object := "cato_user.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_user" || item.Primary == nil {
					continue
				}
				id := item.Primary.Attributes["id"]
				result, err := acc.GetClient(t).UserUserList(context.Background(), acc.CatoAccountID, cato_models.UserListInput{
					Filter: &cato_models.UserFilterInput{ID: []*cato_models.IDFilterInput{{Eq: &id}}},
					Paging: &cato_models.PagingInput{From: 0, Limit: 100},
				})
				if err != nil {
					return err
				}
				page := result.GetUser().GetUserList()
				if page == nil || page.GetItems() == nil {
					return fmt.Errorf("user API did not return a complete list")
				}
				for _, user := range page.GetItems() {
					if user.GetID() == id {
						return fmt.Errorf("owned test user still exists")
					}
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "enabled", "false"), resource.TestCheckResourceAttr(object, "user_import_type", "MANUAL"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.TestCheckResourceAttr(object, "first_name", "Updated")},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "department", ""), resource.TestCheckResourceAttr(object, "job_title", ""), resource.TestCheckResourceAttr(object, "phone_number", ""))},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
