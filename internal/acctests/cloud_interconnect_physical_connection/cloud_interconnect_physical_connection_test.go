//go:build acctest

package cloudinterconnectphysicalconnection

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccCloudInterconnectPhysicalConnection(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	pop := os.Getenv("TFACC_CLOUD_INTERCONNECT_TEST_POP_ID")
	alternate := os.Getenv("TFACC_CLOUD_INTERCONNECT_ALT_POP_ID")
	if pop == "" {
		t.Skip("Set TFACC_CLOUD_INTERCONNECT_TEST_POP_ID to an isolated CrossConnetMock DOT1Q test PoP")
	}
	acc.CheckCMAVars(t)()
	validateMockPOP(t, pop)
	if alternate != "" {
		validateMockPOP(t, alternate)
	}
	ctx := context.Background()
	client := acc.GetClient(t)
	result, err := client.SiteAddCloudInterconnectSite(ctx, m.AddCloudInterconnectSiteInput{Name: acc.GetRandName("physical_site"), SiteType: m.SiteTypeBranch, SiteLocation: &m.AddSiteLocationInput{CountryCode: "GB", Timezone: "Europe/London"}}, acc.CatoAccountID)
	if err != nil {
		t.Fatalf("create isolated Cloud Interconnect parent: %s", err)
	}
	siteID := result.GetSite().GetAddCloudInterconnectSite().GetSiteID()
	if siteID == "" {
		t.Fatal("parent creation returned no ID")
	}
	t.Logf("Owned physical-connection parent: %s", siteID)
	t.Cleanup(func() { cleanupOwnedPhysicalSite(t, siteID) })
	config := func(stage int) string {
		location, bw := pop, "100"
		if stage >= 1 {
			bw = "200"
		}
		if stage >= 2 && alternate != "" {
			location = alternate
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_cloud_interconnect_physical_connection" "test" {
 site = { id = %q }
 ha_role = "PRIMARY"
 pop_location = { id = %q }
 service_provider_name = "CrossConnetMock"
 encapsulation_method = "DOT1Q"
 subnet = "10.253.250.0/30"
 private_site_ip = "10.253.250.1"
 private_cato_ip = "10.253.250.2"
 upstream_bw_limit = %q
 downstream_bw_limit = %q
}`, siteID, location, bw, bw)
	}
	object := "cato_cloud_interconnect_physical_connection.test"
	steps := []resource.TestStep{
		{Config: config(0), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttrSet(object, "vlan"), resource.TestCheckResourceAttr(object, "site.id", siteID))},
		{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true},
		{Config: config(1), Check: resource.TestCheckResourceAttr(object, "upstream_bw_limit", "200")},
		{Config: config(1), PlanOnly: true, ExpectNonEmptyPlan: false},
	}
	if alternate != "" {
		steps = append(steps, resource.TestStep{Config: config(2), Check: resource.TestCheckResourceAttr(object, "pop_location.id", alternate)}, resource.TestStep{Config: config(2), PlanOnly: true, ExpectNonEmptyPlan: false})
	}
	steps = append(steps, resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_cloud_interconnect_physical_connection" || item.Primary == nil {
				continue
			}
			id, err := ownedPrimary(ctx, client, siteID)
			if err != nil {
				return err
			}
			if id != "" {
				return fmt.Errorf("owned physical connection still exists")
			}
		}
		return nil
	}, Steps: acc.WithStableResourceID(steps, object, "id")})
}
func validateMockPOP(t *testing.T, id string) {
	t.Helper()
	result, err := acc.GetClient(t).PopLocationsPopLocationList(context.Background(), acc.CatoAccountID, &m.PopLocationFilterInput{ID: &m.IDFilterInput{In: []string{id}}})
	if err != nil {
		t.Fatal("could not verify the explicitly selected test PoP")
	}
	for _, pop := range result.GetPopLocations().GetPopLocationList().GetItems() {
		if pop == nil || pop.GetID() != id {
			continue
		}
		for _, entry := range pop.GetCloudInterconnect() {
			if entry != nil && entry.GetProviderName() == "CrossConnetMock" && entry.GetTaggingMethod() != nil && *entry.GetTaggingMethod() == m.TaggingMethodDot1q {
				return
			}
		}
	}
	t.Fatal("selected PoP does not expose the isolated CrossConnetMock DOT1Q provider")
}
func missingConnection(err error) bool {
	var api *clientv2.ErrorResponse
	if !errors.As(err, &api) || api.NetworkError != nil || api.GqlErrors == nil || len(*api.GqlErrors) == 0 {
		return false
	}
	for _, e := range *api.GqlErrors {
		if e == nil || (e.Message != "connection not found" && e.Message != "GQL-User-Error: connection not found") {
			return false
		}
	}
	return true
}
func ownedPrimary(ctx context.Context, client *cato.Client, siteID string) (string, error) {
	result, err := client.SiteCloudInterconnectPhysicalConnectionID(ctx, acc.CatoAccountID, m.CloudInterconnectPhysicalConnectionIDInput{Site: &m.SiteRefInput{By: m.ObjectRefByID, Input: siteID}, HaRole: m.HaRolePrimary})
	if missingConnection(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	id := result.GetSite().GetCloudInterconnectPhysicalConnectionID().GetID()
	if id == "" {
		return "", fmt.Errorf("owned slot lookup returned no identity")
	}
	return id, nil
}
func cleanupOwnedPhysicalSite(t *testing.T, siteID string) {
	ctx := context.Background()
	client := acc.GetClient(t)
	id, err := ownedPrimary(ctx, client, siteID)
	if err != nil {
		t.Errorf("check owned connection before cleanup: %s", err)
		return
	}
	if id != "" {
		result, err := client.SiteRemoveCloudInterconnectPhysicalConnection(ctx, m.RemoveCloudInterconnectPhysicalConnectionInput{ID: id}, acc.CatoAccountID)
		if err != nil || result.GetSite().GetRemoveCloudInterconnectPhysicalConnection().GetID() != id {
			t.Error("owned connection cleanup did not confirm identity")
			return
		}
		remaining, err := ownedPrimary(ctx, client, siteID)
		if err != nil || remaining != "" {
			t.Error("owned connection cleanup not confirmed")
			return
		}
	}
	secondary, err := client.SiteCloudInterconnectPhysicalConnectionID(ctx, acc.CatoAccountID, m.CloudInterconnectPhysicalConnectionIDInput{Site: &m.SiteRefInput{By: m.ObjectRefByID, Input: siteID}, HaRole: m.HaRoleSecondary})
	if !missingConnection(err) {
		t.Errorf("unexpected secondary slot; retaining owned parent (%s)", secondary.GetSite().GetCloudInterconnectPhysicalConnectionID().GetID())
		return
	}
	if _, err := client.SiteRemoveSite(ctx, siteID, acc.CatoAccountID); err != nil {
		t.Errorf("remove owned parent: %s", err)
		return
	}
	limit, from := int64(1), int64(0)
	lookup, err := client.EntityLookup(ctx, acc.CatoAccountID, m.EntityTypeSite, &limit, &from, nil, nil, []string{siteID}, nil, nil, nil)
	if err != nil || lookup.GetEntityLookup() == nil || len(lookup.GetEntityLookup().GetItems()) != 0 {
		t.Error("owned parent absence not confirmed")
		return
	}
	t.Logf("Owned physical-connection parent %s removed and absence verified", siteID)
}
