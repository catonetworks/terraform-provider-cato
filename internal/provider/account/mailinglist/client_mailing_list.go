package mailinglist

import (
	"context"
	"errors"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped mailing-list CRUD contract.
type ResourceClient interface {
	NotificationCreateMailingList(context.Context, string, m.CreateMailingListInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationCreateMailingList, error)
	NotificationUpdateMailingList(context.Context, string, m.UpdateMailingListInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationUpdateMailingList, error)
	NotificationDeleteMailingList(context.Context, string, m.MailingListRefInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationDeleteMailingList, error)
	NotificationMailingList(context.Context, string, m.MailingListRefInput, ...clientv2.RequestInterceptor,
	) (*cato.NotificationMailingList, error)
}
type sdkClient struct{ client *catoclient.Service }

// API-key actors may lack audit.updatedBy. Request only fields owned by this resource.
const mailingListSelection = "id name address admin { id name }"
const readDocument = `query notificationMailingList($accountId: ID!, $mailingListRefInput: MailingListRefInput!) {
notification(accountId: $accountId) {
mailingList(input: $mailingListRefInput) { ` + mailingListSelection + ` } } }`
const createDocument = `mutation notificationCreateMailingList($accountId: ID!, $createMailingListInput: CreateMailingListInput!) {
notification(accountId: $accountId) {
createMailingList(input: $createMailingListInput) {
mailingList { ` + mailingListSelection + ` } } } }`
const updateDocument = `mutation notificationUpdateMailingList($accountId: ID!, $updateMailingListInput: UpdateMailingListInput!) {
notification(accountId: $accountId) {
updateMailingList(input: $updateMailingListInput) {
mailingList { ` + mailingListSelection + ` } } } }`
const deleteDocument = `mutation notificationDeleteMailingList($accountId: ID!, $mailingListRefInput: MailingListRefInput!) {
notification(accountId: $accountId) {
deleteMailingList(input: $mailingListRefInput) {
mailingList { id } } } }`

func (c *sdkClient) NotificationCreateMailingList(
	ctx context.Context, account string, input m.CreateMailingListInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationCreateMailingList, error) {
	result := new(cato.NotificationCreateMailingList)
	err := c.client.V2().Client.Post(ctx, "notificationCreateMailingList", createDocument, result,
		map[string]interface{}{"accountId": account, "createMailingListInput": input}, interceptors...)
	return result, err
}

// Explicit arrays are required to clear membership; the generated DTO uses omitempty.
func (c *sdkClient) NotificationUpdateMailingList(
	ctx context.Context, account string, input m.UpdateMailingListInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationUpdateMailingList, error) {
	result := new(cato.NotificationUpdateMailingList)
	err := c.client.V2().Client.Post(ctx, "notificationUpdateMailingList", updateDocument, result,
		map[string]interface{}{
			"accountId": account,
			"updateMailingListInput": map[string]interface{}{
				"mailingList": input.MailingList, "name": input.Name,
				"address": input.Address, "admin": input.Admin,
			},
		}, interceptors...)
	return result, err
}
func (c *sdkClient) NotificationMailingList(
	ctx context.Context, account string, input m.MailingListRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationMailingList, error) {
	result := new(cato.NotificationMailingList)
	err := c.client.V2().Client.Post(ctx, "notificationMailingList", readDocument, result,
		map[string]interface{}{"accountId": account, "mailingListRefInput": input}, interceptors...)
	return result, err
}
func (c *sdkClient) NotificationDeleteMailingList(
	ctx context.Context, account string, input m.MailingListRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.NotificationDeleteMailingList, error) {
	result := new(cato.NotificationDeleteMailingList)
	err := c.client.V2().Client.Post(ctx, "notificationDeleteMailingList", deleteDocument, result,
		map[string]interface{}{"accountId": account, "mailingListRefInput": input}, interceptors...)
	return result, err
}

func listRef(id string) m.MailingListRefInput {
	return m.MailingListRefInput{By: m.ObjectRefByID, Input: id}
}
func (r *mailingListResource) find(ctx context.Context, id string) (*cato.NotificationMailingList_Notification_MailingList, error) {
	result, err := r.getClient().NotificationMailingList(ctx, r.client.AccountId, listRef(id))
	if err != nil {
		if missingList(err, r.client.AccountId, id) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil || result.GetNotification() == nil {
		return nil, fmt.Errorf("API did not return a notification response")
	}
	item := result.GetNotification().GetMailingList()
	if item != nil && item.GetID() != id {
		return nil, fmt.Errorf("API returned a different mailing-list identity")
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
