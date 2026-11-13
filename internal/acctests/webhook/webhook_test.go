//go:build acctest

package webhook

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// checkDestroyedDocument deliberately never selects `auth` (see
// internal/provider/account/webhook/client_webhook.go for why: the generated/CLI field
// selection for this query includes `auth { ... on CustomHeadersAuth {...} }`, which is not
// yet deployed on every tenant and fails at runtime). This mirrors the resource's own
// read document exactly.
const checkDestroyedDocument = `query notificationIntegrationWebhook($accountId: ID!, $webhookIntegrationRefInput: WebhookIntegrationRefInput!) {
notification(accountId: $accountId) {
integration {
webhook(input: $webhookIntegrationRefInput) { id }
}
}
}`

type checkDestroyedResponse struct {
	Notification *struct {
		Integration struct {
			Webhook *struct {
				ID string `json:"id"`
			} `json:"webhook"`
		} `json:"integration"`
	} `json:"notification"`
}

func TestAccWebhook(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	name := acc.GetRandName("webhook")

	config := func(updated bool) string {
		suffix, url := "", "https://example.invalid/webhook"
		if updated {
			suffix, url = "_updated", "https://example.invalid/webhook-updated"
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_webhook" "test" {
  name           = %q
  url            = %q
  enabled        = true
  request_method = "POST"
  body           = "{\"text\":\"acctest\"}"
  auth = {
    bearer = { bearer_token = "acctest-dummy-token" }
  }
}`, name+suffix, url)
	}
	object := "cato_webhook.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_webhook" || item.Primary == nil {
					continue
				}
				id := item.Primary.ID
				result := new(checkDestroyedResponse)
				err := acc.GetClient(t).Client.Post(context.Background(), "notificationIntegrationWebhook", checkDestroyedDocument, result,
					map[string]interface{}{"accountId": acc.CatoAccountID, "webhookIntegrationRefInput": map[string]interface{}{"by": "ID", "input": id}})
				if err != nil {
					if missingFixture(err, id) {
						continue
					}
					return err
				}
				if result.Notification == nil {
					return fmt.Errorf("invalid webhook response")
				}
				if result.Notification.Integration.Webhook != nil {
					return fmt.Errorf("owned webhook still exists")
				}
			}
			return nil
		},
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(false), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(object, "id"),
				resource.TestCheckResourceAttr(object, "name", name),
				resource.TestCheckResourceAttr(object, "request_method", "POST"),
			)},
			{Config: config(false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"body", "auth"}},
			{Config: config(true), Check: resource.TestCheckResourceAttr(object, "url", "https://example.invalid/webhook-updated")},
			{Config: config(true), PlanOnly: true, ExpectNonEmptyPlan: false},
		}, object, "id"),
	})
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
