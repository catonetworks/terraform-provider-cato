package license

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Resource describes the license resource model.
type Resource struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	LicenseID   types.String `tfsdk:"license_id"`
	BW          types.Int64  `tfsdk:"bw"`
	LicenseInfo types.Object `tfsdk:"license_info"`
}

type InfoResource struct {
	AllocatedBandwidth types.Int64  `tfsdk:"allocated_bandwidth"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	LastUpdated        types.String `tfsdk:"last_updated"`
	Plan               types.String `tfsdk:"plan"`
	SiteLicenseGroup   types.String `tfsdk:"site_license_group"`
	SiteLicenseType    types.String `tfsdk:"site_license_type"`
	SKU                types.String `tfsdk:"sku"`
	StartDate          types.String `tfsdk:"start_date"`
	Status             types.String `tfsdk:"status"`
	Total              types.Int64  `tfsdk:"total"`
}

var LicenseResourceType = types.ObjectType{AttrTypes: LicenseResourceAttrTypes}
var LicenseResourceAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"site_id":      types.StringType,
	"license_id":   types.StringType,
	"bw":           types.Int64Type,
	"license_info": LicenseInfoResourceType,
}

var LicenseInfoResourceType = types.ObjectType{AttrTypes: LicenseInfoResourceAttrTypes}
var LicenseInfoResourceAttrTypes = map[string]attr.Type{
	"allocated_bandwidth": types.Int64Type,
	"expiration_date":     types.StringType,
	"last_updated":        types.StringType,
	"plan":                types.StringType,
	"site_license_group":  types.StringType,
	"site_license_type":   types.StringType,
	"sku":                 types.StringType,
	"start_date":          types.StringType,
	"status":              types.StringType,
	"total":               types.Int64Type,
}
