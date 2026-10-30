package customapplication

import (
	"context"
	"fmt"
	"regexp"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped custom application contract.
type ResourceClient interface {
	CustomAppDataAddCustomApplication(context.Context,
		string,
		m.AddCustomApplicationInput,
		...clientv2.RequestInterceptor) (*cato.CustomAppDataAddCustomApplication,
		error)
	CustomAppDataUpdateCustomApplication(context.Context,
		string,
		m.UpdateCustomApplicationInput,
		...clientv2.RequestInterceptor) (*cato.CustomAppDataUpdateCustomApplication,
		error)
	CustomAppDataDeleteCustomApplication(context.Context,
		string,
		m.DeleteCustomApplicationInput,
		...clientv2.RequestInterceptor) (*cato.CustomAppDataDeleteCustomApplication,
		error)
	CustomAppDataCustomApplication(context.Context,
		string,
		m.CustomApplicationRefInput,
		...clientv2.RequestInterceptor) (*cato.CustomAppDataCustomApplication,
		error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) CustomAppDataAddCustomApplication(ctx context.Context,
	account string,
	in m.AddCustomApplicationInput,
	i ...clientv2.RequestInterceptor) (*cato.CustomAppDataAddCustomApplication,
	error) {
	return c.client.V2().CustomAppDataAddCustomApplication(ctx, account, in, i...)
}
func (c *sdkClient) CustomAppDataUpdateCustomApplication(ctx context.Context,
	account string,
	in m.UpdateCustomApplicationInput,
	i ...clientv2.RequestInterceptor) (*cato.CustomAppDataUpdateCustomApplication,
	error) {
	// Preserve explicit category [] despite the SDK's omitempty tag.
	result := new(cato.CustomAppDataUpdateCustomApplication)
	err := c.client.V2().Client.Post(ctx,
		"customAppDataUpdateCustomApplication",
		cato.CustomAppDataUpdateCustomApplicationDocument,
		result,
		map[string]interface{}{
			"accountId": account,
			"updateCustomApplicationInput": map[string]interface{}{"id": in.ID,
				"name":        in.Name,
				"description": in.Description,
				"category":    in.Category,
				"criteria":    in.Criteria},
		}, i...)
	return result, err
}
func (c *sdkClient) CustomAppDataDeleteCustomApplication(ctx context.Context,
	account string,
	in m.DeleteCustomApplicationInput,
	i ...clientv2.RequestInterceptor) (*cato.CustomAppDataDeleteCustomApplication,
	error) {
	return c.client.V2().CustomAppDataDeleteCustomApplication(ctx, account, in, i...)
}
func (c *sdkClient) CustomAppDataCustomApplication(ctx context.Context,
	account string,
	in m.CustomApplicationRefInput,
	i ...clientv2.RequestInterceptor) (*cato.CustomAppDataCustomApplication,
	error) {
	return c.client.V2().CustomAppDataCustomApplication(ctx, account, in, i...)
}

type appRecord = cato.CustomAppDataCustomApplication_CustomAppData_CustomApplication

var numericID = regexp.MustCompile(`^\d+$`)

func missingApplication(err error, id, account string) bool {
	lookup := id
	if numericID.MatchString(id) {
		lookup = "CustomApp_" + account + "_" + id
	}
	expected := "Custom application not found by ID: " + lookup
	return apperror.MatchesGraphQLMessages(err, expected, "GQL-User-Error: "+expected)
}
func (r *appResource) find(ctx context.Context, id string) (*appRecord, error) {
	result,
		err := r.getClient().CustomAppDataCustomApplication(ctx,
		r.client.AccountId,
		m.CustomApplicationRefInput{By: m.ObjectRefByID,
			Input: id})
	if err != nil {
		if missingApplication(err, id, r.client.AccountId) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetCustomAppData().GetCustomApplication()
	if item == nil || item.GetID() != id {
		return nil, fmt.Errorf("API did not confirm the requested application identity")
	}
	return item, nil
}
