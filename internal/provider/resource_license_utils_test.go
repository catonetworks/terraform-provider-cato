package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUpsertLicense_AcceptsSitesAtAndBeyondFirstThousand(t *testing.T) {
	t.Parallel()

	for _, siteID := range []string{"1", "1000", "1001", "10000"} {
		siteID := siteID
		t.Run("site_"+siteID+"_is_supported", func(t *testing.T) {
			t.Parallel()

			fixture := newLicenseFixture(t, licenseFixture{
				ID:           "lic-1",
				SKU:          "CATO_SITE",
				AssignedSite: siteID,
			})
			defer fixture.Close()

			_, err := fixture.upsert(siteID, "lic-1", types.Int64Null())
			if err != nil {
				t.Fatalf("expected site %s to be accepted, got error: %v", siteID, err)
			}
		})
	}
}

func TestUpsertLicense_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		licenses []licenseFixture
		siteID   string
		license  string
		wantErr  string
	}{
		"no_licenses": {
			siteID:  "10",
			license: "lic-1",
			wantErr: "LICENSE API ERROR: No licenses were found on this account.",
		},
		"license_id_missing": {
			licenses: []licenseFixture{{ID: "lic-2", SKU: "CATO_SITE"}},
			siteID:   "10",
			license:  "lic-1",
			wantErr:  "INVALID CONFIGURATION: License ID 'lic-1' not found.",
		},
		"unsupported_sku": {
			licenses: []licenseFixture{{ID: "lic-1", SKU: "CATO_ZTNA_USERS"}},
			siteID:   "10",
			license:  "lic-1",
			wantErr:  "INVALID LICENSE TYPE: Site License ID 'lic-1' is not a valid site license.",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fixture := newLicenseFixture(t, tt.licenses...)
			defer fixture.Close()

			_, err := fixture.upsert(tt.siteID, tt.license, types.Int64Null())
			assertErrorContains(t, err, tt.wantErr)
			assertMutations(t, fixture)
		})
	}
}

func TestUpsertLicense_StaticSiteLicense(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		license  licenseFixture
		siteID   string
		bw       types.Int64
		wantOps  []string
		wantErr  string
		wantSite string
	}{
		"already_assigned_to_same_site": {
			license: licenseFixture{ID: "lic-1", SKU: "CATO_SITE", AssignedSite: "10"},
			siteID:  "10",
		},
		"already_assigned_to_other_site": {
			license: licenseFixture{ID: "lic-1", SKU: "CATO_SITE", AssignedSite: "11"},
			siteID:  "10",
			wantErr: "LICENSE ALREADY ASSIGNED: The license ID 'lic-1' is already assigned to site ID",
		},
		"assigns_unassigned_license": {
			license:  licenseFixture{ID: "lic-1", SKU: "CATO_SITE"},
			siteID:   "10",
			wantOps:  []string{"assignSiteBwLicense"},
			wantSite: "10",
		},
		"rejects_bandwidth": {
			license: licenseFixture{ID: "lic-1", SKU: "CATO_SITE"},
			siteID:  "10",
			bw:      types.Int64Value(100),
			wantErr: "INVALID CONFIGURATION: Bandwidth is not supported for 'CATO_SITE'",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bw := tt.bw
			if bw.IsNull() && !bw.IsUnknown() && bw.ValueInt64() == 0 {
				bw = types.Int64Null()
			}
			fixture := newLicenseFixture(t, tt.license)
			defer fixture.Close()

			_, err := fixture.upsert(tt.siteID, "lic-1", bw)
			if tt.wantErr != "" {
				assertErrorContains(t, err, tt.wantErr)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertMutations(t, fixture, tt.wantOps...)
			if tt.wantSite != "" {
				assertMutationInput(t, fixture.mutations()[0], "site.input", tt.wantSite)
			}
		})
	}
}

func TestUpsertLicense_PooledBandwidthLicense(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		licenses []licenseFixture
		siteID   string
		bw       types.Int64
		wantOps  []string
		wantErr  string
	}{
		"assigns_unassigned_license_with_bandwidth": {
			licenses: []licenseFixture{{ID: "lic-1", SKU: "CATO_PB"}},
			siteID:   "10",
			bw:       types.Int64Value(100),
			wantOps:  []string{"assignSiteBwLicense"},
		},
		"assign_requires_bandwidth": {
			licenses: []licenseFixture{{ID: "lic-1", SKU: "CATO_PB"}},
			siteID:   "10",
			bw:       types.Int64Null(),
			wantErr:  "INVALID CONFIGURATION: Bandwidth must be set for 'CATO_PB'",
		},
		"replaces_existing_assignment": {
			licenses: []licenseFixture{
				{ID: "lic-old", SKU: "CATO_PB", PooledSites: []siteAllocation{{SiteID: "10", BW: 50}}},
				{ID: "lic-1", SKU: "CATO_PB"},
			},
			siteID:  "10",
			bw:      types.Int64Value(100),
			wantOps: []string{"replaceSiteBwLicense"},
		},
		"updates_changed_bandwidth": {
			licenses: []licenseFixture{{ID: "lic-1", SKU: "CATO_PB", PooledSites: []siteAllocation{{SiteID: "10", BW: 50}}}},
			siteID:   "10",
			bw:       types.Int64Value(100),
			wantOps:  []string{"updateSiteBwLicense"},
		},
		"same_bandwidth_is_noop": {
			licenses: []licenseFixture{{ID: "lic-1", SKU: "CATO_PB", PooledSites: []siteAllocation{{SiteID: "10", BW: 100}}}},
			siteID:   "10",
			bw:       types.Int64Value(100),
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fixture := newLicenseFixture(t, tt.licenses...)
			defer fixture.Close()

			_, err := fixture.upsert(tt.siteID, "lic-1", tt.bw)
			if tt.wantErr != "" {
				assertErrorContains(t, err, tt.wantErr)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertMutations(t, fixture, tt.wantOps...)
			if len(tt.wantOps) > 0 {
				assertMutationInput(t, fixture.mutations()[0], "site.input", tt.siteID)
			}
			if !tt.bw.IsNull() && !tt.bw.IsUnknown() && len(tt.wantOps) > 0 {
				assertMutationInput(t, fixture.mutations()[0], "bw", float64(tt.bw.ValueInt64()))
			}
		})
	}
}

type licenseTestFixture struct {
	t        *testing.T
	server   *httptest.Server
	licenses []licenseFixture

	mu            sync.Mutex
	mutationCalls []mutationCall
}

type licenseFixture struct {
	ID           string
	SKU          string
	AssignedSite string
	PooledSites  []siteAllocation
}

type siteAllocation struct {
	SiteID string
	BW     int64
}

type mutationCall struct {
	Operation string
	Input     map[string]any
}

func newLicenseFixture(t *testing.T, licenses ...licenseFixture) *licenseTestFixture {
	t.Helper()

	fixture := &licenseTestFixture{
		t:        t,
		licenses: licenses,
	}
	fixture.server = httptest.NewServer(http.HandlerFunc(fixture.handleGraphQL))
	return fixture
}

func (f *licenseTestFixture) Close() {
	f.server.Close()
}

func (f *licenseTestFixture) upsert(siteID, licenseID string, bw types.Int64) (*cato.Licensing_Licensing_LicensingInfo_Licenses, error) {
	f.t.Helper()

	client, err := cato.New(f.server.URL, "test-token", "3381", nil, nil)
	if err != nil {
		f.t.Fatalf("failed to create cato client: %v", err)
	}

	plan := LicenseResource{
		SiteID:      types.StringValue(siteID),
		LicenseID:   types.StringValue(licenseID),
		BW:          bw,
		LicenseInfo: types.ObjectNull(LicenseInfoResourceAttrTypes),
	}

	return upsertLicense(context.Background(), plan, &catoClientData{
		AccountId: "3381",
		catov2:    client,
	})
}

func (f *licenseTestFixture) mutations() []mutationCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]mutationCall(nil), f.mutationCalls...)
}

func (f *licenseTestFixture) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var body struct {
		OperationName string         `json:"operationName"`
		Variables     map[string]any `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		f.t.Fatalf("failed to decode request body: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	switch body.OperationName {
	case "entityLookup":
		writeJSON(f.t, w, map[string]any{
			"data": map[string]any{
				"entityLookup": map[string]any{"items": siteLookupItems(body.Variables)},
			},
		})
	case "licensing":
		writeJSON(f.t, w, licensingResponse(f.licenses))
	case "assignSiteBwLicense", "replaceSiteBwLicense", "updateSiteBwLicense":
		f.recordMutation(body.OperationName, body.Variables)
		writeJSON(f.t, w, map[string]any{
			"data": map[string]any{
				"sites": map[string]any{body.OperationName: nil},
			},
		})
	default:
		f.t.Fatalf("unexpected operationName: %q", body.OperationName)
	}
}

func (f *licenseTestFixture) recordMutation(operation string, variables map[string]any) {
	input, _ := variables["input"].(map[string]any)

	f.mu.Lock()
	defer f.mu.Unlock()
	f.mutationCalls = append(f.mutationCalls, mutationCall{
		Operation: operation,
		Input:     input,
	})
}

func siteLookupItems(variables map[string]any) []map[string]any {
	if entityIDs, ok := entityIDsVariable(variables); ok {
		items := make([]map[string]any, 0, len(entityIDs))
		for _, siteID := range entityIDs {
			if siteExists(siteID) {
				items = append(items, siteLookupItem(siteID))
			}
		}
		return items
	}

	const totalSites = 10001

	from := intVariable(variables, "from")
	if from < 0 {
		from = 0
	}
	if from >= totalSites {
		return nil
	}

	limit := intVariable(variables, "limit")
	if limit <= 0 || limit > totalSites {
		limit = totalSites
	}

	end := from + limit
	if end > totalSites {
		end = totalSites
	}

	items := make([]map[string]any, 0, end-from)
	for i := from + 1; i <= end; i++ {
		items = append(items, siteLookupItem(fmt.Sprintf("%d", i)))
	}
	return items
}

func siteLookupItem(siteID string) map[string]any {
	return map[string]any{
		"entity": map[string]any{"id": siteID, "name": "site-" + siteID},
	}
}

func intVariable(variables map[string]any, name string) int {
	if value, ok := variables[name].(float64); ok {
		return int(value)
	}
	if value, ok := variables[name].(int); ok {
		return value
	}
	return 0
}

func entityIDsVariable(variables map[string]any) ([]string, bool) {
	if entityIDs, ok := stringSliceVariable(variables, "entityIDs"); ok {
		return entityIDs, true
	}
	return stringSliceVariable(variables, "entityIds")
}

func stringSliceVariable(variables map[string]any, name string) ([]string, bool) {
	values, ok := variables[name].([]any)
	if !ok || len(values) == 0 {
		return nil, false
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		siteID, ok := value.(string)
		if !ok {
			continue
		}
		result = append(result, siteID)
	}
	return result, true
}

func siteExists(siteID string) bool {
	id, err := strconv.Atoi(siteID)
	if err != nil {
		return false
	}
	return id >= 1 && id <= 10001
}

func licensingResponse(licenses []licenseFixture) map[string]any {
	items := make([]map[string]any, 0, len(licenses))
	for _, license := range licenses {
		items = append(items, licenseResponseItem(license))
	}

	return map[string]any{
		"data": map[string]any{
			"licensing": map[string]any{
				"licensingInfo": map[string]any{
					"licenses": items,
					"globalLicenseAllocations": map[string]any{
						"publicIps": map[string]any{"total": 0, "allocated": 0, "available": 0},
						"ztnaUsers": map[string]any{"total": 0, "allocated": 0, "available": 0},
					},
				},
			},
		},
	}
}

func licenseResponseItem(license licenseFixture) map[string]any {
	item := map[string]any{
		"id":             license.ID,
		"sku":            license.SKU,
		"plan":           "COMMERCIAL",
		"status":         "ACTIVE",
		"expirationDate": "2030-01-01",
		"startDate":      "2024-01-01",
		"lastUpdated":    "2024-01-02",
	}

	switch license.SKU {
	case "CATO_SITE", "CATO_SSE_SITE":
		item["siteLicenseGroup"] = "GROUP_1"
		item["regionality"] = "GLOBAL"
		item["siteLicenseType"] = "SASE"
		item["total"] = 1
		if license.AssignedSite != "" {
			item["site"] = siteObject(license.AssignedSite)
		}
	case "CATO_PB", "CATO_PB_SSE":
		item["siteLicenseGroup"] = "GROUP_1"
		item["siteLicenseType"] = "SASE"
		item["total"] = 1000
		item["allocatedBandwidth"] = int64(0)
		sites := make([]map[string]any, 0, len(license.PooledSites))
		for _, site := range license.PooledSites {
			sites = append(sites, map[string]any{
				"allocatedBandwidth":             site.BW,
				"sitePooledBandwidthLicenseSite": siteObject(site.SiteID),
			})
			item["allocatedBandwidth"] = item["allocatedBandwidth"].(int64) + site.BW
		}
		item["sites"] = sites
		item["accounts"] = []map[string]any{}
	default:
		item["total"] = 1
	}

	return item
}

func siteObject(siteID string) map[string]any {
	return map[string]any{"id": siteID, "name": "site-" + siteID}
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %q", want, err.Error())
	}
}

func assertMutations(t *testing.T, fixture *licenseTestFixture, wantOps ...string) {
	t.Helper()

	mutations := fixture.mutations()
	if len(mutations) != len(wantOps) {
		t.Fatalf("expected mutations %v, got %+v", wantOps, mutations)
	}
	for i, want := range wantOps {
		if mutations[i].Operation != want {
			t.Fatalf("expected mutation %d to be %q, got %q", i, want, mutations[i].Operation)
		}
	}
}

func assertMutationInput(t *testing.T, mutation mutationCall, path string, want any) {
	t.Helper()

	got, err := lookupMutationInput(mutation.Input, strings.Split(path, "."))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("expected mutation input %s=%v, got %v", path, want, got)
	}
}

func lookupMutationInput(value any, path []string) (any, error) {
	if len(path) == 0 {
		return value, nil
	}

	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object while resolving %q", strings.Join(path, "."))
	}
	next, ok := object[path[0]]
	if !ok {
		return nil, errors.New("missing mutation input path " + strings.Join(path, "."))
	}
	return lookupMutationInput(next, path[1:])
}
