package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LicensingWrapper struct {
	LicensingInfo LicensingInfo `tfsdk:"licensing_info"`
}

type LicensingInfo struct {
	GlobalLicenseAllocations GlobalLicenseAllocations `tfsdk:"global_license_allocations"`
	Licenses                 []License                `tfsdk:"licenses"`
}

type GlobalLicenseAllocations struct {
	PublicIPs LicenseCountPublicIps `tfsdk:"public_ips"`
	ZTNAUsers LicenseCountZtnaUsers `tfsdk:"ztna_users"`
}

type LicenseCountZtnaUsers struct {
	Total     int `tfsdk:"total"`
	Allocated int `tfsdk:"allocated"`
	Available int `tfsdk:"available"`
}

type LicenseCountPublicIps struct {
	Allocated int `tfsdk:"allocated"`
	Available int `tfsdk:"available"`
}

type License struct {
	ID             *string `tfsdk:"id"`
	SKU            string  `tfsdk:"sku"`
	Plan           string  `tfsdk:"plan"`
	Status         string  `tfsdk:"status"`
	ExpirationDate string  `tfsdk:"expiration_date"`
	StartDate      string  `tfsdk:"start_date"`
	LastUpdated    string  `tfsdk:"last_updated"`
	Total          int     `tfsdk:"total"`
	// DPAVersion            string  `tfsdk:"dpa_version"`
	SiteLicenseGroup string `tfsdk:"site_license_group"`
	SiteLicenseType  string `tfsdk:"site_license_type"`
	// Regionality           string  `tfsdk:"regionality"`
	// ZTNAUsersLicenseGroup string  `tfsdk:"ztna_users_license_group"`
	AllocatedBandwidth *int  `tfsdk:"allocated_bandwidth"`
	Site               *Site `tfsdk:"site"`
	// Sites                 []SiteBandwidth `tfsdk:"sites"`
	// Accounts              []interface{}   `tfsdk:"accounts"` // or []Account if you define it
}

type Site struct {
	ID   string `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type SiteBandwidth struct {
	Site               Site `tfsdk:"site"`
	AllocatedBandwidth int  `tfsdk:"allocated_bandwidth"`
}

var SiteType = types.ObjectType{AttrTypes: SiteAttrType}
var SiteAttrType = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

var SiteBandwidthType = types.ObjectType{AttrTypes: SiteBandwidthAttrType}
var SiteBandwidthAttrType = map[string]attr.Type{
	"site":                SiteType,
	"allocated_bandwidth": types.Int64Type,
}

var LicenseCountPublicIpsType = types.ObjectType{AttrTypes: LicenseCountPublicIpsAttrType}
var LicenseCountPublicIpsAttrType = map[string]attr.Type{
	// "total":     types.Int64Type,
	"allocated": types.Int64Type,
	"available": types.Int64Type,
}

var LicenseCountZtnaUsersType = types.ObjectType{AttrTypes: LicenseCountZtnaUsersAttrType}
var LicenseCountZtnaUsersAttrType = map[string]attr.Type{
	"total":     types.Int64Type,
	"allocated": types.Int64Type,
	"available": types.Int64Type,
}

var GlobalLicenseAllocationsType = types.ObjectType{AttrTypes: GlobalLicenseAllocationsAttrType}
var GlobalLicenseAllocationsAttrType = map[string]attr.Type{
	"public_ips": LicenseCountPublicIpsType,
	"ztna_users": LicenseCountZtnaUsersType,
}

var LicenseType = types.ObjectType{AttrTypes: LicenseAttrType}
var LicenseAttrType = map[string]attr.Type{
	"id":                       types.StringType,
	"sku":                      types.StringType,
	"plan":                     types.StringType,
	"status":                   types.StringType,
	"expiration_date":          types.StringType,
	"start_date":               types.StringType,
	"last_updated":             types.StringType,
	"total":                    types.Int64Type,
	"dpa_version":              types.StringType,
	"site_license_group":       types.StringType,
	"site_license_type":        types.StringType,
	"regionality":              types.StringType,
	"ztna_users_license_group": types.StringType,
	"allocated_bandwidth":      types.Int64Type,
	"site":                     SiteType,
	"sites": types.ListType{
		ElemType: SiteBandwidthType,
	},
	"accounts": types.ListType{
		ElemType: types.MapType{
			ElemType: types.StringType,
		},
	},
}

var LicensingInfoType = types.ObjectType{AttrTypes: LicensingInfoAttrType}
var LicensingInfoAttrType = map[string]attr.Type{
	"global_license_allocations": GlobalLicenseAllocationsType,
	"licenses": types.ListType{
		ElemType: LicenseType,
	},
}

// License Resource Structs and Types

type LicenseResource struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	LicenseIDCurrent types.String `tfsdk:"license_id_current"`
	LicenseID        types.String `tfsdk:"license_id"`
	BW               types.Int64  `tfsdk:"bw"`
	LicenseInfo      types.Object `tfsdk:"license_info"`
}

type LicenseInfoResource struct {
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
	"id":                 types.StringType,
	"site_id":            types.StringType,
	"license_id_current": types.StringType,
	"license_id":         types.StringType,
	"bw":                 types.Int64Type,
	"license_info":       LicenseInfoResourceType,
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
