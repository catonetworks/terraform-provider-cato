package webhook

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped webhook CRUD contract. Every method returns the
// SDK's own *m.Webhook model type (never a type local to this package) so that
// mocks.WebhookClient -- which must reference the return type by name -- does not create an
// import cycle with this package (mirrors every other ResourceClient in this provider:
// mailing_list/subscription_group/groupmembers all return SDK-owned types for the same
// reason).
//
// Deliberately NOT the generated Catov2.NotificationIntegration{Create,Update,Delete}Webhook
// methods (added to the SDK in cato-go-sdk for completeness/parity, see ai_context.md): the
// CLI-sourced field selection those generated operations use requests the full webhook object
// back, including `auth { ... on CustomHeadersAuth { ... } }` -- a GraphQL union member that
// is not yet deployed on every tenant's notification microservice and fails with a
// DOWNSTREAM_SERVICE_ERROR / UnknownType validation error at runtime. This resource instead
// hand-writes minimal inline documents (mirroring the mailing_list/subscription_group
// convention of selecting only the fields this resource owns) that never select `auth` at
// all -- consistent with auth/body being write-only and never read back from the API anyway.
// Because `auth` (and, for create/update/delete, `body`) are never present in the JSON
// response, decoding into *m.Webhook (whose Auth field is the AuthInterface GraphQL
// interface type) is safe: encoding/json never attempts to instantiate a concrete type for
// a field whose key is simply absent from the response.
type ResourceClient interface {
	CreateWebhook(context.Context, string, m.CreateWebhookInput, ...clientv2.RequestInterceptor,
	) (*m.Webhook, error)
	UpdateWebhook(context.Context, string, m.UpdateWebhookInput, ...clientv2.RequestInterceptor,
	) (*m.Webhook, error)
	DeleteWebhook(context.Context, string, m.WebhookIntegrationRefInput, ...clientv2.RequestInterceptor,
	) (*m.Webhook, error)
	NotificationIntegrationWebhook(context.Context, string, m.WebhookIntegrationRefInput, ...clientv2.RequestInterceptor,
	) (*m.Webhook, error)
	// WebhookIDByName resolves a webhook's id from its name on the tenant.
	// Returns (id, found, error). Used to rewrite a related_webhook_ref given
	// by NAME into a by-ID ref before create/update, because the backend's
	// by-NAME resolver for relatedWebhookRef is unreliable (reports an existing
	// webhook "not found") while by-ID always resolves -- live-verified on
	// badev4. Backed by the generated Catov2.NotificationIntegrationWebhookList
	// (its SDK query selects only `auth { authType }`, so it is safe against the
	// undeployed CustomHeadersAuth union member).
	WebhookIDByName(ctx context.Context, account, name string) (string, bool, error)
}

type sdkClient struct{ client *catoclient.Service }

// The envelope types below exist only to give Post() somewhere to decode the full GraphQL
// response shape into; they are deliberately unexported and never appear in the
// ResourceClient interface (see its doc comment).
type createWebhookEnvelope struct {
	Notification struct {
		Integration struct {
			CreateWebhook struct {
				Webhook *m.Webhook `json:"webhook"`
			} `json:"createWebhook"`
		} `json:"integration"`
	} `json:"notification"`
}
type updateWebhookEnvelope struct {
	Notification struct {
		Integration struct {
			UpdateWebhook struct {
				Webhook *m.Webhook `json:"webhook"`
			} `json:"updateWebhook"`
		} `json:"integration"`
	} `json:"notification"`
}
type deleteWebhookEnvelope struct {
	Notification struct {
		Integration struct {
			DeleteWebhook struct {
				Webhook *m.Webhook `json:"webhook"`
			} `json:"deleteWebhook"`
		} `json:"integration"`
	} `json:"notification"`
}
type readWebhookEnvelope struct {
	// Notification is a pointer (unlike the create/update/delete envelopes) so a
	// malformed/incomplete API response (no "notification" key at all) can be
	// distinguished from a genuinely-missing webhook (a present but null "webhook" key).
	Notification *struct {
		Integration struct {
			Webhook *m.Webhook `json:"webhook"`
		} `json:"integration"`
	} `json:"notification"`
}

const webhookSelection = "id name enabled url requestMethod header { name value } " +
	"responseMapping { key value } relatedWebhookId"

const createDocument = `mutation notificationIntegrationCreateWebhook($accountId: ID!, $createWebhookInput: CreateWebhookInput!) {
notification(accountId: $accountId) {
integration {
createWebhook(input: $createWebhookInput) {
webhook { id name } } } } }`

const updateDocument = `mutation notificationIntegrationUpdateWebhook($accountId: ID!, $updateWebhookInput: UpdateWebhookInput!) {
notification(accountId: $accountId) {
integration {
updateWebhook(input: $updateWebhookInput) {
webhook { id name } } } } }`

const deleteDocument = `mutation notificationIntegrationDeleteWebhook($accountId: ID!, $webhookIntegrationRefInput: WebhookIntegrationRefInput!) {
notification(accountId: $accountId) {
integration {
deleteWebhook(input: $webhookIntegrationRefInput) {
webhook { id } } } } }`

const readDocument = `query notificationIntegrationWebhook($accountId: ID!, $webhookIntegrationRefInput: WebhookIntegrationRefInput!) {
notification(accountId: $accountId) {
integration {
webhook(input: $webhookIntegrationRefInput) { ` + webhookSelection + ` } } } }`

func (c *sdkClient) CreateWebhook(
	ctx context.Context, account string, input m.CreateWebhookInput,
	interceptors ...clientv2.RequestInterceptor,
) (*m.Webhook, error) {
	result := new(createWebhookEnvelope)
	err := c.client.V2().Client.Post(ctx, "notificationIntegrationCreateWebhook", createDocument, result,
		map[string]interface{}{"accountId": account, "createWebhookInput": input}, interceptors...)
	return result.Notification.Integration.CreateWebhook.Webhook, err
}

func (c *sdkClient) UpdateWebhook(
	ctx context.Context, account string, input m.UpdateWebhookInput,
	interceptors ...clientv2.RequestInterceptor,
) (*m.Webhook, error) {
	result := new(updateWebhookEnvelope)
	err := c.client.V2().Client.Post(ctx, "notificationIntegrationUpdateWebhook", updateDocument, result,
		map[string]interface{}{"accountId": account, "updateWebhookInput": input}, interceptors...)
	return result.Notification.Integration.UpdateWebhook.Webhook, err
}

func (c *sdkClient) DeleteWebhook(
	ctx context.Context, account string, input m.WebhookIntegrationRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*m.Webhook, error) {
	result := new(deleteWebhookEnvelope)
	err := c.client.V2().Client.Post(ctx, "notificationIntegrationDeleteWebhook", deleteDocument, result,
		map[string]interface{}{"accountId": account, "webhookIntegrationRefInput": input}, interceptors...)
	return result.Notification.Integration.DeleteWebhook.Webhook, err
}

func (c *sdkClient) NotificationIntegrationWebhook(
	ctx context.Context, account string, input m.WebhookIntegrationRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*m.Webhook, error) {
	result := new(readWebhookEnvelope)
	err := c.client.V2().Client.Post(ctx, "notificationIntegrationWebhook", readDocument, result,
		map[string]interface{}{"accountId": account, "webhookIntegrationRefInput": input}, interceptors...)
	if err != nil {
		return nil, err
	}
	if result.Notification == nil {
		return nil, fmt.Errorf("API did not return a notification response")
	}
	return result.Notification.Integration.Webhook, nil
}

func (c *sdkClient) WebhookIDByName(ctx context.Context, account, name string) (string, bool, error) {
	list, err := c.client.Catov2.NotificationIntegrationWebhookList(ctx, account)
	if err != nil {
		return "", false, err
	}
	if list == nil {
		return "", false, nil
	}
	for _, item := range list.GetNotification().GetIntegration().GetWebhookList().GetItems() {
		if item != nil && item.GetName() == name {
			return item.GetID(), true, nil
		}
	}
	return "", false, nil
}

func webhookRef(id string) m.WebhookIntegrationRefInput {
	return m.WebhookIntegrationRefInput{By: m.ObjectRefByID, Input: id}
}

func (r *webhookResource) find(ctx context.Context, id string) (*m.Webhook, error) {
	item, err := r.getClient().NotificationIntegrationWebhook(ctx, r.client.AccountId, webhookRef(id))
	if err != nil {
		if missingWebhook(err, r.client.AccountId, id) {
			return nil, nil
		}
		return nil, err
	}
	if item != nil && item.ID != id {
		return nil, fmt.Errorf("API returned a different webhook identity")
	}
	return item, nil
}

// missingWebhook sniffs the GraphQL not-found error for a webhook reference. Live-verified
// against badev4 (account 24425) on 2026-11-11 via internal/acctests/webhook: the backend
// uses the generic notification-integration wording ("subscription", not "webhook") for
// every integration type, matching mailing_list's missingList format exactly.
func missingWebhook(err error, account, id string) bool {
	expected := fmt.Sprintf("GQL-User-Error: Account ID [%s] subscription [%s] not found", account, id)
	return apperror.MatchesGraphQLMessages(err, expected)
}
