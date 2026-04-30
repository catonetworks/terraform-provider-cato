//go:build acctest

package acc

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

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
)

// env. variables
const (
	envCatoToken     = "CATO_TOKEN"
	envCatoAccountID = "CATO_ACCOUNT_ID"
	envCatoEndpoint  = "CATO_BASEURL"
)

var (
	CatoAccountID = os.Getenv(envCatoAccountID)
	CatoToken     = os.Getenv(envCatoToken)
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

type entityResp struct {
	Data   entityData  `json:"data"`
	Errors []respError `json:"errors"`
}
type respError struct {
	Msg  string   `json:"message"`
	Path []string `json:"path"`
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

type Ref struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type TestLocations []Ref
type TestUsers []Ref
type TestPrivateApps []Ref
type TestConnectorGroups []string
type TestDhcpRelayGroups []Ref
type TestHosts []Ref
type TestAdvancedGroups []Ref

var (
	catoClient          *cato.Client
	catoLocations       TestLocations
	catoConnectorGroups TestConnectorGroups
	catoUsers           TestUsers
	catoPrivateApps     TestPrivateApps
	catoDhcpRelayGroups TestDhcpRelayGroups
	catoHosts           TestHosts
	catoAdvancedGroups  TestAdvancedGroups

	mu  sync.Mutex
	ctx = context.Background()
)

func GetRandName(resource string) string {
	if accmock.ACCMockActive {
		return "test_" + resource
	}
	const length = 10
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[r.Intn(len(charset))]
	}
	return "acctest_" + resource + "_" + string(bytes)
}

func CheckCMAVars(t *testing.T) func() {
	return func() {
		for _, envVar := range []string{envCatoToken, envCatoAccountID, envCatoEndpoint} {
			if os.Getenv(envVar) == "" {
				t.Fatalf("ERROR: env variable '%s' not set", envVar)
			}
		}
	}
}

func GetRandIP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	return fmt.Sprintf("10.%d.%d.%d", 2+r.Intn(252), 2+r.Intn(252), 2+r.Intn(252))
}

func PrintAttributes(resource string) func(st *terraform.State) error {
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

func GetClient(t *testing.T) *cato.Client {
	mu.Lock()
	defer mu.Unlock()
	if catoClient == nil {
		newCatoClient, err := cato.New(os.Getenv(envCatoEndpoint), CatoToken, CatoAccountID, nil,
			map[string]string{"User-Agent": "cato-terraform-test"})
		if err != nil {
			t.Fatalf("ERROR creating cato client: %s", err)
		}
		catoClient = newCatoClient
		t.Logf("Connected to Cato API")
	}
	return catoClient
}

func GetConnectorGroups(t *testing.T) TestConnectorGroups {
	const testAppConn1 = "acctest_app_connector_1"
	const testAppConnGroup1 = "acctest_app_connector_group_1"
	client := GetClient(t)
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

func GetPrivateApps(t *testing.T) TestPrivateApps {
	const privateAppName1 = "acctest_private_app_1"
	const privateAppName2 = "acctest_private_app_2"
	client := GetClient(t)
	mu.Lock()
	defer mu.Unlock()
	if catoPrivateApps == nil {
		// try to fetch private apps
		for _, paName := range []string{privateAppName1, privateAppName2} {
			readInput := cato_models.PrivateApplicationRefInput{By: cato_models.ObjectRefByName, Input: paName}
			result, err := client.PrivateAppReadPrivateApp(ctx, CatoAccountID, readInput)
			if err == nil {
				pa := result.GetPrivateApplication().GetPrivateApplication()
				if pa.GetID() != "" {
					catoPrivateApps = append(catoPrivateApps, Ref{ID: pa.GetID(), Name: pa.GetName()})
					continue
				}
			}
			// create the app
			input := cato_models.CreatePrivateApplicationInput{
				Description:        ptr(paName + " description"),
				InternalAppAddress: GetRandIP(),
				Name:               paName,
			}
			res, err := client.PrivateAppCreatePrivateApp(ctx, CatoAccountID, input)
			if err != nil {
				t.Fatalf("ERROR creating test private app: %v", err)
			}
			pa := res.GetPrivateApplication().GetCreatePrivateApplication().GetApplication()
			catoPrivateApps = append(catoPrivateApps, Ref{ID: pa.GetID(), Name: paName})
		}
	}

	return catoPrivateApps
}

func PublishPrivateAccessPolicy(t *testing.T) {
	client := GetClient(t)
	_, err := client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID)
	if err != nil {
		t.Fatalf("Cato API PolicyPrivateAccessPublishRevision error: %v", err.Error())
	}
}

func GetAdvancedGroups(t *testing.T) TestAdvancedGroups {
	const groupName1 = "acctest_advanced_group_000"
	const groupName2 = "acctest_advanced_group_001"
	const groupRE = "^acctest_advanced_group_00[0-1]$"
	client := GetClient(t)
	mu.Lock()
	defer mu.Unlock()
	if catoAdvancedGroups == nil {
		// try to fetch groups
		groupsInput := &cato_models.GroupListInput{
			Filter: []*cato_models.GroupListFilterInput{
				{Name: []*cato_models.AdvancedStringFilterInput{{Regex: ptr(groupRE)}}},
			},
			Paging: &cato_models.PagingInput{From: 0, Limit: 10},
			Sort:   &cato_models.GroupListSortInput{Name: &cato_models.SortOrderInput{Direction: cato_models.SortOrderAsc}},
		}
		result, err := client.GroupsList(ctx, groupsInput, CatoAccountID)
		var groupsFound []string
		if err != nil {
			t.Fatalf("failed to load groups: %v", err)
			return nil
		}

		items := result.GetGroups().GetGroupList().GetItems()
		for _, item := range items {
			catoAdvancedGroups = append(catoAdvancedGroups, Ref{ID: item.GetID(), Name: item.GetName()})
			groupsFound = append(groupsFound, item.GetName())
		}
		for _, groupName := range []string{groupName1, groupName2} {
			if slices.Contains(groupsFound, groupName) {
				continue
			}
			// create the group
			createGroupInput := cato_models.CreateGroupInput{Name: groupName, Description: ptr(groupName + " terraform tests")}
			res, err := client.GroupsCreateGroup(ctx, createGroupInput, CatoAccountID)
			if err != nil {
				t.Fatalf("ERROR creating test group: %v", err)
			}
			catoAdvancedGroups = append(catoAdvancedGroups, Ref{ID: res.GetGroups().GetCreateGroup().GetGroup().GetID(), Name: groupName})
		}
	}

	return catoAdvancedGroups
}

func ProviderCfg() string {
	return fmt.Sprintf("provider \"cato\" {\n  account_id = %q\n}\n", CatoAccountID)
}

// getEntities is a generic function to call entityLookup API and return a []Ref (name and ID of the entities)
func getEntities(t *testing.T, entityType string) (refs []Ref) {
	var res entityResp
	query := `{"query": "query entityLookup ($accountID:ID! $type:EntityType!) ` +
		`{entityLookup (accountID:$accountID type:$type) {items {entity {id name}}}}",
			"variables": {"accountID": "` + CatoAccountID + `","type": "` + entityType + `"},
			"operationName": "entityLookup"}`

	// Create request
	req, err := http.NewRequest(http.MethodPost, os.Getenv(envCatoEndpoint), strings.NewReader(query)) //nolint:gosec
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", CatoToken)
	client := &http.Client{}
	resp, err := client.Do(req) //nolint:gosec
	if err != nil {
		t.Fatalf("ERROR fetching %q: %v", entityType, err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			t.Logf("ERROR closing response body: %v", errClose)
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ERROR reading %q: %v", entityType, err)
	}
	if err = json.Unmarshal(body, &res); err != nil {
		t.Fatalf("ERROR unmarshalling response: %v", err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("ERROR: cannot fetch %q: %v", entityType, res.Errors)
	}
	if len(res.Data.EntityLookup.Items) == 0 {
		t.Fatalf("ERROR: failed to fetch %q: res.Data.EntityLookup.Items is empty", entityType)
	}
	for _, u := range res.Data.EntityLookup.Items {
		refs = append(refs, Ref{ID: u.Entity.ID, Name: u.Entity.Name})
	}

	return refs
}

func GetUsers(t *testing.T) TestUsers {
	mu.Lock()
	defer mu.Unlock()
	if catoUsers == nil {
		if cmaVars.Users != nil {
			return cmaVars.Users
		}
		catoUsers = getEntities(t, "vpnUser")
	}

	return catoUsers
}

func GetDhcpRelayGroups(t *testing.T) TestDhcpRelayGroups {
	mu.Lock()
	defer mu.Unlock()
	if catoDhcpRelayGroups == nil {
		if cmaVars.Users != nil {
			return cmaVars.Users
		}
		catoDhcpRelayGroups = getEntities(t, "dhcpRelayGroup")
	}

	return catoDhcpRelayGroups
}

func GetLocations(t *testing.T) TestLocations {
	mu.Lock()
	defer mu.Unlock()
	if catoLocations == nil {
		catoLocations = getEntities(t, "location")
	}
	return catoLocations
}

func GetHosts(t *testing.T) TestHosts {
	mu.Lock()
	defer mu.Unlock()
	if catoHosts == nil {
		catoHosts = getEntities(t, "host")
	}
	return catoHosts
}

func ptr[T any](x T) *T { return &x }
