package secondaryawsvsocket

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

// ResourceClient is the account-scoped secondary AWS vSocket contract.
type ResourceClient interface {
	SiteAddSecondaryAWSVSocket(context.Context,
		string, m.AddSecondaryAWSVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteAddSecondaryAWSVSocket, error)
	SiteUpdateSecondaryAWSVSocket(context.Context,
		string, m.UpdateSecondaryAWSVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSecondaryAWSVSocket, error)
	SiteRemoveSecondaryAWSVSocket(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.SiteRemoveSecondaryAWSVSocket, error)
	SiteSecondaryAWSVSocket(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.SiteSecondaryAWSVSocket, error)
	SiteGeneralDetails(context.Context, m.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	AccountSnapshot(context.Context, []string, []string, *string, ...clientv2.RequestInterceptor) (*cato.AccountSnapshot, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteAddSecondaryAWSVSocket(
	ctx context.Context, account string, in m.AddSecondaryAWSVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteAddSecondaryAWSVSocket, error) {
	return c.client.V2().SiteAddSecondaryAWSVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteUpdateSecondaryAWSVSocket(
	ctx context.Context, account string, in m.UpdateSecondaryAWSVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateSecondaryAWSVSocket, error) {
	return c.client.V2().SiteUpdateSecondaryAWSVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteRemoveSecondaryAWSVSocket(
	ctx context.Context, account, id string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveSecondaryAWSVSocket, error) {
	return c.client.V2().SiteRemoveSecondaryAWSVSocket(ctx, account, id, i...)
}
func (c *sdkClient) SiteSecondaryAWSVSocket(
	ctx context.Context, account, id string,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteSecondaryAWSVSocket, error) {
	return c.client.V2().SiteSecondaryAWSVSocket(ctx, account, id, i...)
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

type socketRecord = cato.SiteSecondaryAwsVSocket_Site_SecondaryAWSVSocket

func missingSocket(err error, id string) bool {
	message := fmt.Sprintf("Socket ID %s was not found", id)
	return apperror.MatchesGraphQLMessages(err, message, "GQL-User-Error: "+message)
}
func (r *socketResource) find(ctx context.Context, id string) (*socketRecord, error) {
	result, err := r.getClient().SiteSecondaryAWSVSocket(ctx, r.client.AccountId, id)
	if err != nil {
		if missingSocket(err, id) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetSite().GetSecondaryAWSVSocket()
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
	return vsocket.ValidateEmptySlot(result, account, id, "AWS", m.ProtoTypeVsocketVgxAWS, m.ProtoTypeSocketAWS1500)
}

const consistencyTimeout = 30 * time.Second

func (r *socketResource) waitForSocket(ctx context.Context, id string, in m.AddSecondaryAWSVSocketInput) (*socketRecord, error) {
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
