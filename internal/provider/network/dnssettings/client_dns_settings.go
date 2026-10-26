package dnssettings

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

// ResourceClient is the account-scoped API contract for this resource.
type ResourceClient interface {
	NetworkConfigDNSSettings(ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSSettings,
		error)
	UpdateDNSSettings(ctx context.Context,
		accountID string,
		input nc.DNSSettingsInput) (*cato.NetworkConfigDNSUpdateSettings,
		error)
}
type sdkClient struct{ *nc.DNSClient }

var _ ResourceClient = (*sdkClient)(nil)

func (r *settingsResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{nc.NewDNSClient(r.client.Catov2)}
}
