package subscriptiongroup

import (
	"context"
	"errors"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped subscription-group CRUD contract.
type ResourceClient interface {
	NotificationCreateSubscriptionGroup(context.Context, string, m.CreateSubscriptionGroupInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationCreateSubscriptionGroup, error)
	NotificationUpdateSubscriptionGroup(context.Context, string, m.UpdateSubscriptionGroupInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationUpdateSubscriptionGroup, error)
	NotificationDeleteSubscriptionGroup(context.Context, string, m.SubscriptionGroupRefInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationDeleteSubscriptionGroup, error)
	NotificationSubscriptionGroup(context.Context, string, m.SubscriptionGroupRefInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationSubscriptionGroup, error)
}
type sdkClient struct{ client *catoclient.Service }

// API-key actors may lack audit.updatedBy. Request only fields owned by this resource.
const subscriptionGroupSelection = "id name mailingList { id name } integration { webhookList { id name } }"
const readDocument = `query notificationSubscriptionGroup(
$accountId: ID!, $subscriptionGroupRefInput: SubscriptionGroupRefInput!) {
notification(accountId: $accountId) {
subscriptionGroup(input: $subscriptionGroupRefInput) { ` + subscriptionGroupSelection + ` } } }`
const createDocument = `mutation notificationCreateSubscriptionGroup(
$accountId: ID!, $createSubscriptionGroupInput: CreateSubscriptionGroupInput!) {
notification(accountId: $accountId) {
createSubscriptionGroup(input: $createSubscriptionGroupInput) {
subscriptionGroup { ` + subscriptionGroupSelection + ` } } } }`
const updateDocument = `mutation notificationUpdateSubscriptionGroup(
$accountId: ID!, $updateSubscriptionGroupInput: UpdateSubscriptionGroupInput!) {
notification(accountId: $accountId) {
updateSubscriptionGroup(input: $updateSubscriptionGroupInput) {
subscriptionGroup { ` + subscriptionGroupSelection + ` } } } }`
const deleteDocument = `mutation notificationDeleteSubscriptionGroup(
$accountId: ID!, $subscriptionGroupRefInput: SubscriptionGroupRefInput!) {
notification(accountId: $accountId) {
deleteSubscriptionGroup(input: $subscriptionGroupRefInput) {
subscriptionGroup { id } } } }`

func (c *sdkClient) NotificationCreateSubscriptionGroup(
	ctx context.Context, account string, input m.CreateSubscriptionGroupInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationCreateSubscriptionGroup, error) {
	result := new(cato.NotificationCreateSubscriptionGroup)
	err := c.client.V2().Client.Post(ctx, "notificationCreateSubscriptionGroup", createDocument, result,
		map[string]interface{}{"accountId": account, "createSubscriptionGroupInput": input}, interceptors...)
	return result, err
}

// Empty replacements are accompanied by removals computed from a fresh scoped read.
func (c *sdkClient) NotificationUpdateSubscriptionGroup(
	ctx context.Context, account string, input m.UpdateSubscriptionGroupInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationUpdateSubscriptionGroup, error) {
	result := new(cato.NotificationUpdateSubscriptionGroup)
	err := c.client.V2().Client.Post(ctx, "notificationUpdateSubscriptionGroup", updateDocument, result,
		map[string]interface{}{
			"accountId": account,
			"updateSubscriptionGroupInput": map[string]interface{}{
				"subscriptionGroup": input.SubscriptionGroup, "name": input.Name,
				"mailingList": input.MailingList, "integration": input.Integration,
				"mailingListToRemove": input.MailingListToRemove, "integrationToRemove": input.IntegrationToRemove,
			},
		}, interceptors...)
	return result, err
}
func (c *sdkClient) NotificationSubscriptionGroup(
	ctx context.Context, account string, input m.SubscriptionGroupRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationSubscriptionGroup, error) {
	result := new(cato.NotificationSubscriptionGroup)
	err := c.client.V2().Client.Post(ctx, "notificationSubscriptionGroup", readDocument, result,
		map[string]interface{}{"accountId": account, "subscriptionGroupRefInput": input}, interceptors...)
	return result, err
}
func (c *sdkClient) NotificationDeleteSubscriptionGroup(
	ctx context.Context, account string, input m.SubscriptionGroupRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationDeleteSubscriptionGroup, error) {
	result := new(cato.NotificationDeleteSubscriptionGroup)
	err := c.client.V2().Client.Post(ctx, "notificationDeleteSubscriptionGroup", deleteDocument, result,
		map[string]interface{}{"accountId": account, "subscriptionGroupRefInput": input}, interceptors...)
	return result, err
}

func listRef(id string) m.SubscriptionGroupRefInput {
	return m.SubscriptionGroupRefInput{By: m.ObjectRefByID, Input: id}
}
func (r *subscriptionGroupResource) find(ctx context.Context, id string) (
	*cato.NotificationSubscriptionGroup_Notification_SubscriptionGroup, error,
) {
	result, err := r.getClient().NotificationSubscriptionGroup(ctx, r.client.AccountId, listRef(id))
	if err != nil {
		if missingList(err, r.client.AccountId, id) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil || result.GetNotification() == nil {
		return nil, fmt.Errorf("API did not return a notification response")
	}
	item := result.GetNotification().GetSubscriptionGroup()
	if item != nil && item.GetID() != id {
		return nil, fmt.Errorf("API returned a different subscription-group identity")
	}
	return item, nil
}

func missingList(err error, account, id string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	expected := fmt.Sprintf("GQL-User-Error: Account ID [%s] subscription [%s] not found", account, id)
	for _, e := range *response.GqlErrors {
		if e == nil || e.Message != expected {
			return false
		}
	}
	return true
}
