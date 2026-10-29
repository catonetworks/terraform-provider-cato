//go:build acctest

package dynamicipallocationrule

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccDynamicIPAllocationRule(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 required for live fixtures")
	}
	acc.CheckCMAVars(t)()
	fixtureRange := createTestIPRange(t)
	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "FIRST_IN_POLICY"
			}
			return "LAST_IN_POLICY"
		}(), "${range}", fixtureRange).Replace(`resource "cato_dynamic_ip_allocation_rule" "test" {
at = { position = "${position}" }
rule = {
 country = []
 description = "${description}"
 device_posture_profile = []
 enabled = false
 name = "acctest-dynamic_ip_allocation_rule-${name}"
 platform = ["WINDOWS"]
 range = { global_ip_range = { id = "${range}" } }
 source = { user = [], users_group = [] }
}
}
`)
	}
	name := acc.GetRandName("dynamic_ip_allocation_rule")
	resourceName := "cato_dynamic_ip_allocation_rule.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicyDynamicIPAllocationPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.DynamicIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetDynamicIPAllocation().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_dynamic_ip_allocation_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "rule.id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-dynamic_ip_allocation_rule-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// Create only a dedicated fixture: the account-wide global-IP-ranges resource
// would take ownership of unrelated ranges in a shared development account.
func createTestIPRange(t *testing.T) string {
	t.Helper()
	client := acc.GetClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	result, err := client.ObjectCreateGlobalIPRangeBulk(ctx, acc.CatoAccountID, []*cato_models.CreateGlobalIPRangeInput{{
		Name: acc.GetRandName("dynamic_ip_pool"), IPRange: "10.254.252.0/24",
	}})
	if err != nil {
		t.Fatalf("create isolated IP range: %v", err)
	}
	ranges := result.GetObject().GetCreateGlobalIPRangeBulk().GetGlobalIPRange()
	if len(ranges) != 1 || ranges[0].GetID() == "" {
		t.Fatal("IP range creation returned no identity")
	}
	id := ranges[0].GetID()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		_, deleteErr := client.ObjectDeleteGlobalIPRangeBulk(cleanupCtx, acc.CatoAccountID, []*cato_models.GlobalIPRangeRefInput{{By: cato_models.ObjectRefByID, Input: id}})
		if deleteErr != nil {
			t.Errorf("delete isolated IP range %s: %v", id, deleteErr)
		}
	})
	return id
}
