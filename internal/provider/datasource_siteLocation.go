package provider

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

//go:embed type_site_location_data.json
var sldFilename string
var sldCatalog map[string]SLDCatalogEntry
var sldCatalogLoadError error

//nolint:gochecknoinits
func init() {
	sldCatalogLoadError = json.Unmarshal([]byte(sldFilename), &sldCatalog)
}

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

var (
	SLDFilterAttrTypes = map[string]attr.Type{
		"field":     types.StringType,
		"search":    types.StringType,
		"operation": types.StringType,
	}
	SLDFilterObjectType = types.ObjectType{AttrTypes: SLDFilterAttrTypes}

	SLDAttrTypes = map[string]attr.Type{
		"country_code": types.StringType,
		"country_name": types.StringType,
		"state_code":   types.StringType,
		"state_name":   types.StringType,
		"timezone":     types.ListType{ElemType: types.StringType},
		"city":         types.StringType,
	}
	SLDObjectType = types.ObjectType{AttrTypes: SLDAttrTypes}

	SLDQueryAttrTypes = map[string]attr.Type{
		"filters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: SLDFilterAttrTypes,
			},
		},
		"locations": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: SLDAttrTypes,
			},
		},
	}
	SLDQueryObjectType = types.ObjectType{AttrTypes: SLDQueryAttrTypes}
)

type sldFilterValidator struct{}

func (v sldFilterValidator) Description(_ context.Context) string {
	return "Ensures at least one filter is set with " +
		"field city, state_name, or country_name"
}

func (v sldFilterValidator) MarkdownDescription(_ context.Context) string {
	return "Ensures at least one filter is set with " +
		"field `city`, `state_name`, or `country_name`"
}

func (v sldFilterValidator) ValidateList(
	_ context.Context,
	req validator.ListRequest,
	resp *validator.ListResponse,
) {
	// Allow unknown values - they will be validated once resolved during planning
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"At least one filter must be specified with field city, state_name, or country_name",
		))
		return
	}

	elements := req.ConfigValue.Elements()
	if len(elements) == 0 {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"At least one filter must be specified with field city, state_name, or country_name",
		))
		return
	}
}

type siteLocationDataSource struct {
	client *catoClientData
}

func SiteLocationDataSource() datasource.DataSource {
	return &siteLocationDataSource{}
}

func (d *siteLocationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_siteLocation"
}

func (d *siteLocationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				Description: "Field to filter on (city, stateName, countryName)",
				Optional:    true,
				Validators: []validator.List{
					sldFilterValidator{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "Field to filter on (city, stateName, countryName)",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("state_name", "country_name", "city"),
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

func (d *siteLocationDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

func (d *siteLocationDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	state := &sldQuery{}
	if diags := req.Config.Get(ctx, state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if state.Filters.IsNull() || state.Filters.IsUnknown() {
		setSiteLocations(ctx, resp, state.Filters, []SLDCatalogEntry{})
		return
	}

	if sldCatalogLoadError != nil {
		resp.Diagnostics.Append(
			diag.NewErrorDiagnostic(
				"Site Location Catalog Load Error",
				"Unable to read the site locations catalog file"+
					"Error: "+sldCatalogLoadError.Error(),
			),
		)
		return
	}

	filters := make([]sldFilters, 0)
	if diags := state.Filters.ElementsAs(ctx, &filters, false); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if locations := filterSiteLocations(ctx, filters); len(locations) > 0 {
		setSiteLocations(ctx, resp, state.Filters, locations[:1])
		return
	}

	setSiteLocations(ctx, resp, state.Filters, []SLDCatalogEntry{})
}

func filterSiteLocations(
	ctx context.Context,
	filters []sldFilters,
) []SLDCatalogEntry {
	getLocationProperty := func(field string, location SLDCatalogEntry) string {
		switch field {
		case "city":
			return location.City
		case "state_name":
			return location.StateName
		case "country_name":
			return location.CountryName
		default:
			return ""
		}
	}
	getSearchOperation := func(op string, value string) func(string) bool {
		return func(search string) bool {
			switch op {
			case "exact":
				return search == value
			case "startsWith":
				return strings.HasPrefix(value, search)
			case "endsWith":
				return strings.HasSuffix(value, search)
			case "contains":
				return strings.Contains(value, search)
			default:
				return false
			}
		}
	}
	filteredSiteLocations := sldCatalog

	tflog.Debug(ctx, "filtering locations")
	for _, filter := range filters {
		newFilteredSiteLocations := make(map[string]SLDCatalogEntry, 0)
		for siteLocationKey, siteLocation := range filteredSiteLocations {
			siteLocationProp := getLocationProperty(filter.Field.ValueString(), siteLocation)
			if siteLocationProp == "" {
				continue
			}
			searchOp := getSearchOperation(filter.Operation.ValueString(), siteLocationProp)

			if searchOp(filter.Search.ValueString()) {
				tflog.Debug(ctx, "location found", map[string]interface{}{
					"field":    filter.Field.ValueString(),
					"location": fmt.Sprintf("%v", siteLocation),
				})
				newFilteredSiteLocations[siteLocationKey] = siteLocation
			}
		}

		filteredSiteLocations = newFilteredSiteLocations
	}

	siteLocations := make([]SLDCatalogEntry, 0, len(filteredSiteLocations))
	for _, location := range filteredSiteLocations {
		siteLocations = append(siteLocations, location)
	}

	return siteLocations
}

//nolint:funlen // legacy code
func setSiteLocations(
	ctx context.Context,
	resp *datasource.ReadResponse,
	filters types.List, allLocations []SLDCatalogEntry,
) {
	diags := make(diag.Diagnostics, 0)
	locationsOut := make([]attr.Value, 0, len(allLocations))

	filtersListVal, diags := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: SLDFilterAttrTypes},
		filters,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, loc := range allLocations {
		tflog.Info(ctx, "allLocations.StateCode - "+fmt.Sprintf("%v", loc.StateCode))
		tflog.Info(ctx, "reflect.TypeOf(loc.StateCode) - "+fmt.Sprintf("%v", reflect.TypeOf(loc.StateCode)))

		locObj, diags := types.ObjectValue(
			SLDAttrTypes,
			map[string]attr.Value{
				"country_code": func() attr.Value {
					if loc.CountryCode != "" {
						return types.StringValue(string(loc.CountryCode))
					}
					return types.StringNull()
				}(),
				"country_name": func() attr.Value {
					if loc.CountryName != "" {
						return types.StringValue(string(loc.CountryName))
					}
					return types.StringNull()
				}(),
				"state_code": func() attr.Value {
					if loc.StateCode != "" {
						return types.StringValue(string(loc.StateCode))
					}
					return types.StringNull()
				}(),
				"state_name": func() attr.Value {
					if loc.StateName != "" {
						return types.StringValue(string(loc.StateName))
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
					return types.StringNull()
				}(),
				"city": func() attr.Value {
					if loc.City != "" {
						return types.StringValue(string(loc.City))
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

	tflog.Info(ctx, "locationsListVal types.ListValueFrom locationsOut - "+fmt.Sprintf("%v", reflect.TypeOf(locationsOut)))
	locationsListVal, diags := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: SLDAttrTypes},
		locationsOut,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Info(ctx, "locationsListVal types.ListValueFrom locationsOut - "+fmt.Sprintf("%v", reflect.TypeOf(locationsOut)))
	state, diags := types.ObjectValue(
		SLDQueryAttrTypes,
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
