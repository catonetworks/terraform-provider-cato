//go:build acctest

package fqdncontainer

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccFQDNContainer(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	name := acc.GetRandName("fqdn_container")
	config := func(stage int) string {
		values := `["example.com","api.example.com"]`
		if stage == 1 {
			values = `["example.org"]`
		}
		if stage == 2 {
			values = `[]`
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_fqdn_container" "test" {
 name=%q
 description=%q
 mode="INLINE"
 values=%s
}`, name, fmt.Sprintf("description-%d", stage), values)
	}
	object := "cato_fqdn_container.test"
	var ownedID, removedID string
	capture := func(state *terraform.State) error {
		ownedID = state.RootModule().Resources[object].Primary.Attributes["id"]
		return nil
	}
	steps := []resource.TestStep{}
	for i := 0; i < 3; i++ {
		steps = append(steps, resource.TestStep{Config: config(i), Check: checkMembers(object, i)}, resource.TestStep{Config: config(i), PlanOnly: true, ExpectNonEmptyPlan: false}, resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true})
	}
	steps = acc.WithStableResourceID(steps, object, "id", capture)
	steps = append(steps,
		resource.TestStep{PreConfig: func() {
			require.NotEmpty(t, ownedID)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			client := acc.GetClient(t)
			found, err := client.ContainerList(ctx, acc.CatoAccountID, m.ContainerSearchInput{
				Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID, Input: ownedID}}, Types: []m.ContainerType{m.ContainerTypeFqdn}})
			require.NoError(t, err)
			items := found.GetContainer().GetList().GetContainers()
			require.Len(t, items, 1)
			require.Equal(t, ownedID, items[0].GetID())
			require.Equal(t, name, items[0].GetName())
			deleted := new(cato.ContainerDelete)
			err = client.Client.Post(ctx, "containerDelete", `mutation containerDelete($accountId:ID!,$deleteContainerInput:DeleteContainerInput!){container(accountId:$accountId){delete(input:$deleteContainerInput){container{id}}}}`, deleted,
				map[string]interface{}{"accountId": acc.CatoAccountID, "deleteContainerInput": m.DeleteContainerInput{Ref: &m.ContainerRefInput{By: m.ObjectRefByID, Input: ownedID}}})
			require.NoError(t, err)
			require.Equal(t, ownedID, deleted.GetContainer().GetDelete().GetContainer().GetID())
			removedID = ownedID
		}, Config: config(2), PlanOnly: true, ExpectNonEmptyPlan: true},
		resource.TestStep{Config: config(0), Check: resource.ComposeTestCheckFunc(checkMembers(object, 0), func(state *terraform.State) error {
			id := state.RootModule().Resources[object].Primary.Attributes["id"]
			if id == "" || id == removedID {
				return fmt.Errorf("externally deleted container was not recreated with a new identity")
			}
			return nil
		})},
		resource.TestStep{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
		resource.TestStep{ResourceName: object, ImportState: true, ImportStateVerify: true},
	)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: steps, CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_fqdn_container" || item.Primary == nil {
				continue
			}
			result, err := acc.GetClient(t).ContainerList(context.Background(), acc.CatoAccountID, m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID, Input: item.Primary.ID}}, Types: []m.ContainerType{m.ContainerTypeFqdn}})
			if err != nil {
				return err
			}
			if result.GetContainer().GetList() == nil || result.GetContainer().GetList().Containers == nil || len(result.GetContainer().GetList().Containers) != 0 {
				return fmt.Errorf("container absence not confirmed")
			}
		}
		return nil
	}})
}

// The metadata size can lag a clear; downloaded members determine membership.
func checkMembers(object string, stage int) resource.TestCheckFunc {
	count := []string{"2", "1", "0"}[stage]
	checks := []resource.TestCheckFunc{resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "values.#", count), resource.TestCheckResourceAttrSet(object, "size")}

	if stage == 0 {
		checks = append(checks, resource.TestCheckTypeSetElemAttr(object, "values.*", "example.com"), resource.TestCheckTypeSetElemAttr(object, "values.*", "api.example.com"))
	}
	if stage == 1 {
		checks = append(checks, resource.TestCheckTypeSetElemAttr(object, "values.*", "example.org"))
	}

	return resource.ComposeTestCheckFunc(checks...)
}

// Opt-in: synchronization can send notifications to the explicitly supplied disposable destination.
func TestAccFQDNContainerURL(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	source := os.Getenv("TFACC_FQDN_CONTAINER_URL")
	destination := os.Getenv("TFACC_CONTAINER_NOTIFICATION_MAILING_LIST_ID")
	if os.Getenv("TFACC_CONTAINER_URL_ALLOW_NOTIFICATIONS") != "1" || source == "" || destination == "" {
		t.Skip("URL mode requires an HTTPS source, disposable mailing-list ID, and TFACC_CONTAINER_URL_ALLOW_NOTIFICATIONS=1")
	}
	name := acc.GetRandName("fqdn_container_url")
	object := "cato_fqdn_container.test"
	config := func(updated bool) string {
		interval := 1
		if updated {
			interval = 2
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_fqdn_container" "test" {
 name=%q
 description=%q
 mode="URL"
 file_type="CSV"
 sync_data={
 url=%q
 time_interval=%d
 time_unit="DAY"
 notifications={mailing_list=[{id=%q}],subscription_group=[],webhook=[]}
 }
 }`, name, fmt.Sprintf("URL metadata %d", interval), source, interval, destination)
	}
	steps := []resource.TestStep{
		{Config: config(false), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "mode", "URL"), resource.TestCheckResourceAttr(object, "file_type", "CSV"), resource.TestCheckResourceAttr(object, "sync_data.url", source), resource.TestCheckNoResourceAttr(object, "values.#"))},
		{Config: config(false), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"size"}},
		{Config: config(true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "sync_data.time_interval", "2"), resource.TestCheckNoResourceAttr(object, "values.#"))},
		{Config: config(true), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"size"}},
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: acc.WithStableResourceID(steps, object, "id"), CheckDestroy: func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "cato_fqdn_container" || item.Primary == nil {
				continue
			}
			result, err := acc.GetClient(t).ContainerList(context.Background(), acc.CatoAccountID, m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID, Input: item.Primary.ID}}, Types: []m.ContainerType{m.ContainerTypeFqdn}})
			if err != nil {
				return err
			}
			if result.GetContainer().GetList() == nil || result.GetContainer().GetList().Containers == nil || len(result.GetContainer().GetList().Containers) != 0 {
				return fmt.Errorf("container absence not confirmed")
			}
		}
		return nil
	}})
}
