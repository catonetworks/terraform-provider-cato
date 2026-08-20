//go:build acctest

package acc

import (
	"errors"
	"fmt"
	"os"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const (
	envEnableAccountCRUD        = "TFACC_ENABLE_ACCOUNT_CRUD"
	envEnableAccountCRUDAllowed = "TFACC_ACCOUNT_CRUD_ALLOWED"

	disableAccountDocument = `mutation accountManagementDisableAccount(
		$accountId: ID!
		$accountIdToDisable: ID!
	) {
		accountManagement(accountId: $accountId) {
			disableAccount(accountId: $accountIdToDisable) {
				accountInfo {
					id
					status
				}
			}
		}
	}`
)

type disableAccountResponse struct {
	AccountManagement *disableAccountManagement `json:"accountManagement"`
}

type disableAccountManagement struct {
	DisableAccount *disableAccountPayload `json:"disableAccount"`
}

type disableAccountPayload struct {
	AccountInfo *disabledAccountInfo `json:"accountInfo"`
}

type disabledAccountInfo struct {
	ID     string                    `json:"id"`
	Status cato_models.AccountStatus `json:"status"`
}

func deleteSitesAndDependencies(t *testing.T) error {
	client := GetClient(t)
	sites := selectAcctestRefs(getEntities(t, resSite), "")
	if !sameIDSet(idsFromRefs(sites), idsFromRefs(sites)) {
		return errors.New("acctest site lookup returned blank or duplicate IDs")
	}

	var cleanupErrors []error
	for _, site := range sites {
		if err := deleteSiteBGPPeers(client, site); err != nil {
			cleanupErrors = append(cleanupErrors, err)
			continue
		}

		result, err := client.SiteRemoveSite(ctx, site.ID, CatoAccountID)
		if err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("deleting site %s (%s): %w", site.Name, site.ID, err))
			continue
		}
		if removedID := result.GetSite().GetRemoveSite().GetSiteID(); removedID != site.ID {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"deleting site %s (%s): response ID %q", site.Name, site.ID, removedID,
			))
		}
	}

	return errors.Join(cleanupErrors...)
}

func deleteSiteBGPPeers(client *cato.Client, site Ref) error {
	input := cato_models.BgpPeerListInput{
		Site: &cato_models.SiteRefInput{
			By:    cato_models.ObjectRefByID,
			Input: site.ID,
		},
	}
	result, err := client.SiteBgpPeerList(ctx, input, CatoAccountID)
	if err != nil {
		return fmt.Errorf("listing BGP peers for site %s (%s): %w", site.Name, site.ID, err)
	}
	if result == nil || result.GetSite() == nil || result.GetSite().GetBgpPeerList() == nil {
		return fmt.Errorf("listing BGP peers for site %s (%s): missing response payload", site.Name, site.ID)
	}

	peerList := result.GetSite().GetBgpPeerList()
	peers := peerList.GetBgpPeerBgpPeerListPayload()
	if peerList.GetTotal() != int64(len(peers)) {
		return fmt.Errorf(
			"listing BGP peers for site %s (%s): got %d items, total is %d",
			site.Name, site.ID, len(peers), peerList.GetTotal(),
		)
	}

	peerIDs := make([]string, 0, len(peers))
	var cleanupErrors []error
	for _, peer := range peers {
		if peer == nil {
			continue
		}
		if !acctestRE.MatchString(peer.GetName()) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"refusing to delete site %s (%s): contains non-acctest BGP peer %s (%s)",
				site.Name,
				site.ID,
				peer.GetName(),
				peer.GetID(),
			))
			continue
		}
		peerIDs = append(peerIDs, peer.GetID())
	}
	if !sameIDSet(peerIDs, peerIDs) {
		return fmt.Errorf("listing BGP peers for site %s (%s): blank or duplicate peer IDs", site.Name, site.ID)
	}

	for _, peerID := range peerIDs {
		removeInput := cato_models.RemoveBgpPeerInput{ID: peerID}
		removeResult, removeErr := client.SiteRemoveBgpPeer(ctx, removeInput, CatoAccountID)
		if removeErr != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"deleting BGP peer %s from site %s (%s): %w", peerID, site.Name, site.ID, removeErr,
			))
			continue
		}
		if removedID := removeResult.GetSite().GetRemoveBgpPeer().GetBgpPeer().GetID(); removedID != peerID {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"deleting BGP peer %s from site %s (%s): response ID %q",
				peerID, site.Name, site.ID, removedID,
			))
		}
	}

	return errors.Join(cleanupErrors...)
}

func deleteAcctestGlobalIPRanges(t *testing.T) error {
	client := GetClient(t)
	result, err := client.ObjectGlobalIPRangeList(ctx, CatoAccountID, nil)
	if err != nil {
		return fmt.Errorf("listing global IP ranges: %w", err)
	}
	refs, err := globalIPRangeRefs(result)
	if err != nil {
		return err
	}
	return deleteGlobalIPRangeRefs(client, selectAcctestRefs(refs, ""))
}

func globalIPRangeRefs(result *cato.ObjectGlobalIPRangeList) ([]Ref, error) {
	if result == nil || result.GetObject() == nil || result.GetObject().GetGlobalIPRangeList() == nil {
		return nil, errors.New("listing global IP ranges: missing response payload")
	}

	rangeList := result.GetObject().GetGlobalIPRangeList()
	items := rangeList.GetItems()
	if rangeList.GetTotal() != int64(len(items)) {
		return nil, fmt.Errorf("listing global IP ranges: got %d items, total is %d", len(items), rangeList.GetTotal())
	}

	refs := make([]Ref, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		refs = append(refs, Ref{ID: item.GetID(), Name: item.GetName()})
	}
	return refs, nil
}

func deleteGlobalIPRangeRefs(client *cato.Client, refs []Ref) error {
	expectedIDs := idsFromRefs(refs)
	if !sameIDSet(expectedIDs, expectedIDs) {
		return errors.New("acctest global IP range lookup returned duplicate IDs")
	}
	if len(expectedIDs) == 0 {
		return nil
	}

	input := make([]*cato_models.GlobalIPRangeRefInput, 0, len(expectedIDs))
	for _, id := range expectedIDs {
		input = append(input, &cato_models.GlobalIPRangeRefInput{
			By:    cato_models.ObjectRefByID,
			Input: id,
		})
	}
	removeResult, err := client.ObjectDeleteGlobalIPRangeBulk(ctx, CatoAccountID, input)
	if err != nil {
		return fmt.Errorf("deleting acctest global IP ranges: %w", err)
	}
	if removeResult == nil ||
		removeResult.GetObject() == nil ||
		removeResult.GetObject().GetDeleteGlobalIPRangeBulk() == nil {
		return errors.New("deleting acctest global IP ranges: missing response payload")
	}

	removed := removeResult.GetObject().GetDeleteGlobalIPRangeBulk().GetGlobalIPRange()
	removedIDs := make([]string, 0, len(removed))
	for _, item := range removed {
		removedIDs = append(removedIDs, item.GetID())
	}
	if !sameIDSet(expectedIDs, removedIDs) {
		return fmt.Errorf(
			"deleting acctest global IP ranges: returned IDs %v, expected %v",
			removedIDs, expectedIDs,
		)
	}

	return nil
}

func deleteAcctestAccounts(t *testing.T) error {
	if os.Getenv(envEnableAccountCRUD) != "true" || os.Getenv(envEnableAccountCRUDAllowed) != "true" {
		return nil
	}

	client := GetClient(t)
	accounts := selectAcctestRefs(getEntities(t, resAccount), CatoAccountID)
	accountIDs := idsFromRefs(accounts)
	if !sameIDSet(accountIDs, accountIDs) {
		return errors.New("acctest subaccount lookup returned duplicate IDs")
	}

	var cleanupErrors []error
	for _, account := range accounts {
		if err := disableAcctestAccount(client, account); err != nil {
			cleanupErrors = append(cleanupErrors, err)
			continue
		}

		removeResult, removeErr := client.AccountManagementRemoveAccount(ctx, account.ID, CatoAccountID)
		if removeErr != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"removing subaccount %s (%s): %w", account.Name, account.ID, removeErr,
			))
			continue
		}
		removedID := removeResult.GetAccountManagement().GetRemoveAccount().GetAccountInfo().GetID()
		if removedID != account.ID {
			cleanupErrors = append(cleanupErrors, fmt.Errorf(
				"removing subaccount %s (%s): response ID %q", account.Name, account.ID, removedID,
			))
		}
	}

	return errors.Join(cleanupErrors...)
}

func disableAcctestAccount(client *cato.Client, account Ref) error {
	var result disableAccountResponse
	variables := map[string]any{
		"accountId":          CatoAccountID,
		"accountIdToDisable": account.ID,
	}
	if err := client.Client.Post(
		ctx,
		"accountManagementDisableAccount",
		disableAccountDocument,
		&result,
		variables,
	); err != nil {
		return fmt.Errorf("disabling subaccount %s (%s): %w", account.Name, account.ID, err)
	}
	if result.AccountManagement == nil ||
		result.AccountManagement.DisableAccount == nil ||
		result.AccountManagement.DisableAccount.AccountInfo == nil {
		return fmt.Errorf("disabling subaccount %s (%s): missing response payload", account.Name, account.ID)
	}

	accountInfo := result.AccountManagement.DisableAccount.AccountInfo
	if accountInfo.ID != account.ID {
		return fmt.Errorf(
			"disabling subaccount %s (%s): response ID %q", account.Name, account.ID, accountInfo.ID,
		)
	}
	if accountInfo.Status != cato_models.AccountStatusDisabled {
		return fmt.Errorf(
			"disabling subaccount %s (%s): response status %q",
			account.Name, account.ID, accountInfo.Status,
		)
	}

	return nil
}

func selectAcctestRefs(refs []Ref, excludedID string) []Ref {
	selected := make([]Ref, 0, len(refs))
	for _, ref := range refs {
		if ref.ID == "" || ref.ID == excludedID || !acctestRE.MatchString(ref.Name) {
			continue
		}
		selected = append(selected, ref)
	}
	return selected
}

func idsFromRefs(refs []Ref) []string {
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	return ids
}

func sameIDSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	leftSet := make(map[string]struct{}, len(left))
	for _, id := range left {
		if id == "" {
			return false
		}
		if _, duplicate := leftSet[id]; duplicate {
			return false
		}
		leftSet[id] = struct{}{}
	}

	rightSet := make(map[string]struct{}, len(right))
	for _, id := range right {
		if id == "" {
			return false
		}
		if _, duplicate := rightSet[id]; duplicate {
			return false
		}
		if _, expected := leftSet[id]; !expected {
			return false
		}
		rightSet[id] = struct{}{}
	}

	return true
}
