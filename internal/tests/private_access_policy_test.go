package tests

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPrivAccessPolicy(t *testing.T) {
	cfg := newPrivAccessPolicyCfg(t)
	resPol := "cato_private_access_policy.this"
	resRule := "cato_private_access_rule.rule_1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config:             cfg.getTfConfig(0),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resPol),
					printAttributes(resRule),
					resource.TestCheckResourceAttr(resRule, "action", "BLOCK"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: resRule,
			},
		},
		CheckDestroy: func(*terraform.State) error { publisPrivateAccessPolicy(t); return nil },
	})
}

type privAccessPolicyCfg struct {
	resName   string
	locations testLocations
	t         *testing.T
}

func newPrivAccessPolicyCfg(t *testing.T) privAccessPolicyCfg {
	return privAccessPolicyCfg{
		resName:   getRandName("private_access_policy"),
		locations: getLocations(t),
		t:         t,
	}
}

func (p privAccessPolicyCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(privAccessPolicyTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{"Name": p.resName, "Location": p.locations}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := providerCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var privAccessPolicyTFs = []string{`
	resource "cato_private_access_policy" "this" {
	  enabled = true
	}
	
	resource "cato_private_access_rule" "rule_1" {
	  action = "BLOCK"
	  active_period = {
	    effective_from     = "2026-01-20T03:04:00Z"
	    use_effective_from = true
	    expires_at         = "2027-02-20T03:04:00Z"
	    use_expires_at     = true
	  }
	  applications = [
	    { "name" = "private-app-1" }
	  ]
	  connection_origins = ["REMOTE", "SITE"]
	  countries = [
	    { "name" = "United States" }
	  ]
	  description = "rule1 description"
	  devices = [
	    { "name" = "Test device posture 1" }
	  ]
	  enabled   = true
	  name      = "rule_1"
	  platforms = ["LINUX", "WINDOWS"]
	  schedule = {
	    active_on = "WORKING_HOURS",
	    custom_timeframe = {
	      from = "2026-02-20T01:02:00Z",
	      to   = "2026-02-20T03:04:00Z"
	    },
	    custom_recurring = {
	      days = ["MONDAY", "FRIDAY"],
	      from = "08:04:00",
	      to   = "19:30:00"
	    }
	  }
	  source = {
	    users = [
	      { name = "test user_1" },
	      { name = "test user_2" }
	    ],
	    user_groups = [
	      { name = "group-1" }
	    ]
	  }
	  tracking = {
	    event = {
	      enabled = true
	    },
	    alert = {
	      enabled   = true
	      frequency = "DAILY",
	      subscription_group = [
	        { name = "subscription_group" }
	      ]
	      webhook = [
	        { name = "webhook_1" }
	      ],
	      mailing_list = [
	        { name = "mailing_list" }
	      ]
	    }
	  }
	  user_attributes = {
	    risk_score = {
	      category = "LOW"
	      operator = "LTE"
	    }
	  }
	}

	resource "cato_private_access_rule_bulk" "this" {
	  depends_on = [cato_private_access_rule.rule_1, cato_private_access_policy.this]
	  rule_data = {
	    "rule_1" = {
	      index = 0
	      name  = "rule_1"
	    }
	  }
	}
	`,
}
