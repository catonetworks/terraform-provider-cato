package provider

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed type_site_location_data.json
var siteLocationJson string
var data map[string]interface{}

func init() {
	json.Unmarshal([]byte(siteLocationJson), &data)
}

// SiteLocationData represents the resolved location information
type SiteLocationData struct {
	Timezone  string
	StateCode string
}

// populateSiteLocationData resolves timezone, state_code and country_code from the location data
// based on country, state and city. This function mimics the logic from lines 272-333 in the CLI export_sites.py
func populateSiteLocationData(countryName, stateName, cityName string) SiteLocationData {
	result := SiteLocationData{}

	if countryName == "" {
		return result
	}

	// If city is provided, try exact match first
	if cityName != "" {
		// Create lookup key based on available data
		var lookupKey string
		if stateName != "" {
			lookupKey = fmt.Sprintf("%s___%s___%s", countryName, stateName, cityName)
		} else {
			lookupKey = fmt.Sprintf("%s___%s", countryName, cityName)
		}

		// Look up location details in the embedded data
		if locationData, exists := data[lookupKey]; exists {
			if locationMap, ok := locationData.(map[string]interface{}); ok {
				// Get timezone - always use the first timezone in the array (same as CLI logic line 332)
				if timezones, ok := locationMap["timezone"].([]interface{}); ok && len(timezones) > 0 {
					if timezone, ok := timezones[0].(string); ok {
						result.Timezone = timezone
					}
				}
				// Get state code (corresponds to CLI line 328: cur_site['stateCode'] = location_data.get('stateCode', None))
				if stateCode, ok := locationMap["stateCode"].(string); ok {
					result.StateCode = stateCode
				}
				return result
			}
		}

		// If exact match not found, try to find similar keys for fallback (optional)
		// This is similar to the debugging logic in CLI but simplified for production use
		for key, locationData := range data {
			if strings.Contains(key, countryName) && strings.Contains(key, cityName) {
				if locationMap, ok := locationData.(map[string]interface{}); ok {
					// Get timezone
					if timezones, ok := locationMap["timezone"].([]interface{}); ok && len(timezones) > 0 {
						if timezone, ok := timezones[0].(string); ok {
							result.Timezone = timezone
						}
					}
					// Get state code
					if stateCode, ok := locationMap["stateCode"].(string); ok {
						result.StateCode = stateCode
					}
					return result
				}
			}
		}
	}

	// Fallback: if city is not provided or not found, try to find ANY entry for the country
	// Use the first matching entry's timezone (works well for countries with single timezone)
	for key, locationData := range data {
		if strings.HasPrefix(key, countryName+"___") {
			if locationMap, ok := locationData.(map[string]interface{}); ok {
				// Get timezone from first matching country entry
				if timezones, ok := locationMap["timezone"].([]interface{}); ok && len(timezones) > 0 {
					if timezone, ok := timezones[0].(string); ok {
						result.Timezone = timezone
					}
				}
				// Get state code if present
				if stateCode, ok := locationMap["stateCode"].(string); ok {
					result.StateCode = stateCode
				}
				return result
			}
		}
	}

	// Return empty struct if no match found
	return result
}
