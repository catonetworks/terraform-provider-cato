package secondaryazurevsocket

import (
	"context"
	"fmt"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/vsocket"
)

// ResourceClient is the account-scoped secondary Azure vSocket contract.
type ResourceClient interface {
	SiteAddSecondaryAzureVSocket(context.Context,
		string, m.AddSecondaryAzureVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteAddSecondaryAzureVSocket, error)
	SiteUpdateSecondaryAzureVSocket(context.Context,
		string, m.UpdateSecondaryAzureVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSecondaryAzureVSocket, error)
	SiteRemoveSecondaryAzureVSocket(context.Context, string, string,
		...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveSecondaryAzureVSocket, error)
	SiteSecondaryAzureVSocket(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.SiteSecondaryAzureVSocket, error)
	SiteGeneralDetails(context.Context, m.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	AccountSnapshot(context.Context, []string, []string, *string, ...clientv2.RequestInterceptor) (*cato.AccountSnapshot, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteAddSecondaryAzureVSocket(
	ctx context.Context, account string, in m.AddSecondaryAzureVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteAddSecondaryAzureVSocket, error) {
	return c.client.V2().SiteAddSecondaryAzureVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteUpdateSecondaryAzureVSocket(
	ctx context.Context, account string, in m.UpdateSecondaryAzureVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateSecondaryAzureVSocket, error) {
	return c.client.V2().SiteUpdateSecondaryAzureVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteRemoveSecondaryAzureVSocket(
	ctx context.Context, account, id string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveSecondaryAzureVSocket, error) {
	return c.client.V2().SiteRemoveSecondaryAzureVSocket(ctx, account, id, i...)
}
func (c *sdkClient) SiteSecondaryAzureVSocket(
	ctx context.Context, account, id string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteSecondaryAzureVSocket, error) {
	return c.client.V2().SiteSecondaryAzureVSocket(ctx, account, id, i...)
}
func (c *sdkClient) SiteGeneralDetails(
	ctx context.Context, in m.SiteRefInput, account string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteGeneralDetails, error) {
	return c.client.V2().SiteGeneralDetails(ctx, in, account, i...)
}
func (c *sdkClient) AccountSnapshot(
	ctx context.Context, sites, users []string, account *string,
	i ...clientv2.RequestInterceptor,
) (*cato.AccountSnapshot, error) {
	return c.client.V2().AccountSnapshot(ctx, sites, users, account, i...)
}

type socketRecord = cato.SiteSecondaryAzureVSocket_Site_SecondaryAzureVSocket

func missingSocket(err error, id string) bool {
	message := fmt.Sprintf("Socket ID %s was not found", id)
	return apperror.MatchesGraphQLMessages(err, message, "GQL-User-Error: "+message)
}
func (r *socketResource) find(ctx context.Context, id string) (*socketRecord, error) {
	result, err := r.getClient().SiteSecondaryAzureVSocket(ctx, r.client.AccountId, id)
	if err != nil {
		if missingSocket(err, id) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetSite().GetSecondaryAzureVSocket()
	if item == nil || item.GetID() != id || item.GetSite().GetID() == "" {
		return nil, fmt.Errorf("API did not return the requested secondary socket identity")
	}
	return item, nil
}
func (r *socketResource) resolveEmptySlot(ctx context.Context, ref *m.SiteRefInput) (string, error) {
	details, err := r.getClient().SiteGeneralDetails(ctx, *ref, r.client.AccountId)
	if err != nil {
		return "", err
	}
	site := details.GetSite().GetSiteGeneralDetails().GetSite()
	if site.GetID() == "" || !matchesReference(ref, site.GetID(), site.GetName()) {
		return "", fmt.Errorf("site lookup returned an inconsistent identity")
	}
	account := r.client.AccountId
	result, err := r.getClient().AccountSnapshot(ctx, []string{site.GetID()}, []string{}, &account)
	if err != nil {
		return "", err
	}
	if err := validateEmptySlot(result, account, site.GetID()); err != nil {
		return "", err
	}
	return site.GetID(), nil
}
func validateEmptySlot(result *cato.AccountSnapshot, account, id string) error {
	return vsocket.ValidateEmptySlot(result, account, id, "Azure", m.ProtoTypeVsocketVgxAzure, m.ProtoTypeSocketAz1500)
}

const consistencyTimeout = 30 * time.Second

func (r *socketResource) waitForSocket(ctx context.Context, id string, in m.AddSecondaryAzureVSocketInput) (*socketRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, consistencyTimeout)
	defer cancel()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		item, err := r.find(ctx, id)
		if err != nil {
			return nil, err
		}
		if item != nil && matchesSocket(item, in) {
			return item, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for secondary socket convergence: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
