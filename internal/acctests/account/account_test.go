//go:build acctest

package account

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"text/template"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

const envEnableAccountCRUD = "TFACC_ENABLE_ACCOUNT_CRUD"
const envEnableAccountCRUDAllowed = "TFACC_ACCOUNT_CRUD_ALLOWED"

func TestAccAccount(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableAccountCRUD) != "true" || os.Getenv(envEnableAccountCRUDAllowed) != "true" {
		t.Skipf("set %s=true and %s=true to run account CRUD acceptance test", envEnableAccountCRUD, envEnableAccountCRUDAllowed)
	}

	mockSrv := accmock.NewMockServer(t, "TestAccAccount")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newAccountCfg(t)
	res := "cato_account.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy:             testAccAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.name),
					resource.TestCheckResourceAttr(res, "description", cfg.name+" description"),
					resource.TestCheckResourceAttr(res, "tenancy", "SINGLE_TENANT"),
					resource.TestCheckResourceAttr(res, "type", "CUSTOMER"),
					resource.TestCheckResourceAttr(res, "timezone", "Europe/Paris"),
				),
			},
			{
				ImportState:  true,
				ResourceName: res,
			},
			{
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.name),
					resource.TestCheckResourceAttr(res, "description", cfg.name+" description updated"),
				),
			},
		},
	})
}

func TestAccAccount_ImmutableFieldValidation(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableAccountCRUD) != "true" || os.Getenv(envEnableAccountCRUDAllowed) != "true" {
		t.Skipf("set %s=true and %s=true to run account CRUD acceptance test", envEnableAccountCRUD, envEnableAccountCRUDAllowed)
	}

	mockSrv := accmock.NewMockServer(t, "TestAccAccount_ImmutableFieldValidation")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newAccountCfg(t)
	res := "cato_account.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.name),
				),
			},
			{
				Config:      cfg.getTfConfig(2),
				ExpectError: regexp.MustCompile("Account name cannot be changed after creation"),
			},
		},
	})
}

func testAccAccountDestroy(st *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, map[string]string{
		"User-Agent": "cato-terraform-test",
	})
	if err != nil {
		return err
	}
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "cato_account" {
			continue
		}

		_, readErr := client.AccountManagement(context.Background(), rs.Primary.ID)
		if readErr == nil {
			return fmt.Errorf("account %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

type accountCfg struct {
	name string
	t    *testing.T
}

func newAccountCfg(t *testing.T) accountCfg {
	return accountCfg{
		name: acc.GetRandName("account"),
		t:    t,
	}
}

func (p accountCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(accountTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}

	var buf bytes.Buffer
	data := map[string]any{
		"Name": p.name,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var accountTFs = []string{
	`resource "cato_account" "this" {
		name        = "{{.Name}}"
		description = "{{.Name}} description"
		timezone    = "Europe/Paris"
	}`,
	`resource "cato_account" "this" {
		name        = "{{.Name}}"
		description = "{{.Name}} description updated"
		timezone    = "Europe/Paris"
	}`,
	`resource "cato_account" "this" {
		name        = "{{.Name}}-renamed"
		description = "{{.Name}} description updated"
		timezone    = "Europe/Paris"
	}`,
}
