package sitelocation

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sldQuery struct {
	Filters   types.List `tfsdk:"filters"`
	Locations types.List `tfsdk:"locations"`
}

type sldFilters struct {
	Field     types.String `tfsdk:"field"`
	Search    types.String `tfsdk:"search"`
	Operation types.String `tfsdk:"operation"`
}

// SLDCatalogEntry represents a single site location in the datasource catalog
type SLDCatalogEntry struct {
	City        string   `json:"city"`
	CountryCode string   `json:"countryCode"`
	CountryName string   `json:"countryName"`
	StateCode   string   `json:"StateCode"`
	StateName   string   `json:"StateName"`
	Timezone    []string `json:"timezone"`
}
