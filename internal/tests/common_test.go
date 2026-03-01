package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// env. variables
const (
	envCatoToken     = "CATO_TOKEN"
	envCatoAccountID = "CATO_ACCOUNT_ID"
	envCatoEndpoint  = "CATO_ENDPOINT"
)

var (
	CatoAccountID = os.Getenv(envCatoAccountID)
	CatoToken     = os.Getenv(envCatoToken)
	CatoEndpoint  = os.Getenv(envCatoEndpoint)
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type ref struct {
	Name string
	ID   string
}
type testLocations []ref
type testConnectorGroups []string

var (
	catoClient          *cato.Client
	catoLocations       testLocations
	catoConnectorGroups testConnectorGroups
	mu                  sync.Mutex
	ctx                 = context.Background()
)

func getRandName(resource string) string {
	const length = 10
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[r.Intn(len(charset))]
	}
	return "test_" + resource + "_" + string(bytes)
}

func checkCMAVars(t *testing.T) func() {
	return func() {
		for _, envVar := range []string{envCatoToken, envCatoAccountID, envCatoEndpoint} {
			if os.Getenv(envVar) == "" {
				t.Fatalf("ERROR: env variable '%s' not set", envVar)
			}
		}
	}
}

func getRandIP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("10.%d.%d.%d", 2+r.Intn(252), 2+r.Intn(252), 2+r.Intn(252))
}

func printAttributes(resource string) func(st *terraform.State) error {
	return func(st *terraform.State) error {
		attrs := st.Modules[0].Resources[resource].Primary.Attributes
		keys := make([]string, 0, len(attrs))
		for k := range attrs {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		fmt.Printf("Resource attributes (%s):\n", resource)
		for _, k := range keys {
			fmt.Printf("\t%s: %s\n", k, attrs[k])
		}
		return nil
	}
}

func getClient(t *testing.T) *cato.Client {
	mu.Lock()
	defer mu.Unlock()
	if catoClient == nil {
		newCatoClient, err := cato.New(CatoEndpoint, CatoToken, CatoAccountID, nil,
			map[string]string{"User-Agent": "cato-terraform-test"})
		if err != nil {
			t.Fatalf("ERROR creating cato client: %s", err)
		}
		catoClient = newCatoClient
		t.Logf("Connected to Cato API")
	}
	return catoClient
}

func getLocations(t *testing.T) testLocations {
	client := getClient(t)
	mu.Lock()
	defer mu.Unlock()
	if catoLocations == nil {
		const maxItems = 5
		var newLoc testLocations
		resp, err := client.EntityLookup(ctx, CatoAccountID, cato_models.EntityTypeLocation,
			ptr(int64(5)), ptr(int64(0)), nil, ptr(""), nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("ERROR getting locations: %s", err)
		}
		items := resp.GetEntityLookup().GetItems()
		if len(items) < maxItems {
			t.Fatal("ERROR getting locations: items < 5")
		}
		for i, item := range items {
			if i == maxItems {
				break
			}
			newLoc = append(newLoc, ref{Name: *item.Entity.Name, ID: item.Entity.ID})
		}
		catoLocations = newLoc
	}
	return catoLocations
}
func getConnectorGroups(t *testing.T) testConnectorGroups {
	const testAppConn1 = "acctest_app_connector_1"
	const testAppConnGroup1 = "acctest_app_connector_group_1"
	client := getClient(t)
	mu.Lock()
	defer mu.Unlock()
	if catoConnectorGroups == nil {
		// try to fetch appConnectors

		input := cato_models.ZtnaAppConnectorRefInput{By: cato_models.ObjectRefByName, Input: testAppConn1}
		result, err := client.AppConnectorReadConnector(ctx, CatoAccountID, input)
		if err == nil {
			group := result.GetZtnaAppConnector().GetZtnaAppConnector().GetGroupName()
			if group == "" {
				t.Fatal("ERROR getting app-connector group: group is empty")
			}
			catoConnectorGroups = []string{group} // TODO: is 1 group enough?
			return catoConnectorGroups
		}
	}

	// create a new app-connector
	input := cato_models.AddZtnaAppConnectorInput{
		GroupName: testAppConnGroup1,
		Name:      testAppConn1,
		Type:      cato_models.ZtnaAppConnectorTypeVirtual,
		Location: &cato_models.ZtnaAppConnectorLocationInput{
			City:        "Prague",
			CountryCode: "CZ",
			Timezone:    "Europe/Prague",
		},
		PreferredPopLocation: &cato_models.ZtnaAppConnectorPreferredPopLocationInput{
			Automatic: true,
		},
	}
	_, err := client.AppConnectorCreateConnector(ctx, CatoAccountID, input)
	if err != nil {
		t.Fatalf("Cato API AppConnectorCreateConnector error: %v", err.Error())
	}
	catoConnectorGroups = []string{testAppConnGroup1} // TODO: is 1 group enough?

	return catoConnectorGroups
}

func providerCfg() string {
	return fmt.Sprintf("provider \"cato\" {\n  account_id = \"%s\"\n}\n", CatoAccountID)
}

func ptr[T any](x T) *T { return &x }
