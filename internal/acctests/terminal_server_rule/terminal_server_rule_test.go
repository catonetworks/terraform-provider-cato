//go:build acctest

package terminalserverrule

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

func TestAccTerminalServerRule(t *testing.T) {
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
		}(), "${range}", fixtureRange).Replace(`resource "cato_terminal_server_rule" "test" {
at = { position = "${position}" }
rule = {"allowed_host_ip": {"id": "${range}"}, "description": "${description}", "enabled": false, "exclude_traffic": [], "name": "acctest-terminal_server_rule-${name}"}
}
`)
	}
	name := acc.GetRandName("terminal_server_rule")
	resourceName := "cato_terminal_server_rule.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicyTerminalServerPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.TerminalServerPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetTerminalServer().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_terminal_server_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
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
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-terminal_server_rule-"+name+"-updated")},
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
		Name: acc.GetRandName("terminal_server_ip"), IPRange: "192.0.2.241",
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
