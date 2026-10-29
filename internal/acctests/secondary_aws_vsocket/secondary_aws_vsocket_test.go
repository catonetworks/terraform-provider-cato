//go:build acctest

package secondaryawsvsocket

import (
	"context"
	"fmt"
	"os"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSecondaryAWSVSocket(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv("TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE") != "1" {
		t.Skip("Requires ADD_SECONDARY_ON_HA_SITE_CREATION; set TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE=1 on an enabled test account")
	}
	acc.CheckCMAVars(t)()
	ctx := context.Background()
	client := acc.GetClient(t)
	result, err := client.SiteAddSocketSite(ctx, m.AddSocketSiteInput{ConnectionType: m.SiteConnectionTypeEnumSocketAWS1500, Name: acc.GetRandName("secondary_aws_site"), NativeNetworkRange: acc.GetRandNetworkPrefix() + ".0/24", SiteType: m.SiteTypeBranch, SiteLocation: &m.AddSiteLocationInput{CountryCode: "GB", Timezone: "Europe/London"}}, acc.CatoAccountID)
	if err != nil {
		t.Fatalf("create dedicated AWS site: %s", err)
	}
	siteID := result.GetSite().GetAddSocketSite().GetSiteID()
	if siteID == "" {
		t.Fatal("site creation returned no ID")
	}
	// Cleanup is limited to the new parent and its own secondary socket, even after test failure.
	t.Cleanup(func() { cleanupOwnedSite(t, siteID) })
	config := func(updated bool) string {
		ip, table := "10.253.250.10", "rtb-acctest"
		if updated {
			ip, table = "10.253.250.11", "rtb-acctest-updated"
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_secondary_aws_vsocket" "test" {
 site = { id = %q }
 eni_ip_address = %q
 eni_ip_subnet = "10.253.250.0/24"
 route_table_id = %q
}`, siteID, ip, table)
	}
	object := "cato_secondary_aws_vsocket.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: func(state *terraform.State) error { return checkDestroyedSecondary(t, siteID, state) }, Steps: []resource.TestStep{
		{Config: config(false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "site.id", siteID))},
		{Config: config(false), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
		{Config: config(true), Check: resource.TestCheckResourceAttr(object, "eni_ip_address", "10.253.250.11")},
		{Config: config(true), PlanOnly: true, ExpectNonEmptyPlan: false},
	}})
}

func cleanupOwnedSite(t *testing.T, siteID string) {
	ctx := context.Background()
	client := acc.GetClient(t)
	snapshot, err := client.AccountSnapshot(ctx, []string{siteID}, []string{}, &acc.CatoAccountID)
	if err != nil {
		t.Errorf("read owned fixture before cleanup: %s", err)
		return
	}
	sites := snapshot.GetAccountSnapshot().GetSites()
	if len(sites) != 1 || sites[0].GetID() == nil || *sites[0].GetID() != siteID {
		t.Error("could not confirm owned parent fixture; cleanup retained it")
		return
	}
	if !cleanupOwnedSecondary(t, siteID, sites[0]) {
		return
	}
	if _, err := client.SiteRemoveSite(ctx, siteID, acc.CatoAccountID); err != nil {
		t.Errorf("remove owned AWS site: %s", err)
		return
	}
	limit, from := int64(1), int64(0)
	lookup, err := client.EntityLookup(ctx, acc.CatoAccountID, m.EntityTypeSite, &limit, &from, nil, nil, []string{siteID}, nil, nil, nil)
	if err != nil || lookup.GetEntityLookup() == nil || len(lookup.GetEntityLookup().GetItems()) != 0 {
		t.Error("owned AWS site deletion not confirmed")
	}
}

func checkDestroyedSecondary(t *testing.T, siteID string, state *terraform.State) error {
	ctx := context.Background()
	client := acc.GetClient(t)
	for _, item := range state.RootModule().Resources {
		if item.Type != "cato_secondary_aws_vsocket" || item.Primary == nil {
			continue
		}
		snapshot, err := client.AccountSnapshot(ctx, []string{siteID}, []string{}, &acc.CatoAccountID)
		if err != nil {
			return err
		}
		for _, site := range snapshot.GetAccountSnapshot().GetSites() {
			for _, socket := range site.GetInfoSiteSnapshot().GetSockets() {
				if socket.GetID() != nil && *socket.GetID() == item.Primary.ID {
					return fmt.Errorf("owned secondary socket still exists")
				}
			}
		}
	}
	return nil
}

func cleanupOwnedSecondary(t *testing.T, siteID string, site *cato.AccountSnapshot_AccountSnapshot_Sites) bool {
	ctx := context.Background()
	client := acc.GetClient(t)
	for _, socket := range site.GetInfoSiteSnapshot().GetSockets() {
		if socket == nil || socket.GetID() == nil || socket.GetIsPrimary() == nil {
			t.Error("unknown socket identity; cleanup retained fixture")
			return false
		}
		if !*socket.GetIsPrimary() {
			current, err := client.SiteSecondaryAWSVSocket(ctx, acc.CatoAccountID, *socket.GetID())
			if err != nil {
				t.Errorf("read owned secondary before cleanup: %s", err)
				return false
			}
			if current.GetSite().GetSecondaryAWSVSocket().GetSite().GetID() != siteID {
				t.Error("secondary does not belong to owned fixture")
				return false
			}
			if _, err = client.SiteRemoveSecondaryAWSVSocket(ctx, acc.CatoAccountID, *socket.GetID()); err != nil {
				t.Errorf("remove owned secondary fixture: %s", err)
				return false
			}
		}
	}
	return true
}
