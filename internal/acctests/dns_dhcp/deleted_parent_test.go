//go:build acctest

package dns_dhcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccDNSDHCPDeletedParent(t *testing.T) {
	if os.Getenv(envEnableDNSDHCPCRUD) != "true" {
		t.Skipf("set %s=true to run DNS/DHCP acceptance tests", envEnableDNSDHCPCRUD)
	}
	_, servers, _ := liveInputs(t)
	sdk, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, &http.Client{Timeout: time.Minute}, nil)
	require.NoError(t, err)
	for _, action := range []string{"refresh", "destroy"} {
		t.Run(action, func(t *testing.T) {
			ctx := context.Background()
			created, err := sdk.SiteAddIpsecIkeV2Site(ctx, models.AddIpsecIkeV2SiteInput{
				Name: fmt.Sprintf("acctest-dns-parent-%d", time.Now().UnixNano()), NativeNetworkRange: acc.GetRandNetworkPrefix() + ".0/24", SiteType: models.SiteTypeCloudDc,
				SiteLocation: &models.AddSiteLocationInput{CountryCode: "FR", Timezone: "Europe/Paris"},
			}, acc.CatoAccountID)
			require.NoError(t, err)
			require.NotNil(t, created.GetSite().GetAddIpsecIkeV2Site())
			siteID := created.Site.AddIpsecIkeV2Site.SiteID
			require.NotEmpty(t, siteID)
			deleted := false
			deleteParent := func() {
				_, err := sdk.SiteRemoveSite(ctx, siteID, acc.CatoAccountID)
				require.NoError(t, err)
				deleted = true
				found, err := sdk.EntityLookup(ctx, acc.CatoAccountID, models.EntityTypeSite, nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
				require.NoError(t, err)
				require.Empty(t, found.EntityLookup.Items)
				require.NotNil(t, found.EntityLookup.Total)
				require.Zero(t, *found.EntityLookup.Total)
			}
			t.Cleanup(func() {
				if !deleted {
					deleteParent()
				}
			})
			config := fmt.Sprintf(`
provider "cato" {
 account_id = %q
 retry_max = 0
 version_check_disabled = true
}
resource "cato_dns_site_settings" "parent" {
 site = { id = %q }
 primary_server = %q
}
`, acc.CatoAccountID, siteID, servers[0])
			removed := func(state *terraform.State) error {
				if _, ok := state.RootModule().Resources["cato_dns_site_settings.parent"]; ok {
					return fmt.Errorf("deleted-parent override remains in state")
				}
				return nil
			}
			step := resource.TestStep{Config: config, PreConfig: deleteParent, Destroy: action == "destroy"}
			if action == "refresh" {
				step.Config = ""
				step.RefreshState = true
				step.ExpectNonEmptyPlan = true
				step.Check = removed
			}
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
				Steps:                    []resource.TestStep{{Config: config}, step},
				CheckDestroy: func(_ *terraform.State) error {
					result, err := sdk.EntityLookup(ctx, acc.CatoAccountID, models.EntityTypeSite, nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
					if err != nil {
						return err
					}
					if len(result.EntityLookup.Items) != 0 {
						return fmt.Errorf("disposable parent site remains")
					}
					return nil
				},
			})
		})
	}
}
