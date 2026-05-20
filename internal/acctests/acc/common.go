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
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
)

// env. variables
const (
	envCatoToken     = "CATO_TOKEN"
	envCatoAccountID = "CATO_ACCOUNT_ID"
	envCatoEndpoint  = "CATO_BASEURL"
)

// named references to resources
const (
	// defined in TFACC_TEST_VARS
	resPrivateApps      = "private_apps"
	resAdvancedGroups   = "advanced_groups"
	resGlobalIPRanges   = "global_ip_ranges"
	resFloatingRanges   = "floating_ranges"
	resUserGroups       = "user_groups"
	resSystemGroups     = "system_groups"
	resDevicePostures   = "device_postures"
	resCustomApps       = "custom_apps"
	resCustomCategories = "custom_categories"

	// entity lookup types
	resUsers                   = "vpnUser"
	resDhcpRelayGroup          = "dhcpRelayGroup"
	resLocation                = "location"
	resHost                    = "host"
	resSite                    = "site"
	resNetworkInterface        = "networkInterface"
	resSiteRange               = "siteRange"
	resSubscriptionGroup       = "groupSubscription"
	resWebhookSubscription     = "webhookSubscription"
	resMailingListSubscription = "mailingListSubscription"
)

var (
	CatoAccountID = os.Getenv(envCatoAccountID)
	CatoToken     = os.Getenv(envCatoToken)
	httpClient    = &http.Client{}
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

var (
	catoClient          *cato.Client
	testConnectorGroups []string
	resourceRefs        map[string][]Ref = make(map[string][]Ref)

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

var funcPrefixRE = regexp.MustCompile(`^.*\.`)

func SkipByEnv(t *testing.T) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("could not get caller")
		return
	}

	funcName := funcPrefixRE.ReplaceAllString(runtime.FuncForPC(pc).Name(), "")
	if details, ok := skipTests[funcName]; ok {
		t.Skipf("skipping test '%s': %s", funcName, details)
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

func GetConnectorGroups(t *testing.T) []string {
	const testAppConn1 = "acctest_app_connector_1"
	const testAppConnGroup1 = "acctest_app_connector_group_1"
	client := GetClient(t)
	mu.Lock()
	defer mu.Unlock()
	if testConnectorGroups != nil {
		return testConnectorGroups
	}

	// try to fetch appConnectors
	fetchInput := cato_models.ZtnaAppConnectorRefInput{By: cato_models.ObjectRefByName, Input: testAppConn1}
	result, err := client.AppConnectorReadConnector(ctx, CatoAccountID, fetchInput)
	if err == nil {
		group := result.GetZtnaAppConnector().GetZtnaAppConnector().GetGroupName()
		if group == "" {
			t.Fatalf("ERROR getting app-connector group: group is empty")
		}
		testConnectorGroups = []string{group} // TODO: is 1 group enough?
		return testConnectorGroups
	}

	// nothing fetched, create a new app-connector
	createInput := cato_models.AddZtnaAppConnectorInput{
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
	_, err = client.AppConnectorCreateConnector(ctx, CatoAccountID, createInput)
	if err != nil {
		t.Fatalf("Cato API AppConnectorCreateConnector error: %v", err.Error())
	}
	testConnectorGroups = []string{testAppConnGroup1} // TODO: is 1 group enough?

	return testConnectorGroups
}

func GetPrivateApps(t *testing.T) []Ref {
	const privateAppName1 = "acctest_private_app_1"
	const privateAppName2 = "acctest_private_app_2"
	client := GetClient(t)
	mu.Lock()
	defer mu.Unlock()

	testPrivateApps := resourceRefs[resPrivateApps]
	if testPrivateApps != nil {
		return testPrivateApps
	}

	// try to fetch private apps
	for _, paName := range []string{privateAppName1, privateAppName2} {
		readInput := cato_models.PrivateApplicationRefInput{By: cato_models.ObjectRefByName, Input: paName}
		result, err := client.PrivateAppReadPrivateApp(ctx, CatoAccountID, readInput)
		if err == nil {
			pa := result.GetPrivateApplication().GetPrivateApplication()
			if pa.GetID() != "" {
				testPrivateApps = append(testPrivateApps, Ref{ID: pa.GetID(), Name: pa.GetName()})
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
		testPrivateApps = append(testPrivateApps, Ref{ID: pa.GetID(), Name: paName})
	}
	resourceRefs[resPrivateApps] = testPrivateApps
	return testPrivateApps
}

func PublishPrivateAccessPolicy(t *testing.T) {
	client := GetClient(t)
	_, err := client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID)
	if err != nil {
		t.Fatalf("Cato API PolicyPrivateAccessPublishRevision error: %v", err.Error())
	}
}

func GetAdvancedGroups(t *testing.T) []Ref {
	const groupName1 = "acctest_advanced_group_000"
	const groupName2 = "acctest_advanced_group_001"
	const groupName3 = "acctest_advanced_group_002"
	const groupRE = "^acctest_advanced_group_00[0-2]$"
	client := GetClient(t)
	mu.Lock()
	defer mu.Unlock()

	testAdvancedGroups := resourceRefs[resAdvancedGroups]
	if testAdvancedGroups != nil {
		return testAdvancedGroups
	}

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
		testAdvancedGroups = append(testAdvancedGroups, Ref{ID: item.GetID(), Name: item.GetName()})
		groupsFound = append(groupsFound, item.GetName())
	}

	// check if all groups were found, if not create the missing ones
	for _, groupName := range []string{groupName1, groupName2, groupName3} {
		if slices.Contains(groupsFound, groupName) {
			continue
		}
		// create the group
		createGroupInput := cato_models.CreateGroupInput{Name: groupName, Description: ptr(groupName + " terraform tests")}
		res, err := client.GroupsCreateGroup(ctx, createGroupInput, CatoAccountID)
		if err != nil {
			t.Fatalf("ERROR creating test group: %v", err)
		}
		testAdvancedGroups = append(testAdvancedGroups, Ref{ID: res.GetGroups().GetCreateGroup().GetGroup().GetID(), Name: groupName})
	}

	resourceRefs[resAdvancedGroups] = testAdvancedGroups
	return testAdvancedGroups
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
	resp, err := httpClient.Do(req) //nolint:gosec
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

func getFromVars(t *testing.T, varName string) []Ref {
	mu.Lock()
	defer mu.Unlock()
	refs := resourceRefs[varName]
	if refs == nil {
		refs = cmaVars[varName]
		if len(refs) == 0 {
			t.Fatalf("TFACC_TEST_VARS %s is not set, cannot fetch %s for tests", varName, varName)
			return nil
		}
		resourceRefs[varName] = refs
	}
	return refs
}

func getFromEntityLookup(t *testing.T, lookupName string) []Ref {
	mu.Lock()
	defer mu.Unlock()
	refs := resourceRefs[lookupName]
	if refs == nil {
		refs = getEntities(t, lookupName)
		if len(refs) < 3 {
			t.Fatalf("entity lookup faied for '%s', found just %d items", lookupName, len(refs))
		}
		resourceRefs[lookupName] = refs
	}
	return refs
}

func GetUsers(t *testing.T) []Ref              { return getFromEntityLookup(t, resUsers) }
func GetDhcpRelayGroups(t *testing.T) []Ref    { return getFromEntityLookup(t, resDhcpRelayGroup) }
func GetLocations(t *testing.T) []Ref          { return getFromEntityLookup(t, resLocation) }
func GetHosts(t *testing.T) []Ref              { return getFromEntityLookup(t, resHost) }
func GetSites(t *testing.T) []Ref              { return getFromEntityLookup(t, resSite) }
func GetInterfaces(t *testing.T) []Ref         { return getFromEntityLookup(t, resNetworkInterface) }
func GetSiteRanges(t *testing.T) []Ref         { return getFromEntityLookup(t, resSiteRange) }
func GetSubscriptionGroups(t *testing.T) []Ref { return getFromEntityLookup(t, resSubscriptionGroup) }
func GetWebhooks(t *testing.T) []Ref           { return getFromEntityLookup(t, resWebhookSubscription) }
func GetMailingLists(t *testing.T) []Ref       { return getFromEntityLookup(t, resMailingListSubscription) }

func GetGlobalIPRanges(t *testing.T) []Ref   { return getFromVars(t, resGlobalIPRanges) }
func GetFloatingRanges(t *testing.T) []Ref   { return getFromVars(t, resFloatingRanges) }
func GetUserGroups(t *testing.T) []Ref       { return getFromVars(t, resUserGroups) }
func GetSystemGroups(t *testing.T) []Ref     { return getFromVars(t, resSystemGroups) }
func GetDevicePostures(t *testing.T) []Ref   { return getFromVars(t, resDevicePostures) }
func GetCustomApps(t *testing.T) []Ref       { return getFromVars(t, resCustomApps) }
func GetCustomCategories(t *testing.T) []Ref { return getFromVars(t, resCustomCategories) }

func ptr[T any](x T) *T { return &x }
