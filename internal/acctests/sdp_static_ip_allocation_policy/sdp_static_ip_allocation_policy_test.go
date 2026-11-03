//go:build acctest

package sdp_static_ip_allocation_policy

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

// TestAccSdpStaticIpAllocationPolicy verifies toggling, import, and restore of the
// singleton SdpStaticIpAllocation account-level policy toggle. The policy always exists;
// the test flips it and restores the original value on exit.
func TestAccSdpStaticIpAllocationPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CheckCMAVars(t)()
	mockSrv := accmock.NewMockServer(t, t.Name())
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newPolicyCfg(t)
	defer cfg.restoreOriginal()
	res := "cato_sdp_static_ip_allocation_policy.this"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.tfConfig(!cfg.originalEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", strconv.FormatBool(!cfg.originalEnabled)),
				),
			},
			{
				Config: cfg.tfConfig(cfg.originalEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", strconv.FormatBool(cfg.originalEnabled)),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

type policyCfg struct {
	originalEnabled bool
	t               *testing.T
}

func newPolicyCfg(t *testing.T) policyCfg {
	t.Helper()
	enabled, err := readEnabled(t)
	if err != nil {
		t.Fatal(err)
	}
	return policyCfg{originalEnabled: enabled, t: t}
}

func readEnabled(t *testing.T) (bool, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := acc.GetClient(t).PolicySdpStaticIPAllocationPolicy(ctx, acc.CatoAccountID, nil)
	if err != nil {
		return false, err
	}
	policy := result.GetPolicy().GetSdpStaticIPAllocation()
	if policy == nil {
		return false, fmt.Errorf("API did not return the SdpStaticIpAllocation policy")
	}
	return policy.GetPolicy().GetEnabled(), nil
}

func (cfg policyCfg) restoreOriginal() {
	cfg.t.Helper()
	enabled, err := readEnabled(cfg.t)
	if err != nil {
		cfg.t.Errorf("read SdpStaticIpAllocation policy during cleanup: %v", err)
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
	if _, err := acc.GetClient(cfg.t).PolicySdpStaticIPAllocationUpdatePolicy(ctx, acc.CatoAccountID, nil, cato_models.SdpStaticIPAllocationPolicyUpdateInput{State: &state}); err != nil {
		cfg.t.Errorf("restore original SdpStaticIpAllocation policy toggle: %v", err)
		return
	}
	if _, err := acc.GetClient(cfg.t).PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, acc.CatoAccountID, nil, nil); err != nil {
		cfg.t.Errorf("publish restored SdpStaticIpAllocation policy toggle: %v", err)
	}
}

func (cfg policyCfg) tfConfig(enabled bool) string {
	cfg.t.Helper()
	tmpl, err := template.New("sdp_static_ip_allocation_policy").Parse(policyTF)
	if err != nil {
		cfg.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]bool{"Enabled": enabled}); err != nil {
		cfg.t.Fatal(err)
	}
	return acc.ProviderCfg() + buf.String()
}

const policyTF = `resource "cato_sdp_static_ip_allocation_policy" "this" {
  enabled = {{.Enabled}}
}`
