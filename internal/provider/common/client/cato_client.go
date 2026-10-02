package client

import (
	cato "github.com/catonetworks/cato-go-sdk"
)

// CatoClientData added by JF to support use of two different clients (long story....)
type CatoClientData struct {
	BaseURL              string
	Token                string
	AccountId            string //nolint:revive // Shared client field used across provider resources.
	Catov2               *cato.Client
	AccountSnapshotCache *AccountSnapshotCache
}

func (p *CatoClientData) V2() *cato.Client  { return p.Catov2 }
func (p *CatoClientData) AccountID() string { return p.AccountId }
