//go:build acctest

package if_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccInternetFw_Simple(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.simple"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigSimple(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.active_period.%", "4"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_effective_from", "false"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_expires_at", "false"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "0"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.schedule.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.active_on", "ALWAYS"),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfigSimple(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "BLOCK"),
					resource.TestCheckResourceAttr(res, "rule.active_period.%", "4"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_effective_from", "false"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_expires_at", "false"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "new.test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "0"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+" 2"),
					resource.TestCheckResourceAttr(res, "rule.schedule.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.active_on", "ALWAYS"),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
			},
		},
	})
}

func TestAccInternetFw_IDName(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_IDName")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.id_name"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigIDName(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
				),
			},
		},
	})
}

// TestAccInternetFw_Timeframe tests the datetime format - it should be returned in RFC3339
func TestAccInternetFw_Timeframe(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_Timeframe")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.timeframe"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigTimeframe(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
				),
			},
		},
	})
}

func TestAccInternetFw_UserID(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_UserID")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.user_id"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigUserID(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.name", "acctest_exception_uid_100"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.application.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.application.*",
						map[string]string{"name": "Gmail"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.service.standard.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.standard.*",
						map[string]string{"name": "IMAP"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.0", "10.20.30.40"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.0.name", cfg.users[0].Name),
					// resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.0.id", cfg.users[0].ID),
					// TODO: fix TF, add id to state
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: investigate & fix
			},
			{
				// Update the resource
				Config: cfg.getTfConfigUserID(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "BLOCK"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "new.test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.name", "acctest_exception_uid_100"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.application.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.application.*",
						map[string]string{"name": "Gmail"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.service.standard.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.standard.*",
						map[string]string{"name": "IMAP"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.0", "10.20.30.40"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.0.name", cfg.users[1].Name),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.0.id", cfg.users[1].ID), // TODO: fix tf
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+" 2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: investigate & fix
			},
		},
	})
}
func TestAccInternetFw_Full(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_Full")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.full"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigFull(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.%", "17"),
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.active_period.%", "4"),
					resource.TestCheckResourceAttr(res, "rule.active_period.effective_from", "2024-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(res, "rule.active_period.expires_at", "2124-12-31T23:59:59Z"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_effective_from", "true"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_expires_at", "true"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "SITE"),
					resource.TestCheckResourceAttr(res, "rule.country.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.country.*",
						map[string]string{"id": "FR", "name": "France"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.country.*",
						map[string]string{"id": "US", "name": "United States"},
					),
					resource.TestCheckResourceAttr(res, "rule.description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "rule.destination.%", "13"),
					resource.TestCheckResourceAttr(res, "rule.destination.app_category.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.app_category.*",
						map[string]string{"id": "advertisements", "name": "Advertisements"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.app_category.*",
						map[string]string{"id": "business_systems", "name": "Business Systems"},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.application.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.application.*",
						map[string]string{"id": "gmail", "name": "Gmail"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.application.*",
						map[string]string{"id": "zoom", "name": "Zoom"},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.country.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.country.*",
						map[string]string{"id": "FR", "name": "France"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.country.*",
						map[string]string{"id": "US", "name": "United States"},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.custom_app.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.custom_app.*",
						map[string]string{"id": cfg.customApps[0].ID, "name": cfg.customApps[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.custom_app.*",
						map[string]string{"id": cfg.customApps[1].ID, "name": cfg.customApps[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.custom_category.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.custom_category.*",
						map[string]string{"id": cfg.customCategories[0].ID, "name": cfg.customCategories[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.custom_category.*",
						map[string]string{"id": cfg.customCategories[1].ID, "name": cfg.customCategories[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.destination.domain.*", "one.example.com"),
					resource.TestCheckTypeSetElemAttr(res, "rule.destination.domain.*", "two.example.com"),
					resource.TestCheckResourceAttr(res, "rule.destination.fqdn.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.fqdn.0", "www.one.example.com"),
					resource.TestCheckResourceAttr(res, "rule.destination.global_ip_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[0].ID, "name": cfg.globalIPRanges[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[1].ID, "name": cfg.globalIPRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip.0", "192.168.111.4"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.ip_range.*",
						map[string]string{"from": "192.168.112.0", "to": "192.168.112.100"},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.remote_asn.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.destination.remote_asn.*", "1234"),
					resource.TestCheckTypeSetElemAttr(res, "rule.destination.remote_asn.*", "5678"),
					resource.TestCheckResourceAttr(res, "rule.destination.sanctioned_apps_category.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.sanctioned_apps_category.*",
						map[string]string{"name": "Sanctioned Apps"},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.0", "192.168.111.0/24"),
					resource.TestCheckResourceAttr(res, "rule.device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.device.*",
						map[string]string{"id": cfg.devicePostures[0].ID, "name": cfg.devicePostures[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.device.*",
						map[string]string{"id": cfg.devicePostures[1].ID, "name": cfg.devicePostures[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.%", "6"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.category.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.category.*", "IoT"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.category.*", "Mobile"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.manufacturer.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.manufacturer.*", "ADTRAN"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.manufacturer.*", "ACTi"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.model.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.model.*", " 9"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.model.*", " 7+"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.os.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.os.*", "Aruba OS"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.os.*", "Arch Linux"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.os_version.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.os_version.0", "10.0"),
					resource.TestCheckResourceAttr(res, "rule.device_attributes.type.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.type.*", "Appliance"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_attributes.type.*", "Analog Telephone Adapter"),
					resource.TestCheckResourceAttr(res, "rule.device_os.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_os.*", "WINDOWS"),
					resource.TestCheckTypeSetElemAttr(res, "rule.device_os.*", "MACOS"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.%", "9"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.connection_origin", "SITE"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.country.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.country.*",
						map[string]string{"id": "US"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.country.*",
						map[string]string{"name": "France"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.%", "13"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.app_category.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.app_category.*",
						map[string]string{"id": "business_systems"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.app_category.*",
						map[string]string{"name": "Advertisements"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.application.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.application.*",
						map[string]string{"id": "zoom"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.application.*",
						map[string]string{"name": "Gmail"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.country.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.country.*",
						map[string]string{"id": "US"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.country.*",
						map[string]string{"name": "France"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.custom_app.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.custom_app.*",
						map[string]string{"id": cfg.customApps[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.custom_app.*",
						map[string]string{"name": cfg.customApps[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.custom_category.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.custom_category.*",
						map[string]string{"id": cfg.customCategories[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.custom_category.*",
						map[string]string{"name": cfg.customCategories[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.domain.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.destination.domain.*", "one.example.com"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.destination.domain.*", "two.example.com"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.fqdn.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.fqdn.0", "www.one.example.com"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.global_ip_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.global_ip_range.*",
						map[string]string{"name": cfg.globalIPRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.ip.0", "192.168.111.4"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.ip_range.*",
						map[string]string{"from": "192.168.112.0", "to": "192.168.112.100"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.remote_asn.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.destination.remote_asn.*", "1234"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.destination.remote_asn.*", "5678"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.sanctioned_apps_category.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.destination.sanctioned_apps_category.*",
						map[string]string{"name": "Sanctioned Apps"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.destination.subnet.0", "192.168.111.0/24"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.device.*",
						map[string]string{"id": cfg.devicePostures[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.device.*",
						map[string]string{"name": cfg.devicePostures[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.%", "6"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.category.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.category.*", "IoT"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.category.*", "Mobile"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.manufacturer.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.manufacturer.*", "ADTRAN"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.manufacturer.*", "ACTi"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.model.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.model.*", " 9"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.model.*", " 7+"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.os.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.os.*", "Aruba OS"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.os.*", "Arch Linux"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.os_version.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.os_version.0", "10.0"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_attributes.type.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.type.*", "Appliance"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_attributes.type.*", "Analog Telephone Adapter"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.device_os.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_os.*", "WINDOWS"),
					resource.TestCheckTypeSetElemAttr(res, "rule.exceptions.0.device_os.*", "MACOS"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.name", "acctest_exception_100"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.service.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.service.custom.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.custom.*",
						map[string]string{"port.0": "8022", "protocol": "TCP"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.custom.*",
						map[string]string{"port_range.from": "6000", "port_range.to": "6010", "protocol": "TCP"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.service.standard.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.standard.*",
						map[string]string{"id": "ftp"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.service.standard.*",
						map[string]string{"name": "IMAP"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.%", "13"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.floating_subnet.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.floating_subnet.*",
						map[string]string{"id": cfg.floatingRanges[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.floating_subnet.*",
						map[string]string{"name": cfg.floatingRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.global_ip_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.global_ip_range.*",
						map[string]string{"name": cfg.globalIPRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.group.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.group.*",
						map[string]string{"id": cfg.groups[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.group.*",
						map[string]string{"name": cfg.groups[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.host.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.host.*",
						map[string]string{"name": cfg.hosts[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip.0", "10.20.30.40"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.ip_range.*",
						map[string]string{"from": "192.168.112.0", "to": "192.168.112.100"},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.network_interface.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.network_interface.*",
						map[string]string{"id": cfg.interfaces[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.site.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.site.*",
						map[string]string{"id": cfg.sites[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.site.*",
						map[string]string{"name": cfg.sites[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.site_network_subnet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.site_network_subnet.*",
						map[string]string{"id": cfg.siteRanges[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.subnet.0", "10.20.30.0/24"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.system_group.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.system_group.*",
						map[string]string{"id": cfg.systemGroups[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.system_group.*",
						map[string]string{"name": cfg.systemGroups[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.user.*",
						map[string]string{"id": cfg.users[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.exceptions.0.source.users_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.exceptions.*.source.users_group.*",
						map[string]string{"id": cfg.usersGroups[0].ID},
					),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.schedule.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.active_on", "ALWAYS"),
					resource.TestCheckResourceAttr(res, "rule.schedule.custom_recurring.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.custom_recurring.days.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "rule.schedule.custom_recurring.days.*", "MONDAY"),
					resource.TestCheckTypeSetElemAttr(res, "rule.schedule.custom_recurring.days.*", "FRIDAY"),
					resource.TestCheckResourceAttr(res, "rule.schedule.custom_recurring.from", "08:04:00"),
					resource.TestCheckResourceAttr(res, "rule.schedule.custom_recurring.to", "19:30:00"),
					resource.TestCheckResourceAttr(res, "rule.service.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.custom.*",
						map[string]string{"port.0": "8022", "protocol": "TCP"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.custom.*",
						map[string]string{"port_range.from": "6000", "port_range.to": "6010", "protocol": "TCP"},
					),
					resource.TestCheckResourceAttr(res, "rule.service.standard.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.standard.*",
						map[string]string{"id": "ftp", "name": "FTP"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.standard.*",
						map[string]string{"id": "imap", "name": "IMAP"},
					),
					resource.TestCheckResourceAttr(res, "rule.source.%", "13"),
					resource.TestCheckResourceAttr(res, "rule.source.floating_subnet.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.floating_subnet.*",
						map[string]string{"id": cfg.floatingRanges[0].ID, "name": cfg.floatingRanges[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.floating_subnet.*",
						map[string]string{"id": cfg.floatingRanges[1].ID, "name": cfg.floatingRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.global_ip_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[0].ID, "name": cfg.globalIPRanges[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[1].ID, "name": cfg.globalIPRanges[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.group.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.group.*",
						map[string]string{"id": cfg.groups[0].ID, "name": cfg.groups[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.group.*",
						map[string]string{"id": cfg.groups[1].ID, "name": cfg.groups[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.host.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.host.*",
						map[string]string{"id": cfg.hosts[0].ID, "name": cfg.hosts[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.host.*",
						map[string]string{"id": cfg.hosts[1].ID, "name": cfg.hosts[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.ip.0", "10.99.12.31"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.ip_range.*",
						map[string]string{"from": "10.99.12.10", "to": "10.99.12.20"},
					),
					resource.TestCheckResourceAttr(res, "rule.source.network_interface.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.network_interface.*",
						map[string]string{"id": cfg.interfaces[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.site.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.site.*",
						map[string]string{"id": cfg.sites[0].ID, "name": cfg.sites[0].Name},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.site.*",
						map[string]string{"id": cfg.sites[1].ID, "name": cfg.sites[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.site_network_subnet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.site_network_subnet.*",
						map[string]string{"id": cfg.siteRanges[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.0", "10.99.12.0/24"),
					resource.TestCheckResourceAttr(res, "rule.source.system_group.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.system_group.*",
						map[string]string{"id": cfg.systemGroups[0].ID},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.system_group.*",
						map[string]string{"name": cfg.systemGroups[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.user.*",
						map[string]string{"id": cfg.users[0].ID, "name": cfg.users[0].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.source.users_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.users_group.*",
						map[string]string{"id": cfg.usersGroups[0].ID, "name": cfg.usersGroups[0].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.mailing_list.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.tracking.alert.mailing_list.*",
						map[string]string{"id": cfg.mailingLists[1].ID, "name": cfg.mailingLists[1].Name},
					),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: investigate & fix
			},
			/* TODO: fix the provider, enable update test
			{
				// Update the resource
				Config: cfg.getTfConfigFull(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
				),
			},
			*/
		},
	})
}

type internetFwCfg struct {
	resName            string
	hosts              []acc.Ref
	sites              []acc.Ref
	globalIPRanges     []acc.Ref
	siteRanges         []acc.Ref
	floatingRanges     []acc.Ref
	interfaces         []acc.Ref
	users              []acc.Ref
	usersGroups        []acc.Ref
	groups             []acc.Ref
	systemGroups       []acc.Ref
	devicePostures     []acc.Ref
	customApps         []acc.Ref
	customCategories   []acc.Ref
	subscriptionGroups []acc.Ref
	webhooks           []acc.Ref
	mailingLists       []acc.Ref
	t                  *testing.T
}

func newInternetFwCfg(t *testing.T) internetFwCfg {
	return internetFwCfg{
		resName:            acc.GetRandName("internet_fw"),
		hosts:              acc.GetHosts(t),
		sites:              acc.GetSites(t),
		globalIPRanges:     acc.GetGlobalIPRanges(t),
		siteRanges:         acc.GetSiteRanges(t),
		floatingRanges:     acc.GetFloatingRanges(t),
		interfaces:         acc.GetInterfaces(t),
		users:              acc.GetUsers(t),
		usersGroups:        acc.GetUserGroups(t),
		groups:             acc.GetAdvancedGroups(t),
		systemGroups:       acc.GetSystemGroups(t),
		devicePostures:     acc.GetDevicePostures(t),
		customApps:         acc.GetCustomApps(t),
		customCategories:   acc.GetCustomCategories(t),
		subscriptionGroups: acc.GetSubscriptionGroups(t),
		webhooks:           acc.GetWebhooks(t),
		mailingLists:       acc.GetMailingLists(t),
		t:                  t,
	}
}

func (p internetFwCfg) prepareTfCfg(data map[string]any, tmplText string) string {
	tmpl, err := template.New("tmpl").Parse(tmplText)
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}
	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

// ------------------------------------------------------------------
// Simple cato_if_rule configurations
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, internetFwSimpleTFs[index])
}

var internetFwSimpleTFs = []string{
	`resource "cato_if_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }}"
			enabled = true
			action  = "ALLOW"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "test.com" ]
			}
			source = {}
		}
	}
	`,

	`resource "cato_if_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }} 2"
			enabled = false
			action  = "BLOCK"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "new.test.com" ]
			}
			source = {}
		}
	}
	`,
}

// ------------------------------------------------------------------
// IDNAme cato_if_rule configurations
// - test combination of name and ID attributes in the rule configuration
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigIDName(index int) string {
	data := map[string]any{
		"Name":  p.resName,
		"Users": p.users,
	}
	return p.prepareTfCfg(data, internetFwIDNameTFs[index])
}

var internetFwIDNameTFs = []string{
	`resource "cato_if_rule" "id_name" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }}"
			enabled = true
			action  = "ALLOW"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "test.com" ]
			}
			source = {
				user = [
					{ id   = "{{ (index .Users 0).ID }}" },
					{ name = "{{ (index .Users 1).Name }}" },
				]
			}
		}
	}
	`,
}

// ------------------------------------------------------------------
// Timeframe cato_if_rule configurations
// - test combination of name and ID attributes in the rule configuration
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigTimeframe(index int) string {
	data := map[string]any{
		"Name":  p.resName,
		"Users": p.users,
	}
	return p.prepareTfCfg(data, internetFwTimeframeTFs[index])
}

var internetFwTimeframeTFs = []string{
	`resource "cato_if_rule" "timeframe" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }}"
			enabled = true
			action  = "ALLOW"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "test.com" ]
			}
			source = { }
			schedule = {
				active_on = "ALWAYS"
				custom_timeframe = {
					from = "2026-02-20T01:02:00Z",
					to = "2026-02-20T03:04:00Z"
				}
			}
		}
	}
	`,
}

// ------------------------------------------------------------------
// UserID cato_if_rule configurations
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigUserID(index int) string {
	data := map[string]any{
		"Name":  p.resName,
		"Users": p.users,
	}
	return p.prepareTfCfg(data, internetFwUserIDTFs[index])
}

var internetFwUserIDTFs = []string{
	`resource "cato_if_rule" "user_id" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }}"
			enabled = true
			action  = "ALLOW"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "test.com" ]
			}
			source = {}
			exceptions = [
				{
					name = "acctest_exception_uid_100"
					service = {
						standard = [ { name = "IMAP" } ]
					}
					source = {
						ip = [ "10.20.30.40" ]
						user = [
							{ name = "{{ (index .Users 0).Name }}" },
						]
					}
					destination = {
						application = [ { name = "Gmail" } ]
					}
				}
			]
		}
	}
	`,

	`resource "cato_if_rule" "user_id" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }} 2"
			enabled = false
			action  = "BLOCK"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "new.test.com" ]
			}
			source = {}
			exceptions = [
				{
					name = "acctest_exception_uid_100"
					service = {
						standard = [ { name = "IMAP" } ]
					}
					source = {
						ip = [ "10.20.30.40" ]
						user = [
							{ name = "{{ (index .Users 1).Name }}" },
						]
					}
					destination = {
						application = [ { name = "Gmail" } ]
					}
				}
			]
		}
	}
	`,
}

// ------------------------------------------------------------------
// Full cato_if_rule configurations
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigFull(index int) string {
	data := map[string]any{
		"Name":               p.resName,
		"Hosts":              p.hosts,
		"Sites":              p.sites,
		"GlobalIPRanges":     p.globalIPRanges,
		"SiteRanges":         p.siteRanges,
		"FloatingRanges":     p.floatingRanges,
		"Interfaces":         p.interfaces,
		"Users":              p.users,
		"UserGroups":         p.usersGroups,
		"Groups":             p.groups,
		"SystemGroups":       p.systemGroups,
		"DevicePostures":     p.devicePostures,
		"CustomApps":         p.customApps,
		"CustomCategories":   p.customCategories,
		"SubscriptionGroups": p.subscriptionGroups,
		"Webhooks":           p.webhooks,
		"MailingLists":       p.mailingLists,
	}
	fmt.Printf(".GlobalIPRanges=%#v", data["GlobalIPRanges"])
	return p.prepareTfCfg(data, internetFwFullTFs[index])
}

var internetFwFullTFs = []string{
	`resource "cato_if_rule" "full" {
		at   = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name                         = "{{ .Name }}"
			description                  = "{{ .Name }} description"
			enabled                      = true
			active_period = {
				effective_from = "2024-01-01T00:00:00Z"
				expires_at	 = "2124-12-31T23:59:59Z"
			}
			source = {
				ip = ["10.99.12.31"]
				host = [
					{ id   = "{{ (index .Hosts 0).ID }}" },
					{ name = "{{ (index .Hosts 1).Name }}" },
				]
				site = [
					{ id   = "{{ (index .Sites 0).ID }}" },
					{ name = "{{ (index .Sites 1).Name }}" },
				]
				subnet = [
					"10.99.12.0/24"
				]	
				ip_range = [
					{ from   = "10.99.12.10", to = "10.99.12.20" },
				]
				global_ip_range = [
					{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
					{ name = "{{ (index .GlobalIPRanges 1).Name }}" }
				]
				network_interface = [
					{ id   = "{{ (index .Interfaces 0).ID }}" },
				]
				site_network_subnet = [
					{ id   = "{{ (index .SiteRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id   = "{{ (index .FloatingRanges 0).ID }}" },
					{ name = "{{ (index .FloatingRanges 1).Name }}" },
				]
				user = [
					{ id   = "{{ (index .Users 0).ID }}" },
					# { name = "{{ (index .Users 1).Name }}" },
				]
				users_group = [
					{ id   = "{{ (index .UserGroups 0).ID }}" },
					# { name = "{{ (index .UserGroups 1).Name }}" },
				]
				group = [
					{ id   = "{{ (index .Groups 0).ID }}" },
					{ name = "{{ (index .Groups 1).Name }}" },
				]
				system_group = [
					{ id   = "{{ (index .SystemGroups 0).ID }}" },
					{ name = "{{ (index .SystemGroups 1).Name }}" },
				]
			}
			connection_origin = "SITE"
			country = [
				{ id   = "US" },
				{ name = "France" },
			]
			device = [
				{ id   = "{{ (index .DevicePostures 0).ID }}" },
				{ name = "{{ (index .DevicePostures 1).Name }}" },
			]
			device_os = [
				"WINDOWS",
				"MACOS",
			]
			device_attributes = {
				category     = [
					"IoT",
					"Mobile",
				]
				type         = [
					"Appliance",
					"Analog Telephone Adapter",
				]
				model        = [
					" 9",
					" 7+",
				]
				manufacturer = [
					"ADTRAN",
					"ACTi",
				]
				os = [
					"Aruba OS",
					"Arch Linux",
				]
				os_version = [
					"10.0"
				]
			}
			destination = {
				application = [
					{ name = "Gmail" },
					{ id   = "zoom" },
				]
				custom_app = [
					{ id   = "{{ (index .CustomApps 0).ID }}" },
					{ name = "{{ (index .CustomApps 1).Name }}" },
				]
				app_category = [
					{ id   = "business_systems" },
					{ name = "Advertisements" },
				]
				custom_category = [
					{ id   = "{{ (index .CustomCategories 0).ID }}" },
					{ name = "{{ (index .CustomCategories 1).Name }}" },
				]
				sanctioned_apps_category = [
					{ name = "Sanctioned Apps" }
				]
				country = [
					{ id   = "US" },
					{ name = "France" },
				]
				domain = [
					"one.example.com",
					"two.example.com",
				]
				fqdn = [
					"www.one.example.com"
				]
				ip = [
					"192.168.111.4"
				]
				subnet = [
					"192.168.111.0/24"
				]
				ip_range = [
					{ from = "192.168.112.0", to = "192.168.112.100" },
				]
				global_ip_range = [
					{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
					{ name = "{{ (index .GlobalIPRanges 1).Name }}" }
				]
				remote_asn = [
					"1234",
					"5678",
				]
			}
			service = {
				standard = [
					{ name = "IMAP" },
					{ id = "ftp" },
				]
				custom = [
					{ port = [ "8022" ], protocol = "TCP" },
					{ port_range = { from = 6000, to = 6010 }, protocol = "TCP" },
				]
			}
			action = "ALLOW"
			tracking =  {
				event = {
					enabled = true
				}
				alert = {
					enabled = true
					frequency = "DAILY"
					#subscription_group = [
					#	{ id   = "{{ (index .SubscriptionGroups 0).ID }}" },
					#	#{ name = "{{ (index .SubscriptionGroups 1).Name }}" }
					#]
					#webhook = [
					#	{ id   = "{{ (index .Webhooks 0).ID }}" },
					#	#{ name = "{{ (index .Webhooks 1).Name }}" }
					#]
					mailing_list = [
						#{ id   = "{{ (index .MailingLists 0).ID }}" },
						{ name = "{{ (index .MailingLists 1).Name }}" }
					]
				}
			}
			schedule = {
				active_on = "ALWAYS"
				#custom_timeframe = {
				#	from = "2026-02-20T01:02:00Z",
				#	to = "2026-02-20T03:04:00Z"
				#}
				custom_recurring = {
					days = [ "MONDAY", "FRIDAY" ],
					from =  "08:04:00",
					to   = "19:30:00"
				}
			}
			exceptions = [
				{
					name = "acctest_exception_100"
					source = {
						ip = [ "10.20.30.40" ]
						host = [
							{ name = "{{ (index .Hosts 1).Name }}" }
						]
						site = [
							{ id = "{{ (index .Sites 0).ID }}" },
							{ name = "{{ (index .Sites 1).Name }}" },
						]
						subnet = [
							"10.20.30.0/24"
						]
						ip_range = [
							{ from = "192.168.112.0", to = "192.168.112.100" },
						]
						global_ip_range = [
							{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
							{ name = "{{ (index .GlobalIPRanges 1).Name }}" }
						]
						network_interface = [
							{ id   = "{{ (index .Interfaces 0).ID }}" },
						]
						site_network_subnet = [
							{ id   = "{{ (index .SiteRanges 0).ID }}" },
						]
						floating_subnet = [
							{ id   = "{{ (index .FloatingRanges 0).ID }}" },
							{ name = "{{ (index .FloatingRanges 1).Name }}" },
						]
						user = [
							{ id   = "{{ (index .Users 0).ID }}" },
							# { name = "{{ (index .Users 1).Name }}" },
						]
						users_group = [
							{ id   = "{{ (index .UserGroups 0).ID }}" },
							# { name = "{{ (index .UserGroups 1).Name }}" },
						]
						group = [
							{ id   = "{{ (index .Groups 0).ID }}" },
							{ name = "{{ (index .Groups 1).Name }}" },
						]
						system_group = [
							{ id   = "{{ (index .SystemGroups 0).ID }}" },
							{ name = "{{ (index .SystemGroups 1).Name }}" },
						]
					}
					country = [
						{ id   = "US" },
						{ name = "France" },
					]
					device = [
						{ id   = "{{ (index .DevicePostures 0).ID }}" },
						{ name = "{{ (index .DevicePostures 1).Name }}" },
					]
					device_attributes = {
						category     = [
							"IoT",
							"Mobile",
						]
						type         = [
							"Appliance",
							"Analog Telephone Adapter",
						]
						model        = [
							" 9",
							" 7+",
						]
						manufacturer = [
							"ADTRAN",
							"ACTi",
						]
						os = [
							"Aruba OS",
							"Arch Linux",
						]
						os_version = [
							"10.0"
						]
					}
					device_os = [
						"WINDOWS",
						"MACOS",
					]
					destination = {
						application = [
							{ name = "Gmail" },
							{ id   = "zoom" },
						]
						custom_app = [
							{ id   = "{{ (index .CustomApps 0).ID }}" },
							{ name = "{{ (index .CustomApps 1).Name }}" },
						]
						app_category = [
							{ id   = "business_systems" },
							{ name = "Advertisements" },
						]
						custom_category = [
							{ id   = "{{ (index .CustomCategories 0).ID }}" },
							{ name = "{{ (index .CustomCategories 1).Name }}" },
						]
						sanctioned_apps_category = [
							{ name = "Sanctioned Apps" }
						]
						country = [
							{ id   = "US" },
							{ name = "France" },
						]
						domain = [
							"one.example.com",
							"two.example.com",
						]
						fqdn = [
							"www.one.example.com"
						]
						ip = [
							"192.168.111.4"
						]
						subnet = [
							"192.168.111.0/24"
						]
						ip_range = [
							{ from = "192.168.112.0", to = "192.168.112.100" },
						]
						global_ip_range = [
							{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
							{ name = "{{ (index .GlobalIPRanges 1).Name }}" }
						]
						remote_asn = [
							"1234",
							"5678",
						]
					}
					service = {
						standard = [
							{ name = "IMAP" },
							{ id = "ftp" },
						]
						custom = [
							{ port = [ "8022" ], protocol = "TCP" },
							{ port_range = { from = 6000, to = 6010 }, protocol = "TCP" },
						]
					}
					connection_origin = "SITE"
				}
			]
		}
	}
	`,

	// Updated FULL internet fw rule
	`resource "cato_if_rule" "full" {
		at   = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name                         = "{{ .Name }} 2"
			description                  = "{{ .Name }} description 2"
			enabled                      = false
			active_period = {
				effective_from = "2024-01-02T00:00:00Z"
				expires_at	 = "2125-12-31T23:59:59Z"
			}
			source = {
				ip = ["10.99.12.32"]
				host = [
					{ id   = "{{ (index .Hosts 0).ID }}" },
					{ name = "{{ (index .Hosts 2).Name }}" },
				]
				site = [
					{ id   = "{{ (index .Sites 0).ID }}" },
					{ name = "{{ (index .Sites 2).Name }}" },
				]
				subnet = [
					"10.99.13.0/24"
				]	
				ip_range = [
					{ from   = "10.99.13.10", to = "10.99.13.20" },
				]
				global_ip_range = [
					{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
					{ name = "{{ (index .GlobalIPRanges 2).Name }}" }
				]
				network_interface = [
					{ id   = "{{ (index .Interfaces 1).ID }}" },
				]
				site_network_subnet = [
					{ id   = "{{ (index .SiteRanges 1).ID }}" },
				]
				floating_subnet = [
					{ name = "{{ (index .FloatingRanges 2).Name }}" },
				]
				user = [
					{ id   = "{{ (index .Users 1).ID }}" },
					# { name = "{{ (index .Users 1).Name }}" },
				]
				users_group = [
					{ id   = "{{ (index .UserGroups 1).ID }}" },
					# { name = "{{ (index .UserGroups 1).Name }}" },
				]
				group = [
					{ id   = "{{ (index .Groups 2).ID }}" },
				]
				system_group = [
					{ name   = "{{ (index .SystemGroups 2).Name }}" },
					#{ id   = "{{ (index .SystemGroups 2).ID }}" },
					{ name = "{{ (index .SystemGroups 1).Name }}" },
				]
			}
			connection_origin = "ANY"
			country = [
				{ id   = "IT" },
				{ name = "Germany" },
			]
			device = [
				{ id   = "{{ (index .DevicePostures 0).ID }}" },
				{ name = "{{ (index .DevicePostures 2).Name }}" },
			]
			device_os = [
				"LINUX",
				"MACOS",
			]
			device_attributes = {
				category     = [
					"IoT",
					"Mobile",
				]
				type         = [
					"Appliance",
					"Analog Telephone Adapter",
				]
				model        = [
					" 9",
					" 7+",
				]
				manufacturer = [
					"APPLE",
					"ACTi",
				]
				os = [
					"Aruba OS",
					"Arch Linux",
				]
				os_version = [
					"10.1"
				]
			}
			destination = {
				application = [
					{ name = "Gmail" },
					{ id   = "hibob" },
				]
				custom_app = [
					{ id   = "{{ (index .CustomApps 0).ID }}" },
					{ name = "{{ (index .CustomApps 2).Name }}" },
				]
				app_category = [
					{ id   = "business_systems" },
					{ name = "Advertisements" },
				]
				custom_category = [
					{ id   = "{{ (index .CustomCategories 0).ID }}" },
					{ name = "{{ (index .CustomCategories 2).Name }}" },
				]
				sanctioned_apps_category = [
					{ name = "Sanctioned Apps" }
				]
				country = [
					{ id   = "IT" },
					{ name = "Germany" },
				]
				domain = [
					"three.example.com",
					"two.example.com",
				]
				fqdn = [
					"www.three.example.com"
				]
				ip = [
					"192.168.112.4"
				]
				subnet = [
					"192.168.112.0/24"
				]
				ip_range = [
					{ from = "192.168.112.0", to = "192.168.112.100" },
				]
				global_ip_range = [
					{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
					{ name = "{{ (index .GlobalIPRanges 2).Name }}" }
				]
				remote_asn = [
					"1235",
					"5679",
				]
			}
			service = {
				standard = [
					{ id = "ftp" },
					{ id = "telnet" },
				]
				custom = [
					{ port = [ "8023" ], protocol = "TCP" },
					{ port_range = { from = 7000, to = 7010 }, protocol = "UDP" },
				]
			}
			action = "BLOCK"
			tracking =  {
				event = {
					enabled = false
				}
				alert = {
					enabled = true
					frequency = "HOURLY"
					#subscription_group = [
					#	{ id   = "{{ (index .SubscriptionGroups 0).ID }}" },
					#	#{ name = "{{ (index .SubscriptionGroups 1).Name }}" }
					#]
					#webhook = [
					#	{ id   = "{{ (index .Webhooks 0).ID }}" },
					#	#{ name = "{{ (index .Webhooks 1).Name }}" }
					#]
					mailing_list = [
						#{ id   = "{{ (index .MailingLists 0).ID }}" },
						{ name = "{{ (index .MailingLists 1).Name }}" }
					]
				}
			}
			schedule = {
				active_on = "ALWAYS"
				#custom_timeframe = {
				#	from = "2026-02-20T01:02:00Z",
				#	to = "2026-02-20T03:04:00Z"
				#}
				custom_recurring = {
					days = [ "MONDAY", "TUESDAY" ],
					from =  "08:05:00",
					to   = "19:31:00"
				}
			}
			exceptions = [
				{
					name = "acctest_exception_101"
					source = {
						ip = [ "10.20.30.41" ]
						host = [
							{ name = "{{ (index .Hosts 0).Name }}" }
						]
						site = [
							{ id = "{{ (index .Sites 0).ID }}" },
							{ name = "{{ (index .Sites 2).Name }}" },
						]
						subnet = [
							"10.20.31.0/24"
						]
						ip_range = [
							{ from = "192.168.113.0", to = "192.168.113.100" },
						]
						global_ip_range = [
							{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
							{ name = "{{ (index .GlobalIPRanges 2).Name }}" }
						]
						network_interface = [
							{ id   = "{{ (index .Interfaces 1).ID }}" },
						]
						site_network_subnet = [
							{ id   = "{{ (index .SiteRanges 1).ID }}" },
						]
						floating_subnet = [
							{ id   = "{{ (index .FloatingRanges 0).ID }}" },
							{ name = "{{ (index .FloatingRanges 2).Name }}" },
						]
						user = [
							{ id   = "{{ (index .Users 0).ID }}" },
							# { name = "{{ (index .Users 2).Name }}" },
						]
						users_group = [
							{ id   = "{{ (index .UserGroups 0).ID }}" },
							# { name = "{{ (index .UserGroups 2).Name }}" },
						]
						group = [
							{ id   = "{{ (index .Groups 0).ID }}" },
							{ name = "{{ (index .Groups 2).Name }}" },
						]
						system_group = [
							{ id   = "{{ (index .SystemGroups 0).ID }}" },
							{ name = "{{ (index .SystemGroups 2).Name }}" },
						]
					}
					country = [
						{ id   = "IT" },
						{ name = "Belgium" },
					]
					device = [
						{ id   = "{{ (index .DevicePostures 0).ID }}" },
						{ name = "{{ (index .DevicePostures 2).Name }}" },
					]
					device_attributes = {
						category     = [
							"IoT",
							"Mobile",
						]
						type         = [
							"Appliance",
							"Analog Telephone Adapter",
						]
						model        = [
							" 9",
							" 8+",
						]
						manufacturer = [
							"ACTi",
						]
						os = [
							"Aruba OS",
							"Arch Linux",
						]
						os_version = [
							"10.1"
						]
					}
					device_os = [
						"WINDOWS",
						"LINUX",
					]
					destination = {
						application = [
							{ name = "Hibob" },
							{ id   = "zoom" },
						]
						custom_app = [
							{ id   = "{{ (index .CustomApps 0).ID }}" },
							{ name = "{{ (index .CustomApps 2).Name }}" },
						]
						app_category = [
							{ id   = "business_systems" },
							{ name = "Anonymizers" },
						]
						custom_category = [
							{ id   = "{{ (index .CustomCategories 0).ID }}" },
							{ name = "{{ (index .CustomCategories 2).Name }}" },
						]
						sanctioned_apps_category = [
							{ name = "Sanctioned Apps" }
						]
						country = [
							{ id   = "IT" },
							{ name = "Germany" },
						]
						domain = [
							"one.example.com",
							"three.example.com",
						]
						fqdn = [
							"www.three.example.com"
						]
						ip = [
							"192.168.112.4"
						]
						subnet = [
							"192.168.112.0/24"
						]
						ip_range = [
							{ from = "192.168.112.0", to = "192.168.112.100" },
						]
						global_ip_range = [
							{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
							{ name = "{{ (index .GlobalIPRanges 2).Name }}" }
						]
						remote_asn = [
							"1235",
							"5679",
						]
					}
					service = {
						standard = [
							{ name = "IMAP" },
							{ id = "telnet" },
						]
						custom = [
							{ port = [ "8020" ], protocol = "UDP" },
							{ port_range = { from = 7000, to = 7010 }, protocol = "TCP" },
						]
					}
					connection_origin = "ANY"
				}
			]
		}
	}
	`,
}

// TODO: fix source.network_interface.name  ("aws-site \ LAN 01" does not work)
// TODO: fix source.site_network_subnet.name
// TODO: fix API bug when source.user has both id and name
// TODO: fix API bug when source.users_group has both id and name, there are also duplicate names
// TODO: fix API bug when tracking.alert.subscription_group has both id and name
// TODO: fix API bug when tracking.alert.webhook has both id and name
// TODO: fix API bug when tracking.alert.mailing_list has both id and name
// TODO: fix API bug when scheduke.custom_timeframe.from/to
//		 unexpected new value: .rule.schedule.custom_timeframe.to: was
//		 cty.StringVal("2026-02-20T03:04:00Z"), but now
//		 cty.StringVal("2026-02-20T03:04:00.000").
//		 -> implement workaround at the TF level -> parse responses

// TODO: add tests for tracking.alert.webhook / subscription_group / mailing_list (only 1 at a time is allowed)

// TODO: fix TF bugs - updating rule does not work - allow update step in TestAccInternetFw_Full
