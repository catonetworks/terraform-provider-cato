package catoapi

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

var _ network.CreateStaticHostAdapter = (*Adapter)(nil)

func (a *Adapter) CreateStaticHost(ctx context.Context, accountID, siteID string, host entities.StaticHost) (string, error) {
	const operation = "create static host"
	result, err := a.client.SiteAddStaticHost(ctx, siteID, cato_models.AddStaticHostInput{
		Name: host.Name, IP: host.IP, MacAddress: host.MacAddress,
	}, accountID)
	if err != nil {
		return "", requestError(operation, err)
	}
	if result == nil || result.Site.AddStaticHost == nil || result.Site.AddStaticHost.HostID == "" {
		return "", invalidResponse(operation)
	}
	return result.Site.AddStaticHost.HostID, nil
}
