//go:build acctest

package sdpstaticipallocationrule

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSdpStaticIPAllocationRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	acc.CheckCMAVars(t)()
	staticIP, anchorIP := staticIPs(t)
	anchorUser := createTestUser(t)
	fixtureUser := createTestUser(t)
	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "FIRST_IN_POLICY"
			}
			return "LAST_IN_POLICY"
		}(), "${user}", fixtureUser, "${ip}", staticIP).Replace(`resource "cato_sdp_static_ip_allocation_rule" "test" {
at = { position = "${position}" }
rule = {
 description = "${description}"
 enabled = false
 ip = "${ip}"
 name = "acctest-sdp_static_ip_allocation_rule-${name}"
 source = { user = { id = "${user}" } }
}
}
`)
	}
	name := acc.GetRandName("sdp_static_ip_allocation_rule")
	singleRuleConfig := config
	anchorName := name + "-anchor"
	config = func(name, description string) string {
		anchor := strings.Replace(singleRuleConfig(anchorName, "initial"), `"test" {`, `"anchor" {`, 1)
		anchor = strings.NewReplacer(fmt.Sprintf("%q", fixtureUser), fmt.Sprintf("%q", anchorUser), fmt.Sprintf("%q", staticIP), fmt.Sprintf("%q", anchorIP)).Replace(anchor)
		target := strings.Replace(singleRuleConfig(name, description), `"test" {`, `"test" {
 depends_on = [cato_sdp_static_ip_allocation_rule.anchor]`, 1)
		anchor = strings.TrimPrefix(anchor, acc.ProviderCfg())
		return anchor + "\n" + target
	}
	resourceName := "cato_sdp_static_ip_allocation_rule.test"

	published := func(state *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		revisionType := cato_models.PolicyRevisionTypePublic
		policy, err := acc.GetClient(t).PolicySdpStaticIPAllocationPolicy(ctx, acc.CatoAccountID,
			&cato_models.SdpStaticIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
		if err != nil {
			return err
		}

		ids := []string{}
		for _, item := range policy.GetPolicy().GetSdpStaticIPAllocation().GetPolicy().GetRules() {
			ids = append(ids, item.GetRule().GetID())
		}
		return acc.CheckPolicyPlacement(state, resourceName, "rule.id", ids)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicySdpStaticIPAllocationPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.SdpStaticIPAllocationPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetSdpStaticIPAllocation().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_sdp_static_ip_allocation_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "rule.id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-sdp_static_ip_allocation_rule-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
		}, resourceName,

			// A dedicated disabled local user avoids altering existing user assignments.
			"rule.id", published),
	})
}

func createTestUser(t *testing.T) string {
	t.Helper()
	client := acc.GetClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	result, err := client.UserCreateUser(ctx, acc.CatoAccountID, cato_models.CreateUserInput{
		Email: acc.GetRandName("sdp_static_ip") + "@example.invalid", FirstName: "Terraform", LastName: "Acceptance",
	})
	if err != nil {
		t.Fatalf("create isolated user: %v", err)
	}
	id := result.GetUser().GetCreateUser().GetUser().GetID()
	if id == "" {
		t.Fatal("user creation returned no identity")
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		_, deleteErr := client.UserDeleteUser(cleanupCtx, acc.CatoAccountID, cato_models.DeleteUserInput{UserID: []scalars.Long{scalars.Long(id)}})
		if deleteErr != nil {
			t.Errorf("delete isolated user %s: %v", id, deleteErr)
		}
	})
	_, err = client.UserDisableUser(ctx, acc.CatoAccountID, cato_models.DisableUserInput{UserID: []scalars.Long{scalars.Long(id)}})
	if err != nil {
		t.Fatalf("disable isolated user: %v", err)
	}
	return id
}

func staticIPs(t *testing.T) (primary, anchor string) {
	t.Helper()
	primary, anchor = os.Getenv("TFACC_SDP_STATIC_IP"), os.Getenv("TFACC_SDP_STATIC_IP_ANCHOR")
	if primary == "" || anchor == "" || primary == anchor {
		t.Skip("Two distinct unused static IPs are required: TFACC_SDP_STATIC_IP and TFACC_SDP_STATIC_IP_ANCHOR")
	}
	return primary, anchor
}
