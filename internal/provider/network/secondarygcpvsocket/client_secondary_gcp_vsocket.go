package secondarygcpvsocket

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped secondary GCP vSocket contract.
type ResourceClient interface {
	SiteAddSecondaryGCPVSocket(context.Context,
		string, m.AddSecondaryGCPVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteAddSecondaryGCPVSocket, error)
	SiteUpdateSecondaryGCPVSocket(context.Context,
		string, m.UpdateSecondaryGCPVSocketInput, ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSecondaryGCPVSocket, error)
	SiteRemoveSecondaryGCPVSocket(context.Context, string, m.RemoveSecondaryGCPVSocketInput,
		...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveSecondaryGCPVSocket, error)
	SiteSecondaryGCPVSocket(context.Context, string, m.SecondaryGCPVSocketInput,
		...clientv2.RequestInterceptor,
	) (*cato.SiteSecondaryGCPVSocket, error)
	SiteGeneralDetails(context.Context, m.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	AccountSnapshot(context.Context, []string, []string, *string, ...clientv2.RequestInterceptor) (*cato.AccountSnapshot, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteAddSecondaryGCPVSocket(
	ctx context.Context, account string, in m.AddSecondaryGCPVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteAddSecondaryGCPVSocket, error) {
	return c.client.V2().SiteAddSecondaryGCPVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteUpdateSecondaryGCPVSocket(
	ctx context.Context, account string, in m.UpdateSecondaryGCPVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateSecondaryGCPVSocket, error) {
	return c.client.V2().SiteUpdateSecondaryGCPVSocket(ctx, account, in, i...)
}
func (c *sdkClient) SiteRemoveSecondaryGCPVSocket(
	ctx context.Context, account string, input m.RemoveSecondaryGCPVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveSecondaryGCPVSocket, error) {
	return c.client.V2().SiteRemoveSecondaryGCPVSocket(ctx, account, input, i...)
}
func (c *sdkClient) SiteSecondaryGCPVSocket(
	ctx context.Context, account string, input m.SecondaryGCPVSocketInput,
	i ...clientv2.RequestInterceptor,
) (*cato.SiteSecondaryGCPVSocket, error) {
	return c.client.V2().SiteSecondaryGCPVSocket(ctx, account, input, i...)
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

type socketRecord = cato.SiteSecondaryGcpVSocket_Site_SecondaryGCPVSocket_SocketInfo

func missingSocket(err error, id string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	message := fmt.Sprintf("Socket with ID: %s not found", id)
	wireMessage := fmt.Sprintf("404 Not Found: \"SOCKET with id [%s] not found\"", id)
	for _, item := range *response.GqlErrors {
		if item == nil || (item.Message != message && item.Message != "GQL-User-Error: "+message && item.Message != wireMessage) {
			return false
		}
	}
	return true
}
func (r *socketResource) find(ctx context.Context, id string) (*socketRecord, error) {
	result, err := r.getClient().SiteSecondaryGCPVSocket(ctx, r.client.AccountId, m.SecondaryGCPVSocketInput{ID: id})
	if err != nil {
		if missingSocket(err, id) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetSite().GetSecondaryGCPVSocket().GetSocketInfo()
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
	snapshot := result.GetAccountSnapshot()
	if snapshot == nil || snapshot.GetID() == nil || *snapshot.GetID() != account {
		return fmt.Errorf("snapshot did not confirm the configured account")
	}
	sites := snapshot.GetSites()
	if len(sites) != 1 || sites[0] == nil || sites[0].GetID() == nil || *sites[0].GetID() != id {
		return fmt.Errorf("snapshot did not confirm the parent site")
	}
	return validatePrimary(sites[0].GetInfoSiteSnapshot())
}
func validatePrimary(info *cato.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot) error {
	kind := info.GetConnType()
	if kind == nil || *kind != m.ProtoTypeSocketGCP1500 {
		return fmt.Errorf("parent must be a compatible GCP vSocket site")
	}
	sockets := info.GetSockets()
	if len(sockets) != 1 || sockets[0] == nil {
		return fmt.Errorf("site must have only a primary socket; import an existing secondary instead")
	}
	socket := sockets[0]
	if socket.GetID() == nil || *socket.GetID() == "" || socket.GetIsPrimary() == nil || !*socket.GetIsPrimary() {
		return fmt.Errorf("site must have a known primary socket")
	}
	return nil
}

const consistencyTimeout = 30 * time.Second

func (r *socketResource) waitForSocket(ctx context.Context, id string, in m.AddSecondaryGCPVSocketInput) (*socketRecord, error) {
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
