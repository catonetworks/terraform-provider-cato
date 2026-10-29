//go:build acctest

package secondarygcpvsocket

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

// This fixture is explicitly disposable: the test imports, updates and deletes it.
// Creating an unregistered GCP socket cannot supply this fixture: the read API
// rejects it as a non-GCP socket and updates require REGISTERED status.
func TestAccSecondaryGCPVSocket(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv("TFACC_GCP_SECONDARY_ALLOW_MUTATIONS") != "1" {
		t.Skip("Requires a dedicated registered GCP secondary; test updates and deletes it. Set TFACC_GCP_SECONDARY_ALLOW_MUTATIONS=1.")
	}
	for _, key := range []string{"TFACC_GCP_SECONDARY_SOCKET_ID", "TFACC_GCP_INTERFACE_IP", "TFACC_GCP_LOAD_BALANCER_IP"} {
		if os.Getenv(key) == "" {
			t.Fatalf("Set %s for the dedicated registered fixture", key)
		}
	}
	acc.CheckCMAVars(t)()
	ctx := context.Background()
	client := acc.GetClient(t)
	id := os.Getenv("TFACC_GCP_SECONDARY_SOCKET_ID")
	result, err := client.SiteSecondaryGCPVSocket(ctx, acc.CatoAccountID, m.SecondaryGCPVSocketInput{ID: id})
	if err != nil {
		t.Fatalf("Read designated registered secondary: %s", err)
	}
	item := result.GetSite().GetSecondaryGCPVSocket().GetSocketInfo()
	if item == nil || item.GetID() != id || item.GetSite().GetID() == "" {
		t.Fatal("Designated fixture identity was not confirmed")
	}
	siteID := item.GetSite().GetID()
	config := func(updated bool) string {
		ip, balancer := item.GetInterfaceIP(), item.GetLoadBalancerIP()
		if updated {
			ip, balancer = os.Getenv("TFACC_GCP_INTERFACE_IP"), os.Getenv("TFACC_GCP_LOAD_BALANCER_IP")
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_secondary_gcp_vsocket" "test" {
 site = { id = %q }
 gcp_config = { interface_ip = %q, load_balancer_ip = %q }
}`, siteID, ip, balancer)
	}
	object := "cato_secondary_gcp_vsocket.test"
	initial := config(false) + fmt.Sprintf("\nimport {\n to = cato_secondary_gcp_vsocket.test\n id = %q\n}\n", id)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error { return checkDestroyedSecondary(t, siteID, state) }, Steps: []resource.TestStep{
			{Config: initial, Check: resource.TestCheckResourceAttr(object, "id", id)},
			{Config: config(false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(object, "gcp_config.interface_ip", os.Getenv("TFACC_GCP_INTERFACE_IP")),
				resource.TestCheckResourceAttr(object, "gcp_config.load_balancer_ip", os.Getenv("TFACC_GCP_LOAD_BALANCER_IP")))},
			{Config: config(true), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
func checkDestroyedSecondary(t *testing.T, siteID string, state *terraform.State) error {
	ctx := context.Background()
	client := acc.GetClient(t)
	snapshot, err := client.AccountSnapshot(ctx, []string{siteID}, []string{}, &acc.CatoAccountID)
	if err != nil {
		return err
	}
	account := snapshot.GetAccountSnapshot()
	if account == nil || account.GetID() == nil || *account.GetID() != acc.CatoAccountID {
		return fmt.Errorf("account-scoped snapshot missing")
	}
	sites := account.GetSites()
	if len(sites) != 1 || sites[0] == nil || sites[0].GetID() == nil || *sites[0].GetID() != siteID {
		return fmt.Errorf("parent fixture identity missing")
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "cato_secondary_gcp_vsocket" || item.Primary == nil {
			continue
		}
		if err := secondaryAbsent(sites[0].GetInfoSiteSnapshot(), item.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func secondaryAbsent(info *cato.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot, id string) error {
	for _, socket := range info.GetSockets() {
		if socket == nil || socket.GetID() == nil {
			return fmt.Errorf("socket identity missing")
		}
		if *socket.GetID() == id {
			return fmt.Errorf("designated secondary socket still exists")
		}
	}
	return nil
}
