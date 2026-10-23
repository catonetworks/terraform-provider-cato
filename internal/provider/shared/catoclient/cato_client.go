package catoclient

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// Service object for Cato client
type Service struct {
	BaseURL              string
	Token                string
	AccountId            string             //nolint:revive // Shared client field used across provider resources.
	Catov2               *ProviderSDKClient // *cato.Client
	AccountSnapshotCache *AccountSnapshotCache
}

func (p *Service) V2() *cato.Client  { return p.Catov2.Client }
func (p *Service) AccountID() string { return p.AccountId }

// EntityLookupMinimal forwards DHCP lookups through the provider SDK adapter.
func (p *Service) EntityLookupMinimal(
	ctx context.Context,
	accountID string,
	typeArg cato_models.EntityType,
	limit *int64,
	from *int64,
	parent *cato_models.EntityInput,
	sort []*cato_models.SortInput,
	filters []*cato_models.LookupFilterInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.EntityLookup, error) {
	return p.Catov2.EntityLookupMinimal(ctx, accountID, typeArg, limit, from, parent, sort, filters, interceptors...)
}
