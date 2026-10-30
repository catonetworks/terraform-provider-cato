//go:build acctest

package swprule

import (
	"context"
	"fmt"
	"strings"
	"testing"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSiteWebProxyRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CheckCMAVars(t)()
	ctx := context.Background()
	client := acc.GetClient(t)
	public := m.PolicyRevisionTypePublic
	probe, err := client.PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID, &m.SiteWebProxyPolicyInput{Revision: &m.PolicyRevisionInput{Type: &public}})
	if err != nil || probe.GetPolicy().GetSiteWebProxy().GetPolicy() == nil {
		t.Fatal("Site Web Proxy policy API access is required")
	}
	parent, err := client.SiteAddSocketSite(ctx, m.AddSocketSiteInput{ConnectionType: m.SiteConnectionTypeEnumSocketX1600, Name: acc.GetRandName("swp_site"), NativeNetworkRange: acc.GetRandNetworkPrefix() + ".0/24", SiteType: m.SiteTypeBranch, SiteLocation: &m.AddSiteLocationInput{CountryCode: "GB", Timezone: "Europe/London"}}, acc.CatoAccountID)
	if err != nil {
		t.Fatalf("create isolated proxy parent: %s", err)
	}
	siteID := parent.GetSite().GetAddSocketSite().GetSiteID()
	if siteID == "" {
		t.Fatal("parent returned no ID")
	}
	t.Logf("Owned proxy test parent: %s", siteID)
	t.Cleanup(func() { cleanupProxyParent(t, siteID) })

	name := acc.GetRandName("swp")
	domain := strings.ReplaceAll(name, "_", "-") + ".example.invalid"
	rule := func(label, description, fqdn, port, protocol, at, kerberos string) string {
		if label == "test" {
			at = "depends_on=[cato_site_web_proxy_rule.anchor]\n" + at
		}
		return fmt.Sprintf(`resource "cato_site_web_proxy_rule" %q {
 %s
 rule={
 name=%q
 description=%q
 enabled=false
 fqdn=%q
 port=%q
 allowed_destination_protocol=%q
 should_associate_with_all_sites=false
 associated_site=[{id=%q}]
 authentication_config={
method="NONE"
%s
}
 }
}`, label, at, name+"-"+label, description, fqdn, port, protocol, siteID, kerberos)
	}
	config := func(stage int) string {
		at := `at={position="LAST_IN_POLICY"}`
		description, port, protocol, kerberos := "initial", "3128", "BOTH", ""
		if stage >= 1 {
			at = `at={position="BEFORE_RULE",ref=cato_site_web_proxy_rule.anchor.rule.id}`
			description, port, protocol = "updated", "3129", "HTTP"
		}
		if stage >= 2 {
			at = ""
			protocol = "HTTPS"
			kerberos = ""
		}
		return acc.ProviderCfg() + "\n" + rule("anchor", "anchor", "anchor-"+domain, "3128", "BOTH", `at={position="LAST_IN_POLICY"}`, "") + "\n" + rule("test", description, domain, port, protocol, at, kerberos)
	}
	object := "cato_site_web_proxy_rule.test"
	steps := []resource.TestStep{}
	for i := 0; i < 3; i++ {
		steps = append(steps, resource.TestStep{Config: config(i), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "rule.id"), resource.TestCheckResourceAttr(object, "rule.enabled", "false"))}, resource.TestStep{Config: config(i), PlanOnly: true, ExpectNonEmptyPlan: false}, resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
			item := state.RootModule().Resources[object]
			if item == nil || item.Primary == nil {
				return "", fmt.Errorf("missing test state")
			}
			return item.Primary.Attributes["rule.id"], nil
		}, ImportStateVerifyIgnore: []string{"at"}})
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: steps, CheckDestroy: func(state *terraform.State) error { return checkProxyDestroy(t, state) }})
}

func cleanupProxyParent(t *testing.T, siteID string) {
	t.Helper()
	ctx := context.Background()
	client := acc.GetClient(t)
	public := m.PolicyRevisionTypePublic

	policy, readErr := client.PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID, &m.SiteWebProxyPolicyInput{Revision: &m.PolicyRevisionInput{Type: &public}})
	if readErr != nil || policy.GetPolicy().GetSiteWebProxy().GetPolicy() == nil {
		t.Error("Cannot verify proxy fixture references; retaining parent")
		return
	}
	for _, entry := range policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() {
		for _, site := range entry.GetRule().GetAssociatedSite() {
			if site.GetID() == siteID {
				t.Error("Proxy fixture still referenced; retaining parent")
				return
			}
		}
	}
	if _, removeErr := client.SiteRemoveSite(ctx, siteID, acc.CatoAccountID); removeErr != nil {
		t.Errorf("remove owned proxy parent: %s", removeErr)
		return
	}
	limit, from := int64(1), int64(0)
	lookup, lookupErr := client.EntityLookup(ctx, acc.CatoAccountID, m.EntityTypeSite, &limit, &from, nil, nil, []string{siteID}, nil, nil, nil)
	if lookupErr != nil || lookup.GetEntityLookup() == nil || len(lookup.GetEntityLookup().GetItems()) != 0 {
		t.Error("Owned parent deletion unconfirmed")
		return
	}
	t.Log("Owned proxy parent removed and absence verified")
}

func checkProxyDestroy(t *testing.T, state *terraform.State) error {
	ctx := context.Background()
	client := acc.GetClient(t)
	public := m.PolicyRevisionTypePublic

	policy, readErr := client.PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID, &m.SiteWebProxyPolicyInput{Revision: &m.PolicyRevisionInput{Type: &public}})
	if readErr != nil {
		return readErr
	}
	if policy.GetPolicy().GetSiteWebProxy().GetPolicy() == nil || policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() == nil {
		return fmt.Errorf("complete public policy unavailable")
	}
	for _, entry := range policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() {
		for _, item := range state.RootModule().Resources {
			if item.Type == "cato_site_web_proxy_rule" && item.Primary != nil && entry.GetRule().GetID() == item.Primary.Attributes["rule.id"] {
				return fmt.Errorf("owned proxy rule still exists")
			}
		}
	}
	return nil
}
