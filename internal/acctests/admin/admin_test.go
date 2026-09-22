//go:build acctest

package admin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"text/template"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAdmin(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccAdmin")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newAdminCfg(t)
	res := "cato_admin.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy:             testAccAdminDestroy,
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "%", "10"),
					resource.TestCheckResourceAttr(res, "account_id", cfg.accountID),
					resource.TestCheckResourceAttrSet(res, "admin_id"),
					resource.TestCheckResourceAttr(res, "email", "terraform-test-admin"+cfg.accountID+"@test.com"),
					resource.TestCheckResourceAttr(res, "first_name", "John"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "last_name", "Admin"),
					resource.TestCheckResourceAttr(res, "managed_roles.#", "1"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.%", "2"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.id", "1"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.name", "Editor"),
					resource.TestCheckResourceAttr(res, "mfa_enabled", "true"),
					resource.TestCheckResourceAttr(res, "password_never_expires", "true"),
					acc.PrintAttributes(res),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "account_id", cfg.accountID),
					resource.TestCheckResourceAttrSet(res, "admin_id"),
					resource.TestCheckResourceAttr(res, "email", "terraform-test-admin"+cfg.accountID+"@test.com"),
					resource.TestCheckResourceAttr(res, "first_name", "John 2"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "last_name", "Admin 2"),
					resource.TestCheckResourceAttr(res, "managed_roles.#", "1"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.%", "2"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.id", "2"),
					resource.TestCheckResourceAttr(res, "managed_roles.0.name", "Editor"),
					resource.TestCheckResourceAttr(res, "mfa_enabled", "true"),
					resource.TestCheckResourceAttr(res, "password_never_expires", "false"),
					acc.PrintAttributes(res),
				),
			},
		},
	})
}

func testAccAdminDestroy(st *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, map[string]string{
		"User-Agent": "cato-terraform-test",
	})
	if err != nil {
		return err
	}

	for _, rs := range st.RootModule().Resources {
		if rs.Type != "cato_admin" {
			continue
		}

		adminID := rs.Primary.Attributes["admin_id"]
		result, readErr := client.Admins(context.Background(), acc.CatoAccountID, nil, nil, nil, nil, []string{adminID})
		if readErr == nil && result.GetAdmins() != nil && len(result.GetAdmins().Items) > 0 {
			return fmt.Errorf("admin %s still exists", adminID)
		}
	}
	return nil
}

type adminCfg struct {
	accountID string
	t         *testing.T
}

func newAdminCfg(t *testing.T) adminCfg {
	return adminCfg{
		accountID: acc.CatoAccountID,
		t:         t,
	}
}

func (p adminCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(adminTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{"AccountID": p.accountID}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var adminTFs = []string{
	`resource "cato_admin" "this" {
		email                  = "terraform-test-admin{{.AccountID}}@test.com"
		first_name             = "John"
		last_name              = "Admin"
		password_never_expires = true

		managed_roles = [
			{ id = "1" }
		]
	}`,

	`resource "cato_admin" "this" {
		email                  = "terraform-test-admin{{.AccountID}}@test.com"
		first_name             = "John 2"
		last_name              = "Admin 2"
		password_never_expires = false

		managed_roles = [
			{ id = "2" }
		]
	}`,
}
