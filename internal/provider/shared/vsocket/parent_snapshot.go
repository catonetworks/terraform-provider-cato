package vsocket

import (
	"fmt"
	"slices"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
)

// ValidateEmptySlot confirms account, parent identity, compatible cloud and a single known primary socket.
func ValidateEmptySlot(result *cato.AccountSnapshot, account, id, cloud string, kinds ...m.ProtoType) error {
	snapshot := result.GetAccountSnapshot()
	if snapshot == nil || snapshot.GetID() == nil || *snapshot.GetID() != account {
		return fmt.Errorf("snapshot did not confirm the configured account")
	}
	sites := snapshot.GetSites()
	if len(sites) != 1 || sites[0] == nil || sites[0].GetID() == nil || *sites[0].GetID() != id {
		return fmt.Errorf("snapshot did not confirm the parent site")
	}
	return ValidatePrimary(sites[0].GetInfoSiteSnapshot(), cloud, kinds...)
}

// ValidatePrimary rejects incompatible clouds or occupied/unknown secondary slots.
func ValidatePrimary(info *cato.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot, cloud string, kinds ...m.ProtoType) error {
	kind := info.GetConnType()
	if kind == nil || !slices.Contains(kinds, *kind) {
		return fmt.Errorf("parent must be a compatible %s vSocket site", cloud)
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
