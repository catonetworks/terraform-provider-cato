package catoclient

import (
	cato "github.com/catonetworks/cato-go-sdk"
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
