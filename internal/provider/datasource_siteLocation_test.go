package provider

import (
	"context"
	"reflect"
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
					CountryName: "United States",
					CountryCode: "US",
					StateName:   "California",
					StateCode:   "US-CA",
					Timezone:    []string{"America/Los_Angeles"},
				},
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

func TestInsertSLDCatalogEntry(t *testing.T) {
	t.Run("inserting a new entry into an empty list", func(t *testing.T) {
		newEntry := SLDCatalogEntry{
			City:        "Austin",
			StateName:   "Texas",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{newEntry}

		got := insertSLDCatalogEntry(nil, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a city name lexicographically lower than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "Chicago",
				StateName:   "Illinois",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "Boston",
			StateName:   "Massachusetts",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{newEntry, entries[0]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a city name lexicographically greater than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "Boston",
				StateName:   "Texas",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "Chicago",
			StateName:   "Alabama",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{entries[0], newEntry}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a city name lexicographically lower than a few entries but greater than others", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "Atlanta",
				StateName:   "Georgia",
				CountryName: "United States",
			},
			{
				City:        "Chicago",
				StateName:   "Illinois",
				CountryName: "United States",
			},
			{
				City:        "Denver",
				StateName:   "Colorado",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "Boston",
			StateName:   "Massachusetts",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{entries[0], newEntry, entries[1], entries[2]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a state name lexicographically lower than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "Illinois",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "Georgia",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{newEntry, entries[0]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a state name lexicographically greater than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "Georgia",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "Illinois",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{entries[0], newEntry}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a state name lexicographically lower than some entries but greater than others", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "Georgia",
				CountryName: "United States",
			},
			{
				City:        "CityA",
				StateName:   "Illinois",
				CountryName: "United States",
			},
			{
				City:        "CityA",
				StateName:   "Texas",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "Massachusetts",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{entries[0], entries[1], newEntry, entries[2]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a country name lexicographically lower than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "StateA",
				CountryName: "United States",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "StateA",
			CountryName: "Argentina",
		}
		expected := []SLDCatalogEntry{newEntry, entries[0]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a country name lexicographically greater than existing entries", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "StateA",
				CountryName: "Argentina",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "StateA",
			CountryName: "United States",
		}
		expected := []SLDCatalogEntry{entries[0], newEntry}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert a new entry with a country name lexicographically lower than some entries but greater than others", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "StateA",
				CountryName: "Argentina",
			},
			{
				City:        "CityA",
				StateName:   "StateA",
				CountryName: "United States",
			},
			{
				City:        "CityA",
				StateName:   "StateA",
				CountryName: "Zimbabwe",
			},
		}
		newEntry := SLDCatalogEntry{
			City:        "CityA",
			StateName:   "StateA",
			CountryName: "Brazil",
		}
		expected := []SLDCatalogEntry{entries[0], newEntry, entries[1], entries[2]}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
	})

	t.Run("insert new entries in different positions", func(t *testing.T) {
		entries := []SLDCatalogEntry{
			{
				City:        "CityA",
				StateName:   "StateB",
				CountryName: "CountryC",
			},
			{
				City:        "CityA",
				StateName:   "StateC",
				CountryName: "CountryC",
			},
			{
				City:        "CityB",
				StateName:   "StateB",
				CountryName: "CountryC",
			},
			{
				City:        "CityC",
				StateName:   "StateB",
				CountryName: "CountryD",
			},
			{
				City:        "CityC",
				StateName:   "StateC",
				CountryName: "CountryE",
			},
		}

		newEntry := SLDCatalogEntry{
			City:        "CityB",
			StateName:   "StateA",
			CountryName: "CountryC",
		}
		expected := []SLDCatalogEntry{
			entries[0],
			entries[1],
			newEntry,
			entries[2],
			entries[3],
			entries[4],
		}

		got := insertSLDCatalogEntry(entries, newEntry)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("unexpected entries: got %#v want %#v", got, expected)
		}
		newEntry2 := SLDCatalogEntry{
			City:        "CityC",
			StateName:   "StateB",
			CountryName: "CountryB",
		}
		expected2 := []SLDCatalogEntry{
			entries[0],
			entries[1],
			newEntry,
			entries[2],
			newEntry2,
			entries[3],
			entries[4],
		}

		got2 := insertSLDCatalogEntry(got, newEntry2)

		if !reflect.DeepEqual(got2, expected2) {
			t.Fatalf("unexpected entries: got %#v want %#v", got2, expected2)
		}
	})
}
