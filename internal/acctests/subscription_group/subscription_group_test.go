//go:build acctest

package subscriptiongroup

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSubscriptionGroup(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	name := acc.GetRandName("subscription_group")
	id1 := mailingFixture(t, name+"_one")
	id2 := mailingFixture(t, name+"_two")
	config := func(updated, clearMembers bool) string {
		suffix, members := "", fmt.Sprintf("mailing_list = [{id = %q}]", id1)
		if updated {
			suffix = "_updated"
			members = fmt.Sprintf("mailing_list = [{name = %q},{id = %q}]", name+"_one", id2)
		}
		if clearMembers {
			members = ""
		}
		return acc.ProviderCfg() + fmt.Sprintf("resource \"cato_subscription_group\" \"test\" {\n name = %q\n%s\n}", name+suffix, members)
	}
	object := "cato_subscription_group.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_subscription_group" || item.Primary == nil {
					continue
				}
				result := new(cato.NotificationSubscriptionGroup)
				err := acc.GetClient(t).Client.Post(context.Background(), "notificationSubscriptionGroup", `query notificationSubscriptionGroup($accountId:ID!,$subscriptionGroupRefInput:SubscriptionGroupRefInput!){notification(accountId:$accountId){subscriptionGroup(input:$subscriptionGroupRefInput){id}}}`, result, map[string]interface{}{"accountId": acc.CatoAccountID, "subscriptionGroupRefInput": m.SubscriptionGroupRefInput{By: m.ObjectRefByID, Input: item.Primary.ID}})
				if err != nil {
					if missingFixture(err, item.Primary.ID) {
						continue
					}
					return err
				}
				if result.GetNotification().GetSubscriptionGroup() != nil {
					return fmt.Errorf("owned subscription group still exists")
				}
			}
			return nil
		}, Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "mailing_list.#", "1"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.TestCheckResourceAttr(object, "mailing_list.#", "2")},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "mailing_list.#", "0"), resource.TestCheckResourceAttr(object, "integration.#", "0"))},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(false, false), Check: resource.TestCheckResourceAttr(object, "mailing_list.#", "1")},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: object, ImportState: true, ImportStateVerify: true},
		}, object, "id")})
}
func mailingFixture(t *testing.T, name string) string {
	t.Helper()
	ctx := context.Background()
	client := acc.GetClient(t)
	result := new(cato.NotificationCreateMailingList)
	err := client.Client.Post(ctx, "notificationCreateMailingList", `mutation notificationCreateMailingList($accountId:ID!,$createMailingListInput:CreateMailingListInput!){notification(accountId:$accountId){createMailingList(input:$createMailingListInput){mailingList{id name}}}}`, result, map[string]interface{}{"accountId": acc.CatoAccountID, "createMailingListInput": m.CreateMailingListInput{Name: name, Address: []string{"terraform@example.invalid"}}})
	require.NoError(t, err)
	item := result.GetNotification().GetCreateMailingList().GetMailingList()
	require.NotEmpty(t, item.GetID())
	require.Equal(t, name, item.GetName())
	id := item.GetID()
	t.Cleanup(func() {
		deleted := new(cato.NotificationDeleteMailingList)
		err := client.Client.Post(context.Background(), "notificationDeleteMailingList", `mutation notificationDeleteMailingList($accountId:ID!,$mailingListRefInput:MailingListRefInput!){notification(accountId:$accountId){deleteMailingList(input:$mailingListRefInput){mailingList{id}}}}`, deleted, map[string]interface{}{"accountId": acc.CatoAccountID, "mailingListRefInput": m.MailingListRefInput{By: m.ObjectRefByID, Input: id}})
		require.NoError(t, err)
		require.Equal(t, id, deleted.GetNotification().GetDeleteMailingList().GetMailingList().GetID())
	})
	return id
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
