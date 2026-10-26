//go:build acctest

package always_on_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAlwaysOnRule(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, t.Name())
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newAlwaysOnRuleCfg(t)
	res := "cato_always_on_rule.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "ENFORCE"),
					resource.TestCheckResourceAttr(res, "rule.allow_fail_open", "true"),
					resource.TestCheckResourceAttr(res, "rule.allow_user_bypass", "true"),
					resource.TestCheckResourceAttr(res, "rule.anti_tamper_mode", "MONITOR"),
					resource.TestCheckResourceAttr(res, "rule.bypass_duration.time", "100"),
					resource.TestCheckResourceAttr(res, "rule.bypass_duration.unit", "MINUTES"),
					resource.TestCheckResourceAttr(res, "rule.description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "rule.device_posture_profile.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.device_posture_profile.*",
						map[string]string{"id": cfg.devicePostures[0].ID, "name": cfg.devicePostures[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.device_posture_profile.*",
						map[string]string{"id": cfg.devicePostures[1].ID, "name": cfg.devicePostures[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttrSet(res, "rule.index"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.platform.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.platform.*", "ANDROID"),
					resource.TestCheckTypeSetElemAttr(res, "rule.platform.*", "LINUX"),
					resource.TestCheckResourceAttr(res, "rule.source.users.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.users.*",
						map[string]string{"id": cfg.users[0].ID, "name": cfg.users[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.users.*",
						map[string]string{"id": cfg.users[1].ID, "name": cfg.users[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.user_groups.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.user_groups.*",
						map[string]string{"id": cfg.userGroups[0].ID, "name": cfg.userGroups[0].Name},
					),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "IGNORE"),
					resource.TestCheckResourceAttr(res, "rule.allow_fail_open", "false"),
					resource.TestCheckResourceAttr(res, "rule.allow_user_bypass", "false"),
					resource.TestCheckResourceAttr(res, "rule.anti_tamper_mode", "PROTECT"),
					resource.TestCheckResourceAttr(res, "rule.bypass_duration.time", "2"),
					resource.TestCheckResourceAttr(res, "rule.bypass_duration.unit", "HOURS"),
					resource.TestCheckResourceAttr(res, "rule.description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttr(res, "rule.device_posture_profile.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.device_posture_profile.*",
						map[string]string{"id": cfg.devicePostures[0].ID, "name": cfg.devicePostures[0].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttrSet(res, "rule.index"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.platform.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.platform.*", "MACOS"),
					resource.TestCheckTypeSetElemAttr(res, "rule.platform.*", "LINUX"),
					resource.TestCheckResourceAttr(res, "rule.source.users.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.users.*",
						map[string]string{"id": cfg.users[1].ID, "name": cfg.users[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.user_groups.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.user_groups.*",
						map[string]string{"id": cfg.userGroups[1].ID, "name": cfg.userGroups[1].Name},
					),
				),
			},
		},
	})
}

type alwaysOnRuleCfg struct {
	resName        string
	users          []acc.Ref
	userGroups     []acc.Ref
	devicePostures []acc.Ref
	t              *testing.T
}

func newAlwaysOnRuleCfg(t *testing.T) alwaysOnRuleCfg {
	return alwaysOnRuleCfg{
		resName:        acc.GetRandName("always_on_rule"),
		users:          acc.GetUsers(t),
		userGroups:     acc.GetUserGroups(t),
		devicePostures: acc.GetDevicePostures(t),
		t:              t,
	}
}

func (p alwaysOnRuleCfg) getTfConfig(index int) string {
	p.t.Helper()
	tmpl, err := template.New("tmpl").Parse(alwaysOnRuleTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name":           p.resName,
		"Description":    p.resName + " description",
		"Users":          p.users,
		"UserGroups":     p.userGroups,
		"DevicePostures": p.devicePostures,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var alwaysOnRuleTFs = []string{
	// Create resource
	`resource "cato_always_on_rule" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			action            = "ENFORCE"
			allow_fail_open   = true
			allow_user_bypass = true
			anti_tamper_mode  = "MONITOR"
			bypass_duration = {
				time = 100
				unit = "MINUTES"
			}
			description       = "{{.Name}} description"
			device_posture_profile = [
	              { name = "{{ (index .DevicePostures 0).Name }}" },
	              { name = "{{ (index .DevicePostures 1).Name }}" },
			]
			enabled           = true
	  		name              = "{{.Name}}"
			platform = [
				"ANDROID",
				"LINUX",
			]
			source = {
	            users = [
	              { name = "{{ (index .Users 0).Name }}" },
	              { name = "{{ (index .Users 1).Name }}" }
	            ],
	            user_groups = [
	              { name = "{{ (index .UserGroups 0).Name }}" }
	            ]
			}
		}
	}
	`,

	// Update resource
	`resource "cato_always_on_rule" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			action            = "IGNORE"
			allow_fail_open   = false
			allow_user_bypass = false
			anti_tamper_mode  = "PROTECT"
			bypass_duration = {
				time = 2
				unit = "HOURS"
			}
			description       = "{{.Name}} description 2"
			device_posture_profile = [
	              { id = "{{ (index .DevicePostures 0).ID }}" },
			]
			enabled           = true
	  		name              = "{{.Name}}"
			platform = [
				"MACOS",
				"LINUX",
			]
			source = {
	            users = [
	              { name = "{{ (index .Users 1).Name }}" }
	            ],
	            user_groups = [
	              { id = "{{ (index .UserGroups 1).ID }}" }
	            ]
			}
		}
	}
	`,
}
