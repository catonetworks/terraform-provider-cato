//go:build acctest

package mailinglist

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

func TestAccMailingList(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	name := acc.GetRandName("mailing_list")
	limit := int64(1)
	admins, err := acc.GetClient(t).Admins(context.Background(), acc.CatoAccountID, &limit, nil, nil, nil, nil)
	if err != nil {
		t.Fatal("Cannot read an administrator fixture: ", err)
	}
	items := admins.GetAdmins().GetItems()
	if len(items) == 0 || items[0].GetID() == "" {
		t.Fatal("An existing administrator is required for membership-clearing coverage")
	}
	adminID := items[0].GetID()
	config := func(updated, clearMembers bool) string {
		suffix, address := "", "terraform@example.invalid"
		if updated {
			suffix = "_updated"
			address = "updated@example.invalid"
		}
		members := fmt.Sprintf("address = [%q]", address)
		if clearMembers {
			members = fmt.Sprintf("admin = [{id = %q}]", adminID)
		}
		return acc.ProviderCfg() + fmt.Sprintf("resource \"cato_mailing_list\" \"test\" {\nname = %q\n%s\n}", name+suffix, members)
	}
	object := "cato_mailing_list.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_mailing_list" || item.Primary == nil {
				continue
			}
			id := item.Primary.ID
			result, err := acc.GetClient(t).NotificationMailingList(context.Background(), acc.CatoAccountID, m.MailingListRefInput{By: m.ObjectRefByID, Input: id})
			if err != nil {
				if missingFixture(err, id) {
					continue
				}
				return err
			}
			if result == nil || result.GetNotification() == nil {
				return fmt.Errorf("invalid mailing-list response")
			}
			if result.GetNotification().GetMailingList() != nil {
				return fmt.Errorf("owned mailing list still exists")
			}
		}
		return nil
	}, Steps: acc.WithStableResourceID([]resource.TestStep{
		{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "address.#", "1"))},
		{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
		{Config: config(true, false), Check: resource.TestCheckResourceAttr(object, "name", name+"_updated")},
		{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
		{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "address.#", "0")},
		{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
		{Config: config(true, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "admin.#", "0"), resource.TestCheckResourceAttr(object, "address.#", "1"))},
		{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: object, ImportState: true, ImportStateVerify: true},
	}, object, "id")})
}

func missingFixture(err error, id string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	expected := fmt.Sprintf("GQL-User-Error: Account ID [%s] subscription [%s] not found", acc.CatoAccountID, id)
	for _, e := range *response.GqlErrors {
		if e == nil || e.Message != expected {
			return false
		}
	}
	return true
}
