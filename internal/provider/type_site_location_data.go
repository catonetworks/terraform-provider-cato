package provider

import (
	_ "embed"
	"encoding/json"
)

//go:embed type_site_location_data.json
var siteLocationJson string
var data map[string]interface{}

func init() {
	json.Unmarshal([]byte(siteLocationJson), &data)
}
