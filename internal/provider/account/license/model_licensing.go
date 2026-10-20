package license

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
)

type LicensingWrapper struct {
	LicensingInfo LicensingInfo `tfsdk:"licensing_info"`
}

type LicensingInfo struct {
	GlobalLicenseAllocations GlobalLicenseAllocations `tfsdk:"global_license_allocations"`
	Licenses                 []License                `tfsdk:"licenses"`
}

type GlobalLicenseAllocations struct {
	PublicIPs CountPublicIps `tfsdk:"public_ips"`
	ZTNAUsers CountZtnaUsers `tfsdk:"ztna_users"`
}

type CountZtnaUsers struct {
	Total     int `tfsdk:"total"`
	Allocated int `tfsdk:"allocated"`
	Available int `tfsdk:"available"`
}

type CountPublicIps struct {
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
	Total          *int    `tfsdk:"total"`
	// DPAVersion            string          `tfsdk:"dpa_version"`
	SiteLicenseGroup *string `tfsdk:"site_license_group"`
	SiteLicenseType  *string `tfsdk:"site_license_type"`
	// Regionality           string          `tfsdk:"regionality"`
	// ZTNAUsersLicenseGroup string          `tfsdk:"ztna_users_license_group"`
	AllocatedBandwidth *int             `tfsdk:"allocated_bandwidth"`
	Site               *Site            `tfsdk:"site"`
	Sites              []*SiteBandwidth `tfsdk:"sites"`
	Accounts           []interface{}    `tfsdk:"accounts"` // or []Account if you define it
}

type Site struct {
	ID   string `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type SiteBandwidth struct {
	Site               Site `tfsdk:"site"`
	AllocatedBandwidth int  `tfsdk:"allocated_bandwidth"`
}

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

// DataSource describes the licensing data source model.
type DataSource struct {
	SKU                      types.String   `tfsdk:"sku"`
	IsActive                 types.Bool     `tfsdk:"is_active"`
	IsAssigned               types.Bool     `tfsdk:"is_assigned"`
	GlobalLicenseAllocations types.Object   `tfsdk:"global_license_allocations"`
	Licenses                 []types.Object `tfsdk:"licenses"`
}

var DataSourceObjectType = types.ObjectType{AttrTypes: DataSourceAttrTypes}
var DataSourceAttrTypes = map[string]attr.Type{
	"sku":                        types.StringType,
	"is_active":                  types.BoolType,
	"is_assigned":                types.BoolType,
	"global_license_allocations": GlobalLicenseAllocationsDataSourceObjectType,
	"licenses": types.ListType{
		ElemType: ObjectType,
	},
}

var GlobalLicenseAllocationsDataSourceObjectType = types.ObjectType{AttrTypes: GlobalLicenseAllocationsDataSourceAttrTypes}
var GlobalLicenseAllocationsDataSourceAttrTypes = map[string]attr.Type{
	"public_ips": types.ObjectType{AttrTypes: CountPublicIpsAttrTypes},
	"ztna_users": types.ObjectType{AttrTypes: CountZtnaUsersAttrTypes},
}

var CountZtnaUsersObjectType = types.ObjectType{AttrTypes: CountZtnaUsersAttrTypes}
var CountZtnaUsersAttrTypes = map[string]attr.Type{
	"total":     types.Int64Type,
	"allocated": types.Int64Type,
	"available": types.Int64Type,
}

var CountPublicIpsObjectType = types.ObjectType{AttrTypes: CountPublicIpsAttrTypes}
var CountPublicIpsAttrTypes = map[string]attr.Type{
	"total":     types.Int64Type,
	"allocated": types.Int64Type,
	"available": types.Int64Type,
}

var FilterObjectType = types.ObjectType{AttrTypes: FilterAttrTypes}
var FilterAttrTypes = map[string]attr.Type{
	"sku":         types.StringType,
	"is_active":   types.BoolType,
	"is_assigned": types.BoolType,
}

var ObjectType = types.ObjectType{AttrTypes: AttrTypes}
var AttrTypes = map[string]attr.Type{
	// Base License Fields
	"id":              types.StringType,
	"sku":             types.StringType,
	"plan":            types.StringType,
	"status":          types.StringType,
	"expiration_date": types.StringType,
	"start_date":      types.StringType,
	"last_updated":    types.StringType,
	// Conditional fields dependent on SKU
	"total": &types.Int64Type,
	// "site_license_group":  &types.StringType,
	// "site_license_type":   &types.StringType,
	// "allocated_bandwidth": &types.Int64Type,
	// "site":                &idname.RefObjectType,
	"sites": types.ListType{
		ElemType: &SiteAllocationObjectType,
	},
	// "accounts": types.ListType{
	// 	ElemType: &types.StringType,
	// },
}

var SiteAllocationObjectType = types.ObjectType{AttrTypes: SiteAllocationAttrTypes}
var SiteAllocationAttrTypes = map[string]attr.Type{
	"site":                idname.RefObjectType,
	"allocated_bandwidth": types.Int64Type,
}
