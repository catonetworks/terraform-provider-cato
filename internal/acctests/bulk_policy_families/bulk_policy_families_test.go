//go:build acctest

package bulk_policy_families

import (
	"context"
	"encoding/json"
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

type policyFamily struct {
	prefix, mover string
	rule          func(*testing.T) string
}

// TestAccBulkPolicyFamilies checks create, refresh, import, update, ordering and
// destroy using disabled test rules. Existing account policies are not toggled.
func TestAccBulkPolicyFamilies(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("set TF_ACC=1 to run live policy tests")
	}
	acc.SkipByEnv(t)
	families := []policyFamily{
		{"anti_malware_file_hash", "amfh", func(_ *testing.T) string {
			return fmt.Sprintf(`action="BLOCK", expiration_date=%q, file_name="validation.bin", sha256=%q`, time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339), strings.Repeat("a", 64))
		}},
		{"client_connectivity", "cc", func(*testing.T) string { return `action="ALLOW", confidence_level="ANY"` }},
		{"split_tunnel", "st", func(*testing.T) string {
			return `action="EXCLUDE", routing_priority="TUNNEL", coverage="ALL", platform=["WINDOWS"], destination={domain=["validation.invalid"]}`
		}},
		{"dynamic_ip_allocation", "dia", func(t *testing.T) string {
			ranges := acc.GetGlobalIPRanges(t)
			return fmt.Sprintf(`range={global_ip_range={id=%q}}`, ranges[0].ID)
		}},
		{"terminal_server", "ts", func(t *testing.T) string {
			return fmt.Sprintf(`allowed_host_ip={id=%q}`, globalIPRangeID(t))
		}},
		{"site_web_proxy", "swp", func(_ *testing.T) string {
			return `fqdn="validation.invalid", port="8443", allowed_destination_protocol="HTTPS", should_associate_with_all_sites=false, authentication_config={method="NONE"}`
		}},
		{"socket_bypass", "sb", func(_ *testing.T) string {
			return `action={port="AUTOMATIC",event_enabled=false}, source={ip=["198.18.0.230"]}, destination={ip=["192.0.2.230"]}`
		}},
		{"sdp_static_ip_allocation", "ssia", func(t *testing.T) string {
			ip := os.Getenv("TFACC_SDP_STATIC_IP")
			if ip == "" {
				t.Skip("set TFACC_SDP_STATIC_IP to an unused IP in the account static range")
			}
			return fmt.Sprintf(`ip=%q, source={user={id=%q}}`, ip, lookupID(t, cato_models.EntityTypeVpnUser))
		}},
		{"remote_port_fwd", "rpf", func(t *testing.T) string {
			return fmt.Sprintf(`external_ip={id=%q}, external_port_range={from="49151", to="49151"}, internal_ip="198.18.0.231", internal_port_range={from="49151", to="49151"}, restriction_type="ALLOW_LIST", remote_ips={subnet=["192.0.2.0/24"]}`, lookupID(t, cato_models.EntityTypeAllocatedIP))
		}},
	}
	for _, family := range families {
		t.Run(family.prefix, func(t *testing.T) {
			requirePolicyReadAccess(t, family.prefix)
			name := acc.GetRandName("bulk_family")
			extra := family.rule(t)
			ruleAddress := "cato_" + family.prefix + "_rule.test"
			sectionAddress := "cato_" + family.prefix + "_section.test"
			config := func(description string) string {
				return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_%[1]s_section" "test" {
 at={position="LAST_IN_POLICY"}
 section={name=%[2]q}
}
resource "cato_%[1]s_section" "anchor" {
 depends_on=[cato_%[1]s_section.test]
 at={position="BEFORE_SECTION",ref=cato_%[1]s_section.test.section.id}
 section={name="%[7]s-anchor"}
}
resource "cato_%[1]s_rule" "test" {
 at={position="LAST_IN_SECTION", ref=cato_%[1]s_section.test.section.id}
 rule={name=%[3]q, description=%[4]q, enabled=false, %[5]s}
}
resource "cato_bulk_%[6]s_move_rule" "test" {
 depends_on=[cato_%[1]s_rule.test,cato_%[1]s_section.anchor]
 section_data={%[2]q={section_index=1}, "%[7]s-anchor"={section_index=2}}
 rule_data={%[3]q={section_name=%[2]q, index_in_section=1}}
}
`, family.prefix, name+"-section", name+"-rule", description, extra, family.mover, name)
			}
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
				PreCheck:                 acc.CheckCMAVars(t),
				Steps: []resource.TestStep{
					{Config: config("initial"), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet(ruleAddress, "rule.id"),
						resource.TestCheckResourceAttr(ruleAddress, "rule.enabled", "false"),
						resource.TestCheckResourceAttrSet(sectionAddress, "section.id"),
						checkPublishedSectionOrder(t, family.prefix, name+"-section", name+"-anchor"),
					)},
					{ResourceName: ruleAddress, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
					{ResourceName: sectionAddress, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
					{Config: config("updated"), Check: resource.TestCheckResourceAttr(ruleAddress, "rule.description", "updated")},
				},
			})
		})
	}
}

func lookupID(t *testing.T, kind cato_models.EntityType) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := acc.GetClient(t).EntityLookup(ctx, acc.CatoAccountID, kind, new(int64(1)), nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("lookup %s: %v", kind, err)
	}
	items := response.GetEntityLookup().GetItems()
	if len(items) == 0 {
		t.Skipf("account needs an existing %s fixture for this family", kind)
	}
	return items[0].GetEntity().GetID()
}

// Read the published policy independently of Terraform state. Two owned sections
// ensure a successful mover actually changes their relative order.
func checkPublishedSectionOrder(t *testing.T, family, first, second string) resource.TestCheckFunc {
	t.Helper()
	return func(_ *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		response, err := readPublishedPolicy(ctx, t, family)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(response)
		if err != nil {
			return err
		}
		var snapshot struct {
			Policy map[string]struct {
				Policy struct {
					Sections []struct{ Section struct{ Name string } }
				}
			}
		}
		if err := json.Unmarshal(raw, &snapshot); err != nil {
			return err
		}
		for _, policy := range snapshot.Policy {
			firstIndex, secondIndex := -1, -1
			for i, row := range policy.Policy.Sections {
				if row.Section.Name == first {
					firstIndex = i
				}
				if row.Section.Name == second {
					secondIndex = i
				}
			}
			if firstIndex >= 0 && secondIndex > firstIndex {
				return nil
			}
		}
		return fmt.Errorf("published %s policy does not place %q before %q", family, first, second)
	}
}

func readPublishedPolicy(ctx context.Context, t *testing.T, family string) (any, error) {
	t.Helper()
	client := acc.GetClient(t)
	published := &cato_models.PolicyRevisionInput{Type: new(cato_models.PolicyRevisionTypePublic)}
	var response any
	var err error
	switch family {
	case "anti_malware_file_hash":
		response, err = client.PolicyAntiMalwareFileHashPolicy(ctx, acc.CatoAccountID, &cato_models.AntiMalwareFileHashPolicyInput{Revision: published})
	case "client_connectivity":
		response, err = client.PolicyClientConnectivityPolicy(ctx, acc.CatoAccountID, &cato_models.ClientConnectivityPolicyInput{Revision: published})
	case "split_tunnel":
		response, err = client.PolicySplitTunnelPolicy(ctx, acc.CatoAccountID, &cato_models.SplitTunnelPolicyInput{Revision: published})
	case "dynamic_ip_allocation":
		response, err = client.PolicyDynamicIPAllocationPolicy(ctx, acc.CatoAccountID, &cato_models.DynamicIPAllocationPolicyInput{Revision: published})
	case "sdp_static_ip_allocation":
		response, err = client.PolicySdpStaticIPAllocationPolicy(ctx, acc.CatoAccountID, &cato_models.SdpStaticIPAllocationPolicyInput{Revision: published})
	case "terminal_server":
		response, err = client.PolicyTerminalServerPolicy(ctx, acc.CatoAccountID, &cato_models.TerminalServerPolicyInput{Revision: published})
	case "site_web_proxy":
		response, err = client.PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID, &cato_models.SiteWebProxyPolicyInput{Revision: published})
	case "socket_bypass":
		response, err = client.PolicySocketBypassPolicy(ctx, acc.CatoAccountID, &cato_models.SocketBypassPolicyInput{Revision: published})
	case "remote_port_fwd":
		response, err = client.RemotePortFwdPolicy(ctx, acc.CatoAccountID, &cato_models.RemotePortFwdPolicyInput{Revision: published})
	default:
		return nil, fmt.Errorf("no published-policy reader for %s", family)
	}
	return response, err
}

// Fixture discovery must never create or delete shared account objects.
func globalIPRangeID(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := acc.GetClient(t).ObjectGlobalIPRangeList(ctx, acc.CatoAccountID, nil)
	if err != nil {
		t.Fatalf("list existing global IP ranges: %v", err)
	}
	items := response.GetObject().GetGlobalIPRangeList().GetItems()
	if len(items) == 0 {
		t.Skip("account needs an existing global IP range fixture")
	}
	return items[0].GetID()
}

// Some policy families require account entitlements or API-key permissions.
// Check access before creating any fixtures; all other API failures stay fatal.
func requirePolicyReadAccess(t *testing.T, family string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := readPublishedPolicy(ctx, t, family); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "insufficient privileges") {
			t.Skipf("test account lacks read access to %s; use a config with this policy entitlement", family)
		}
		t.Fatalf("read %s policy prerequisite: %v", family, err)
	}
}
