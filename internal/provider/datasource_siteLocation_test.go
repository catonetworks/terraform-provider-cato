package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFilterSiteLocations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                  string
		filters               []sldFilters
		expectedSiteLocations []SLDCatalogEntry
	}{
		{
			name: "Filter by city exact match",
			filters: []sldFilters{
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("New York City"),
					Operation: types.StringValue("exact"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{
				{
					City:        "New York City",
					CountryName: "United States",
					CountryCode: "US",
					StateName:   "New York",
					StateCode:   "US-NY",
					Timezone:    []string{"America/New_York"},
				},
			},
		},
		{
			name: "Filter by city startsWith",
			filters: []sldFilters{
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("Los Angeles"),
					Operation: types.StringValue("startsWith"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{
				{
					City:        "Los Angeles",
					CountryName: "Philippines",
					CountryCode: "PH",
					StateName:   "Caraga",
					StateCode:   "",
					Timezone:    []string{"Asia/Manila"},
				},
				{
					City:        "Los Angeles",
					CountryName: "Spain",
					CountryCode: "ES",
					StateName:   "Madrid",
					StateCode:   "",
					Timezone:    []string{"Europe/Madrid"},
				},
				{
					City:        "Los Angeles",
					CountryName: "United States",
					CountryCode: "US",
					StateName:   "California",
					StateCode:   "US-CA",
					Timezone:    []string{"America/Los_Angeles"},
				},
			},
		},
		{
			name: "Filter by state contains",
			filters: []sldFilters{
				{
					Field:     types.StringValue("state_name"),
					Search:    types.StringValue("anillo"),
					Operation: types.StringValue("contains"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{
				{
					City:        "Canillo",
					CountryName: "Andorra",
					CountryCode: "AD",
					StateName:   "Canillo",
					StateCode:   "",
					Timezone:    []string{"Europe/Andorra"},
				},
				{
					City:        "El Tarter",
					CountryName: "Andorra",
					CountryCode: "AD",
					StateName:   "Canillo",
					StateCode:   "",
					Timezone:    []string{"Europe/Andorra"},
				},
			},
		},
		{
			name: "Filter by country_name endsWith",
			filters: []sldFilters{
				{
					Field:     types.StringValue("country_name"),
					Search:    types.StringValue("Futuna"),
					Operation: types.StringValue("endsWith"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{
				{
					City:        "Alo",
					CountryCode: "WF",
					CountryName: "Wallis and Futuna",
					Timezone: []string{
						"Pacific/Wallis",
					},
					StateCode: "",
					StateName: "Alo",
				},
				{
					City:        "Leava",
					CountryCode: "WF",
					CountryName: "Wallis and Futuna",
					Timezone: []string{
						"Pacific/Wallis",
					},
					StateCode: "",
					StateName: "Sigave",
				},
				{
					City:        "Mata-Utu",
					CountryCode: "WF",
					CountryName: "Wallis and Futuna",
					Timezone: []string{
						"Pacific/Wallis",
					},
					StateCode: "",
					StateName: "Uvea",
				},
				{
					City:        "Mua",
					CountryCode: "WF",
					CountryName: "Wallis and Futuna",
					Timezone: []string{
						"Pacific/Wallis",
					},
					StateCode: "",
					StateName: "Uvea",
				},
			},
		},
		{
			name: "Multiple filters (AND logic)",
			filters: []sldFilters{
				{
					Field:     types.StringValue("country_name"),
					Search:    types.StringValue("United States"),
					Operation: types.StringValue("startsWith"),
				},
				{
					Field:     types.StringValue("state_name"),
					Search:    types.StringValue("York"),
					Operation: types.StringValue("contains"),
				},
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("k City"),
					Operation: types.StringValue("endsWith"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{
				{
					City:        "Battery Park City",
					CountryName: "United States",
					CountryCode: "US",
					StateName:   "New York",
					StateCode:   "US-NY",
					Timezone:    []string{"America/New_York"},
				},
				{
					City:        "New York City",
					CountryName: "United States",
					CountryCode: "US",
					StateName:   "New York",
					StateCode:   "US-NY",
					Timezone:    []string{"America/New_York"},
				},
			},
		},
		{
			name: "Filter with no matches",
			filters: []sldFilters{
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("NonExistent"),
					Operation: types.StringValue("exact"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{},
		},
		{
			name: "Filter with partial matches",
			filters: []sldFilters{
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("New York City"),
					Operation: types.StringValue("exact"),
				},
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("NonExistent"),
					Operation: types.StringValue("exact"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{},
		},
		{
			name: "Filter with unknown operation",
			filters: []sldFilters{
				{
					Field:     types.StringValue("city"),
					Search:    types.StringValue("Angeles"),
					Operation: types.StringValue("noOp"),
				},
			},
			expectedSiteLocations: []SLDCatalogEntry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := filterSiteLocations(ctx, tt.filters)
			sort.SliceStable(results, func(i, j int) bool {
				a, b := results[i], results[j]

				if a.CountryName != b.CountryName {
					return a.CountryName < b.CountryName
				}

				if a.City != b.City {
					return a.City < b.City
				}

				return true
			})

			if len(results) != len(tt.expectedSiteLocations) {
				t.Errorf("expected %d results, got %d", len(tt.expectedSiteLocations), len(results))
			}

			for i, expectedSiteLocation := range tt.expectedSiteLocations {
				if results[i].City != expectedSiteLocation.City {
					t.Errorf("expected city %s, got %s", expectedSiteLocation.City, results[i].City)
				}

				if results[i].CountryName != expectedSiteLocation.CountryName {
					t.Errorf("expected country name %s, got %s", expectedSiteLocation.CountryName, results[i].CountryName)
				}

				if results[i].CountryCode != expectedSiteLocation.CountryCode {
					t.Errorf("expected country code %s, got %s", expectedSiteLocation.CountryCode, results[i].CountryCode)
				}

				if results[i].StateName != expectedSiteLocation.StateName {
					t.Errorf("expected state name %s, got %s", expectedSiteLocation.StateName, results[i].StateName)
				}

				if results[i].StateCode != expectedSiteLocation.StateCode {
					t.Errorf("expected state code %s, got %s", expectedSiteLocation.StateCode, results[i].StateCode)
				}

				if len(results[i].Timezone) != len(expectedSiteLocation.Timezone) {
					t.Errorf("expected timezone length %d, got %d", len(expectedSiteLocation.Timezone), len(results[i].Timezone))
				} else {
					for j, expectedTimezone := range expectedSiteLocation.Timezone {
						if results[i].Timezone[j] != expectedTimezone {
							t.Errorf("expected timezone %s, got %s", expectedTimezone, results[i].Timezone[j])
						}
					}
				}
			}
		})
	}
}
