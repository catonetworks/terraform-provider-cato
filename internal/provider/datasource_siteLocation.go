package provider

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	_ "modernc.org/sqlite"
)

//go:embed type_site_location_data.db
var siteLocationDB []byte

var (
	dbOnce     sync.Once
	dbPath     string
	dbInitErr  error
	locationDB *sql.DB
)

func initSiteLocationDB() error {
	dbOnce.Do(func() {
		// Create temp file for the database
		tmpDir := os.TempDir()
		dbPath = filepath.Join(tmpDir, "cato_site_location.db")

		// Write embedded database to temp file
		if err := os.WriteFile(dbPath, siteLocationDB, 0644); err != nil {
			dbInitErr = fmt.Errorf("failed to write site location database: %w", err)
			return
		}

		// Open database connection
		db, err := sql.Open("sqlite", dbPath+"?mode=ro")
		if err != nil {
			dbInitErr = fmt.Errorf("failed to open site location database: %w", err)
			return
		}

		// Test connection
		if err := db.Ping(); err != nil {
			dbInitErr = fmt.Errorf("failed to ping site location database: %w", err)
			return
		}

		locationDB = db
	})
	return dbInitErr
}

func SiteLocationDataSource() datasource.DataSource {
	return &siteLocationDataSource{}
}

type siteLocationDataSource struct {
	client *catoClientData
}

type siteLocationQuery struct {
	Filters   types.List `tfsdk:"filters"`
	Locations types.List `tfsdk:"locations"`
}

type filters struct {
	Field     types.String `tfsdk:"field"`
	Search    types.String `tfsdk:"search"`
	Operation types.String `tfsdk:"operation"`
}

type locations struct {
	CountryCode types.String `tfsdk:"country_code"`
	CountryName types.String `tfsdk:"country_name"`
	StateCode   types.String `tfsdk:"state_code"`
	StateName   types.String `tfsdk:"state_name"`
	Timezone    types.List   `tfsdk:"timezone"`
	City        types.String `tfsdk:"city"`
}

type SiteLocationJsonObj struct {
	City        string   `json:"city"`
	CountryCode string   `json:"countryCode"`
	CountryName string   `json:"countryName"`
	StateCode   string   `json:"StateCode"`
	StateName   string   `json:"StateName"`
	Timezone    []string `json:"timezone"`
}

var SiteLocationQueryObjectType = types.ObjectType{AttrTypes: SiteLocationQueryAttrTypes}
var SiteLocationQueryAttrTypes = map[string]attr.Type{
	"filters": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: SiteLocationFilterAttrTypes,
		},
	},
	"locations": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: SiteLocationAttrTypes,
		},
	},
}

var SiteLocationFilterObjectType = types.ObjectType{AttrTypes: SiteLocationFilterAttrTypes}
var SiteLocationFilterAttrTypes = map[string]attr.Type{
	"field":     types.StringType,
	"search":    types.StringType,
	"operation": types.StringType,
}

var SiteLocationObjectType = types.ObjectType{AttrTypes: SiteLocationAttrTypes}
var SiteLocationAttrTypes = map[string]attr.Type{
	"country_code": types.StringType,
	"country_name": types.StringType,
	"state_code":   types.StringType,
	"state_name":   types.StringType,
	"timezone":     types.ListType{ElemType: types.StringType},
	"city":         types.StringType,
}

// Custom validator for filters
type filtersValidator struct{}

func (v filtersValidator) Description(ctx context.Context) string {
	return "Validates filter configuration: use state_code OR state_name (not both), and country_code OR country_name (not both)"
}

func (v filtersValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates filter configuration: use `state_code` OR `state_name` (not both), and `country_code` OR `country_name` (not both)"
}

func (v filtersValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()
	if len(elements) == 0 {
		return
	}

	// Track which fields are used
	hasStateCode := false
	hasStateName := false
	hasCountryCode := false
	hasCountryName := false

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		if fieldAttr, exists := attrs["field"]; exists {
			if fieldStr, ok := fieldAttr.(types.String); ok && !fieldStr.IsNull() && !fieldStr.IsUnknown() {
				field := fieldStr.ValueString()
				switch field {
				case "state_code":
					hasStateCode = true
				case "state_name":
					hasStateName = true
				case "country_code":
					hasCountryCode = true
				case "country_name":
					hasCountryName = true
				}
			}
		}
	}

	// Validate mutually exclusive fields
	if hasStateCode && hasStateName {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"Cannot use both 'state_code' and 'state_name' filters. Use one or the other.",
		))
	}

	if hasCountryCode && hasCountryName {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"Cannot use both 'country_code' and 'country_name' filters. Use one or the other.",
		))
	}
}

func (d *siteLocationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_siteLocation"
}

func (d *siteLocationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves account site locations.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				Description: "Field to filter on (city, state_code/state_name, country_code/country_name). Note: use state_code OR state_name, not both; use country_code OR country_name, not both.",
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					filtersValidator{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "Field to filter on (city, state_code, state_name, country_code, country_name). Use state_code OR state_name (not both), and country_code OR country_name (not both).",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("state_code", "state_name", "country_code", "country_name", "city"),
							},
						},
						"search": schema.StringAttribute{
							Description: "String value to search for",
							Required:    true,
						},
						"operation": schema.StringAttribute{
							Description: "Filter operation (exact, startsWith, endsWith, contains)",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("exact", "startsWith", "endsWith", "contains"),
							},
						},
					},
				},
			},
			"locations": schema.ListNestedAttribute{
				Description: "Filtered site locations",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"country_code": schema.StringAttribute{
							Description: "Country code",
							Computed:    true,
						},
						"country_name": schema.StringAttribute{
							Description: "Country name",
							Computed:    true,
						},
						"state_code": schema.StringAttribute{
							Description: "State code",
							Computed:    true,
						},
						"state_name": schema.StringAttribute{
							Description: "State name",
							Computed:    true,
						},
						"timezone": schema.ListAttribute{
							Description: "Timezone",
							ElementType: types.StringType,
							Computed:    true,
						},
						"city": schema.StringAttribute{
							Description: "City name",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *siteLocationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

func (d *siteLocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Initialize database
	if err := initSiteLocationDB(); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error initializing site location database", err.Error()))
		return
	}

	var state siteLocationQuery
	if diags := req.Config.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if state.Filters.IsNull() || state.Filters.IsUnknown() {
		setLocations(ctx, resp, state.Filters, []SiteLocationJsonObj{})
		return
	}

	var filterList []filters
	if diags := state.Filters.ElementsAs(ctx, &filterList, false); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Build SQL query based on filters
	filteredLocations, err := queryLocationsFromDB(ctx, filterList)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error querying site locations", err.Error()))
		return
	}

	tflog.Debug(ctx, "Filtered locations from SQLite", map[string]interface{}{
		"count": len(filteredLocations),
	})

	setLocations(ctx, resp, state.Filters, filteredLocations)
}

// queryLocationsFromDB queries the SQLite database with the given filters
func queryLocationsFromDB(ctx context.Context, filterList []filters) ([]SiteLocationJsonObj, error) {
	if locationDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Build WHERE clause from filters
	var conditions []string
	var args []interface{}

	for _, filter := range filterList {
		field := filter.Field.ValueString()
		search := filter.Search.ValueString()
		operation := filter.Operation.ValueString()

		// Map Terraform field names to database column names
		var dbColumn string
		switch field {
		case "city":
			dbColumn = "city"
		case "state_code":
			dbColumn = "stateCode"
		case "state_name":
			dbColumn = "stateName"
		case "country_code":
			dbColumn = "countryCode"
		case "country_name":
			dbColumn = "countryName"
		default:
			continue
		}

		// Build condition based on operation
		switch operation {
		case "exact":
			conditions = append(conditions, fmt.Sprintf("%s = ?", dbColumn))
			args = append(args, search)
		case "startsWith":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", dbColumn))
			args = append(args, search+"%")
		case "endsWith":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", dbColumn))
			args = append(args, "%"+search)
		case "contains":
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", dbColumn))
			args = append(args, "%"+search+"%")
		}
	}

	// Build query
	query := "SELECT city, countryCode, countryName, stateCode, stateName, timezone FROM locations"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " LIMIT 1000" // Limit results to prevent excessive data

	tflog.Debug(ctx, "Executing SQLite query", map[string]interface{}{
		"query": query,
		"args":  args,
	})

	rows, err := locationDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var results []SiteLocationJsonObj
	for rows.Next() {
		var loc SiteLocationJsonObj
		var stateCode, stateName, timezone sql.NullString

		if err := rows.Scan(&loc.City, &loc.CountryCode, &loc.CountryName, &stateCode, &stateName, &timezone); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		if stateCode.Valid {
			loc.StateCode = stateCode.String
		}
		if stateName.Valid {
			loc.StateName = stateName.String
		}
		if timezone.Valid && timezone.String != "" {
			// Database stores single timezone, convert to slice for compatibility
			loc.Timezone = []string{timezone.String}
		}

		results = append(results, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}

func setLocations(ctx context.Context, resp *datasource.ReadResponse, filters types.List, allLocations []SiteLocationJsonObj) {
	diags := make(diag.Diagnostics, 0)
	locationsOut := make([]attr.Value, 0, len(allLocations))

	filtersListVal, diags := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: SiteLocationFilterAttrTypes},
		filters,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, loc := range allLocations {
		tflog.Debug(ctx, "Processing location", map[string]interface{}{
			"city":        loc.City,
			"countryCode": loc.CountryCode,
			"stateCode":   loc.StateCode,
		})

		locObj, diags := types.ObjectValue(
			SiteLocationAttrTypes,
			map[string]attr.Value{
				"country_code": func() attr.Value {
					if loc.CountryCode != "" {
						return types.StringValue(loc.CountryCode)
					}
					return types.StringNull()
				}(),
				"country_name": func() attr.Value {
					if loc.CountryName != "" {
						return types.StringValue(loc.CountryName)
					}
					return types.StringNull()
				}(),
				"state_code": func() attr.Value {
					if loc.StateCode != "" {
						return types.StringValue(loc.StateCode)
					}
					return types.StringNull()
				}(),
				"state_name": func() attr.Value {
					if loc.StateName != "" {
						return types.StringValue(loc.StateName)
					}
					return types.StringNull()
				}(),
				"timezone": func() attr.Value {
					if len(loc.Timezone) > 0 {
						timezoneValues := make([]attr.Value, len(loc.Timezone))
						for i, tz := range loc.Timezone {
							timezoneValues[i] = types.StringValue(tz)
						}
						listValue, diags := types.ListValue(types.StringType, timezoneValues)
						if diags.HasError() {
							resp.Diagnostics.Append(diags...)
							return types.ListNull(types.StringType)
						}
						return listValue
					}
					return types.ListNull(types.StringType)
				}(),
				"city": func() attr.Value {
					if loc.City != "" {
						return types.StringValue(loc.City)
					}
					return types.StringNull()
				}(),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		locationsOut = append(locationsOut, locObj)
	}

	locationsListVal, diags := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: SiteLocationAttrTypes},
		locationsOut,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	state, diags := types.ObjectValue(
		SiteLocationQueryAttrTypes,
		map[string]attr.Value{
			"filters":   filtersListVal,
			"locations": locationsListVal,
		},
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
