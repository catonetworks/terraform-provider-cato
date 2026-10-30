//go:build acctest

package always_on_policy

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
	"text/template"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAlwaysOnPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CheckCMAVars(t)()
	t.Setenv("DISABLE_POLICY_RULE_CLEANUP", "true")
	mockSrv := accmock.NewMockServer(t, t.Name())
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newAlwaysOnPolicyCfg(t)
	defer cfg.restoreOriginal()
	res := "cato_always_on_policy.this"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Initial apply changes and publishes the existing policy's toggle.
				Config: cfg.getTfConfig(!cfg.originalEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", strconv.FormatBool(!cfg.originalEnabled)),
				),
			},
			{
				// Update restores the original toggle before Terraform forgets the resource.
				Config: cfg.getTfConfig(cfg.originalEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", strconv.FormatBool(cfg.originalEnabled)),
				),
			},
			{
				// Import uses the configured account ID without an id attribute.
				ResourceName:                         res,
				ImportState:                          true,
				ImportStateId:                        acc.CatoAccountID,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "enabled",
			},
		},
	})
}

type alwaysOnPolicyCfg struct {
	originalEnabled bool
	t               *testing.T
}

func newAlwaysOnPolicyCfg(t *testing.T) alwaysOnPolicyCfg {
	t.Helper()
	enabled, err := readEnabled(t)
	if err != nil {
		t.Fatal(err)
	}
	return alwaysOnPolicyCfg{originalEnabled: enabled, t: t}
}

func readEnabled(t *testing.T) (bool, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := acc.GetClient(t).PolicyZtnaAlwaysOnPolicy(ctx, acc.CatoAccountID, nil)
	if err != nil {
		return false, err
	}
	policy := result.GetPolicy().GetZtnaAlwaysOn()
	if policy == nil || policy.GetPolicy().GetRules() == nil || policy.GetPolicy().GetSections() == nil {
		return false, fmt.Errorf("API did not return a complete Always On policy")
	}
	return policy.GetPolicy().GetEnabled(), nil
}

func (cfg alwaysOnPolicyCfg) restoreOriginal() {
	cfg.t.Helper()
	enabled, err := readEnabled(cfg.t)
	if err != nil {
		cfg.t.Errorf("read Always On policy during cleanup: %v", err)
	}
	if err == nil && enabled == cfg.originalEnabled {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	state := cato_models.PolicyToggleStateDisabled
	if cfg.originalEnabled {
		state = cato_models.PolicyToggleStateEnabled
	}
	result, err := acc.GetClient(cfg.t).PolicyZtnaAlwaysOnUpdatePolicy(ctx, acc.CatoAccountID, nil,
		cato_models.ZtnaAlwaysOnPolicyUpdateInput{State: &state})
	if err != nil {
		cfg.t.Errorf("restore original Always On policy toggle: %v", err)
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetUpdatePolicy()
	if payload.GetStatus() == nil || *payload.GetStatus() != cato_models.PolicyMutationStatusSuccess || len(payload.GetErrors()) > 0 {
		cfg.t.Error("API did not successfully restore the original Always On policy toggle")
		return
	}
	published, err := acc.GetClient(cfg.t).PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, acc.CatoAccountID, nil, nil)
	if err != nil {
		cfg.t.Errorf("publish restored Always On policy toggle: %v", err)
		return
	}
	publishPayload := published.GetPolicy().GetZtnaAlwaysOn().GetPublishPolicyRevision()
	if publishPayload.GetStatus() == nil || *publishPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess || len(publishPayload.GetErrors()) > 0 {
		cfg.t.Error("API did not successfully publish the restored Always On policy toggle")
		return
	}
	enabled, err = readEnabled(cfg.t)
	if err != nil || enabled != cfg.originalEnabled {
		cfg.t.Errorf("original Always On policy toggle was not restored: enabled=%t, error=%v", enabled, err)
	}
}

func (cfg alwaysOnPolicyCfg) getTfConfig(enabled bool) string {
	cfg.t.Helper()
	tmpl, err := template.New("always_on_policy").Parse(alwaysOnPolicyTF)
	if err != nil {
		cfg.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]bool{"Enabled": enabled}); err != nil {
		cfg.t.Fatal(err)
	}
	return acc.ProviderCfg() + buf.String()
}

const alwaysOnPolicyTF = `resource "cato_always_on_policy" "this" {
  enabled = {{.Enabled}}
}`
