package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"strings"
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

type entityResp struct {
	Data entityData `json:"data"`
}
type entityData struct {
	EntityLookup entityLookup `json:"entityLookup"`
}
type entityLookup struct {
	Items []entityItem `json:"items"`
}
type entityItem struct {
	Entity entityDetail `json:"entity"`
}
type entityDetail struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ref struct {
	Name string
	ID   string
}
type testLocations []ref
type testUsers []ref
type testPrivateApps []ref
type testConnectorGroups []string

var (
	catoClient          *cato.Client
	catoLocations       testLocations
	catoConnectorGroups testConnectorGroups
	catoUsers           testUsers
	catoPrivateApps     testPrivateApps
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

func getUsers(t *testing.T) testUsers {
	mu.Lock()
	defer mu.Unlock()
	if catoUsers == nil {
		// try to fetch users
		// result, err := client.EntityLookup(ctx, CatoAccountID, cato_models.EntityTypeVpnUser, ptr(int64(5)), ptr(int64(0)), nil, nil, nil, nil, nil, nil)

		var res entityResp
		query := `{"query": "query entityLookup ($accountID:ID! $type:EntityType!) {entityLookup (accountID:$accountID type:$type) {items {entity {id name}}}}",
			"variables": {"accountID": "` + CatoAccountID + `","type": "vpnUser"},
			"operationName": "entityLookup"}`

		// Create request
		req, err := http.NewRequest(http.MethodPost, CatoEndpoint, strings.NewReader(query))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", CatoToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("ERROR fetching users: %v", err)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ERROR reading users: %v", err)
		}
		if err = json.Unmarshal(body, &res); err != nil {
			t.Fatalf("ERROR unmarshalling response: %v", err)
		}
		for _, u := range res.Data.EntityLookup.Items {
			catoUsers = append(catoUsers, ref{ID: u.Entity.ID, Name: u.Entity.Name})
		}
	}

	return catoUsers
}

func getPrivateApps(t *testing.T) testPrivateApps {
	const privateAppName1 = "acctest_private_app_1"
	const privateAppName2 = "acctest_private_app_2"
	// client := getClient(t)
	mu.Lock()
	defer mu.Unlock()
	if catoPrivateApps == nil {
		// TODO: enable when API gets fixed!
		return []ref{{ID: "219", Name: "acctest_private_app_1"}, {ID: "220", Name: "acctest_private_app_2"}}
	}

	/*
			// try to fetch private apps
			for _, paName := range []string{privateAppName1, privateAppName2} {
				readInput := cato_models.PrivateApplicationRefInput{By: cato_models.ObjectRefByName, Input: paName}
				result, err := client.PrivateAppReadPrivateApp(ctx, CatoAccountID, readInput)
				if err == nil {
					pa := result.GetPrivateApplication().GetPrivateApplication()
					if pa.GetID() != "" {
						catoPrivateApps = append(catoPrivateApps, ref{ID: pa.GetID(), Name: pa.GetName()})
						continue
					}
				}
				// create the app
				input := cato_models.CreatePrivateApplicationInput{
					Description:        ptr(paName + " description"),
					InternalAppAddress: getRandIP(),
					Name:               paName,
				}
				res, err := client.PrivateAppCreatePrivateApp(ctx, CatoAccountID, input)
				if err != nil {
					t.Fatalf("ERROR creating test private app: %v", err)
				}
				pa := res.GetPrivateApplication().GetCreatePrivateApplication().GetApplication()
				catoPrivateApps = append(catoPrivateApps, ref{ID: pa.GetID(), Name: paName})
			}
		}

	*/
	return catoPrivateApps
}

func publishPrivateAccessPolicy(t *testing.T) {
	client := getClient(t)
	_, err := client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID)
	if err != nil {
		t.Fatalf("Cato API PolicyPrivateAccessPublishRevision error: %v", err.Error())
	}
}

func providerCfg() string {
	return fmt.Sprintf("provider \"cato\" {\n  account_id = \"%s\"\n}\n", CatoAccountID)
}

func ptr[T any](x T) *T { return &x }
