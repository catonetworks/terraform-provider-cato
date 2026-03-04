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
	// t.Skip("Skipping this test for now")

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
					resource.TestCheckResourceAttrSet(resRule, "id"),
					resource.TestCheckResourceAttr(resRule, "action", "BLOCK"),
					resource.TestCheckResourceAttr(resRule, "active_period.%", "4"),
					resource.TestCheckResourceAttr(resRule, "active_period.effective_from", "2026-01-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "active_period.expires_at", "2027-02-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "active_period.use_effective_from", "true"),
					resource.TestCheckResourceAttr(resRule, "active_period.use_expires_at", "true"),
					resource.TestCheckResourceAttr(resRule, "applications.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "applications.*",
						map[string]string{"name": cfg.applications[0].Name, "id": cfg.applications[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "applications.*",
						map[string]string{"name": cfg.applications[1].Name, "id": cfg.applications[1].ID},
					),
					resource.TestCheckResourceAttr(resRule, "connection_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resRule, "connection_origins.*", "REMOTE"),
					resource.TestCheckTypeSetElemAttr(resRule, "connection_origins.*", "SITE"),
					resource.TestCheckResourceAttr(resRule, "countries.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "countries.*",
						map[string]string{"name": "United States", "id": "US"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "countries.*",
						map[string]string{"name": "Czechia", "id": "CZ"},
					),
					resource.TestCheckResourceAttr(resRule, "description", "rule1 description"),
					resource.TestCheckResourceAttr(resRule, "devices.#", "1"),
					resource.TestCheckResourceAttr(resRule, "devices.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "devices.0.id"),
					resource.TestCheckResourceAttr(resRule, "devices.0.name", "Test device posture 1"),
					resource.TestCheckResourceAttr(resRule, "enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "name", cfg.resName),
					resource.TestCheckResourceAttr(resRule, "platforms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resRule, "platforms.*", "LINUX"),
					resource.TestCheckTypeSetElemAttr(resRule, "platforms.*", "WINDOWS"),
					resource.TestCheckResourceAttr(resRule, "schedule.%", "3"),
					resource.TestCheckResourceAttr(resRule, "schedule.active_on", "WORKING_HOURS"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.days.#", "2"),
					resource.TestCheckTypeSetElemAttr(resRule, "schedule.custom_recurring.days.*", "MONDAY"),
					resource.TestCheckTypeSetElemAttr(resRule, "schedule.custom_recurring.days.*", "FRIDAY"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.%", "3"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.from", "08:04:00"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.to", "19:30:00"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.%", "2"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.from", "2026-02-20T01:02:00Z"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.to", "2026-02-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "source.%", "2"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.#", "1"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "source.user_groups.0.id"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.0.name", cfg.userGroups[0].Name),
					resource.TestCheckResourceAttr(resRule, "source.users.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "source.users.*",
						map[string]string{"name": cfg.users[0].Name, "id": cfg.users[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "source.users.*",
						map[string]string{"name": cfg.users[1].Name, "id": cfg.users[1].ID},
					),
					resource.TestCheckResourceAttr(resRule, "tracking.%", "2"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.mailing_list.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.0.name", cfg.mailingLists[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.subscription_group.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.0.name", cfg.subscriptionGroups[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.webhook.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.0.name", cfg.webhooks[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.event.%", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.event.enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.%", "1"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.%", "2"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.category", "LOW"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.operator", "LTE"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: resRule,
			},
			{
				// Update the resource
				Config:             cfg.getTfConfig(1),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resPol),
					printAttributes(resRule),
					resource.TestCheckResourceAttrSet(resRule, "id"),
					resource.TestCheckResourceAttr(resRule, "action", "ALLOW"),
					resource.TestCheckResourceAttr(resRule, "active_period.%", "4"),
					resource.TestCheckResourceAttr(resRule, "active_period.effective_from", "2026-02-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "active_period.expires_at", "2027-03-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "active_period.use_effective_from", "true"),
					resource.TestCheckResourceAttr(resRule, "active_period.use_expires_at", "true"),
					resource.TestCheckResourceAttr(resRule, "applications.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "applications.*",
						map[string]string{"name": cfg.applications[1].Name, "id": cfg.applications[1].ID},
					),
					resource.TestCheckResourceAttr(resRule, "connection_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resRule, "connection_origins.*", "REMOTE"),
					resource.TestCheckTypeSetElemAttr(resRule, "connection_origins.*", "SITE"),
					resource.TestCheckResourceAttr(resRule, "countries.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "countries.*",
						map[string]string{"name": "United States", "id": "US"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "countries.*",
						map[string]string{"name": "Italy", "id": "IT"},
					),
					resource.TestCheckResourceAttr(resRule, "description", "rule1 description"),
					resource.TestCheckResourceAttr(resRule, "devices.#", "1"),
					resource.TestCheckResourceAttr(resRule, "devices.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "devices.0.id"),
					resource.TestCheckResourceAttr(resRule, "devices.0.name", "Test device posture 1"),
					resource.TestCheckResourceAttr(resRule, "enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "name", cfg.resName),
					resource.TestCheckResourceAttr(resRule, "platforms.#", "1"),
					resource.TestCheckTypeSetElemAttr(resRule, "platforms.*", "LINUX"),
					resource.TestCheckResourceAttr(resRule, "schedule.%", "3"),
					resource.TestCheckResourceAttr(resRule, "schedule.active_on", "WORKING_HOURS"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.days.#", "3"),
					resource.TestCheckTypeSetElemAttr(resRule, "schedule.custom_recurring.days.*", "MONDAY"),
					resource.TestCheckTypeSetElemAttr(resRule, "schedule.custom_recurring.days.*", "TUESDAY"),
					resource.TestCheckTypeSetElemAttr(resRule, "schedule.custom_recurring.days.*", "FRIDAY"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.%", "3"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.from", "08:05:00"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_recurring.to", "19:31:00"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.%", "2"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.from", "2026-01-20T01:02:00Z"),
					resource.TestCheckResourceAttr(resRule, "schedule.custom_timeframe.to", "2026-04-20T03:04:00Z"),
					resource.TestCheckResourceAttr(resRule, "source.%", "2"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.#", "1"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "source.user_groups.0.id"),
					resource.TestCheckResourceAttr(resRule, "source.user_groups.0.name", cfg.userGroups[0].Name),
					resource.TestCheckResourceAttr(resRule, "source.users.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resRule, "source.users.*",
						map[string]string{"name": cfg.users[1].Name, "id": cfg.users[1].ID},
					),
					resource.TestCheckResourceAttr(resRule, "tracking.%", "2"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.mailing_list.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.mailing_list.0.name", cfg.mailingLists[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.subscription_group.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.subscription_group.0.name", cfg.subscriptionGroups[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.#", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.0.%", "2"),
					resource.TestCheckResourceAttrSet(resRule, "tracking.alert.webhook.0.id"),
					resource.TestCheckResourceAttr(resRule, "tracking.alert.webhook.0.name", cfg.webhooks[0].Name),
					resource.TestCheckResourceAttr(resRule, "tracking.event.%", "1"),
					resource.TestCheckResourceAttr(resRule, "tracking.event.enabled", "true"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.%", "1"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.%", "2"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.category", "LOW"),
					resource.TestCheckResourceAttr(resRule, "user_attributes.risk_score.operator", "LTE"),
				),
			},
		},
		CheckDestroy: func(*terraform.State) error { publishPrivateAccessPolicy(t); return nil },
	})
}

type privAccessPolicyCfg struct {
	resName            string
	applications       testPrivateApps
	users              testUsers
	userGroups         []ref
	devices            []ref
	subscriptionGroups []ref
	webhooks           []ref
	mailingLists       []ref
	t                  *testing.T
}

func newPrivAccessPolicyCfg(t *testing.T) privAccessPolicyCfg {
	return privAccessPolicyCfg{
		resName:            getRandName("private_access_policy"),
		applications:       getPrivateApps(t),
		users:              getUsers(t),
		userGroups:         []ref{{Name: "group-1"}},               // TODO: fetch dynamically
		devices:            []ref{{Name: "Test device posture 1"}}, // TODO: fetch dynamically
		subscriptionGroups: []ref{{Name: "subscription_group"}},    // TODO: fetch dynamically
		webhooks:           []ref{{Name: "webhook_1"}},             // TODO: fetch dynamically
		mailingLists:       []ref{{Name: "mailing_list"}},          // TODO: fetch dynamically

		t: t,
	}
}

func (p privAccessPolicyCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(privAccessPolicyTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name":               p.resName,
		"Applications":       p.applications,
		"Devices":            p.devices,
		"Users":              p.users,
		"UserGroups":         p.userGroups,
		"SubscriptionGroups": p.subscriptionGroups,
		"Webhooks":           p.webhooks,
		"MailingLists":       p.mailingLists,
	}
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
	    { "name" = "{{ (index .Applications 0).Name }}" },
	    { "id" = "{{ (index .Applications 1).ID }}" }
	  ]
	  connection_origins = ["REMOTE", "SITE"]
	  countries = [
	    { "name" = "United States" },
	    { "id" = "CZ" }
	  ]
	  description = "rule1 description"
	  devices = [
	    { "name" = "{{ (index .Devices 0).Name }}" }
	  ]
	  enabled   = true
	  name      = "{{.Name}}"
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
	      { name = "{{ (index .Users 0).Name }}" },
	      { name = "{{ (index .Users 1).Name }}" }
	    ],
	    user_groups = [
	      { name = "{{ (index .UserGroups 0).Name }}" }
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
	        { name = "{{ (index .SubscriptionGroups 0).Name }}" }
	      ]
	      webhook = [
	        { name = "{{ (index .Webhooks 0).Name }}" }
	      ],
	      mailing_list = [
	        { name = "{{ (index .MailingLists 0).Name }}" }
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
	    "{{.Name}}" = {
	      index = 0
	      name  = "{{.Name}}"
	    }
	  }
	}
	`,

	`resource "cato_private_access_policy" "this" {
	  enabled = true
	}
	
	resource "cato_private_access_rule" "rule_1" {
	  action = "ALLOW"
	  active_period = {
	    effective_from     = "2026-02-20T03:04:00Z"
	    use_effective_from = true
	    expires_at         = "2027-03-20T03:04:00Z"
	    use_expires_at     = true
	  }
	  applications = [
	    { "id" = "{{ (index .Applications 1).ID }}" }
	  ]
	  connection_origins = ["REMOTE", "SITE"]
	  countries = [
	    { "name" = "United States" },
	    { "id" = "IT" }
	  ]
	  description = "rule1 description"
	  devices = [
	    { "name" = "{{ (index .Devices 0).Name }}" }
	  ]
	  enabled   = true
	  name      = "{{.Name}}"
	  platforms = ["LINUX"]
	  schedule = {
	    active_on = "WORKING_HOURS",
	    custom_timeframe = {
	      from = "2026-01-20T01:02:00Z",
	      to   = "2026-04-20T03:04:00Z"
	    },
	    custom_recurring = {
	      days = ["MONDAY", "TUESDAY", "FRIDAY"],
	      from = "08:05:00",
	      to   = "19:31:00"
	    }
	  }
	  source = {
	    users = [
	      { name = "{{ (index .Users 1).Name }}" }
	    ],
	    user_groups = [
	      { name = "{{ (index .UserGroups 0).Name }}" }
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
	        { name = "{{ (index .SubscriptionGroups 0).Name }}" }
	      ]
	      webhook = [
	        { name = "{{ (index .Webhooks 0).Name }}" }
	      ],
	      mailing_list = [
	        { name = "{{ (index .MailingLists 0).Name }}" }
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
	    "{{.Name}}" = {
	      index = 0
	      name  = "{{.Name}}"
	    }
	  }
	}
	`,
}
