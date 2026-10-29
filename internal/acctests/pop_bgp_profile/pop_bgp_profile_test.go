//go:build acctest

package popbgpprofile

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccPopBGPProfile(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("pop_bgp_profile")
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(name))
	const communityRange = 1 << 16
	seed := hash.Sum32() % communityRange
	config := func(updated, clearDescription bool) string {
		suffix, from := "", 65001
		if updated {
			suffix = "_updated"
			from = 65002
		}
		optional := fmt.Sprintf("description = %q", name+suffix)
		if clearDescription {
			optional = ""
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_pop_bgp_profile" "test" {
 name = %q
 %s
 communities = [{from = %q,to = %q}]
}`, name+suffix, optional, fmt.Sprint(from), fmt.Sprint(seed))
	}
	object := "cato_pop_bgp_profile.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, item := range state.RootModule().Resources {
				if item.Type != "cato_pop_bgp_profile" || item.Primary == nil {
					continue
				}
				result, err := acc.GetClient(t).PopLocationsBgpProfile(context.Background(), acc.CatoAccountID, item.Primary.ID)
				if err != nil {
					if missingFixture(err) {
						continue
					}
					return err
				}
				if result == nil || result.GetPopLocations() == nil {
					return fmt.Errorf("invalid BGP profile response")
				}
				if result.GetPopLocations().GetBgpProfile() != nil {
					return fmt.Errorf("owned BGP profile still exists")
				}
			}
			return nil
		}, Steps: []resource.TestStep{
			{Config: config(false, false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "communities.#", "1"))},
			{Config: config(false, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: object, ImportState: true, ImportStateVerify: true},
			{Config: config(true, false), Check: resource.TestCheckResourceAttr(object, "name", name+"_updated")},
			{Config: config(true, false), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(true, true), Check: resource.TestCheckResourceAttr(object, "description", "")},
			{Config: config(true, true), PlanOnly: true, ExpectNonEmptyPlan: false},
		}})
}
func missingFixture(err error) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item == nil || (item.Message != "BGP profile not found" && item.Message != "GQL-User-Error: BGP profile not found") {
			return false
		}
	}
	return true
}
