package provider

import (
	"context"
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
	return "Ensures at least one filter is set with field city, state_name, or country_name"
}

func (v filtersValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures at least one filter is set with field `city`, `state_name`, or `country_name`"
}

func (v filtersValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
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

func (d *siteLocationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_siteLocation"
}

func (d *siteLocationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				Description: "Field to filter on (city, stateName, countryName)",
				Required:    false,
				Optional:    true,
				// Validators: []validator.List{
				// 	filtersValidator{},
				// },
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

func (d *siteLocationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoClientData)
}

func (d *siteLocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var allLocations map[string]SiteLocationJsonObj
	if err := json.Unmarshal([]byte(siteLocationJson), &allLocations); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error unmarshalling site location JSON", err.Error()))
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

	filteredLocations := make([]SiteLocationJsonObj, 0)
	for _, location := range allLocations {
		tflog.Debug(ctx, "filteredLocations.location - "+fmt.Sprintf("%v", location))

		matches := true
		for _, filter := range filterList {
			field := filter.Field.ValueString()
			search := filter.Search.ValueString()
			operation := filter.Operation.ValueString()
			tflog.Debug(ctx, "location.field - "+fmt.Sprintf("%v", field))

			var value string
			switch field {
			case "city":
				tflog.Info(ctx, "city present - "+fmt.Sprintf("%v", location.City))
				value = location.City
			case "state_name":
				if location.StateName == "" {
					tflog.Debug(ctx, "state_name is not present on location - "+fmt.Sprintf("%v", location))
					matches = false
				}
				tflog.Debug(ctx, "state_name present - "+fmt.Sprintf("%v", location.StateName))
				value = location.StateName
			case "country_name":
				tflog.Debug(ctx, "country_name present - "+fmt.Sprintf("%v", location.CountryName))
				value = location.CountryName
			default:
				continue
			}

			if matches {
				if operation == "exact" && !(value == search) {
					tflog.Debug(ctx, "field '"+field+"' exact not matched - value '"+value+"' and search '"+search+"' "+fmt.Sprintf("%v", !(value == search)))
					matches = false
				} else if operation == "startsWith" && !(strings.HasPrefix(value, search)) {
					tflog.Debug(ctx, "field '"+field+"' startsWith not matched - value '"+value+"' and search '"+search+"' "+fmt.Sprintf("%v", !(strings.HasPrefix(value, search))))
					matches = false
				} else if operation == "endsWith" && !(strings.HasSuffix(value, search)) {
					tflog.Debug(ctx, "field '"+field+"' endsWith not matched - value '"+value+"' and search '"+search+"' "+fmt.Sprintf("%v", !(strings.HasSuffix(value, search))))
					matches = false
				} else if operation == "contains" && !(strings.Contains(value, search)) {
					tflog.Debug(ctx, "field '"+field+"' contains not matched - value '"+value+"' and search '"+search+"' "+fmt.Sprintf("%v", !(strings.Contains(value, search))))
					matches = false
				}
			}
		}

		tflog.Debug(ctx, "field match - "+fmt.Sprintf("%v", matches)+"' "+fmt.Sprintf("%v", location)+"'")
		if matches {
			filteredLocations = append(filteredLocations, location)
		}
	}
	tflog.Debug(ctx, "field match filteredLocations - "+fmt.Sprintf("%v", filteredLocations))
	setLocations(ctx, resp, state.Filters, filteredLocations)
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
		tflog.Info(ctx, "allLocations.StateCode - "+fmt.Sprintf("%v", loc.StateCode))
		tflog.Info(ctx, "reflect.TypeOf(loc.StateCode) - "+fmt.Sprintf("%v", reflect.TypeOf(loc.StateCode)))

		locObj, diags := types.ObjectValue(
			SiteLocationAttrTypes,
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
		types.ObjectType{AttrTypes: SiteLocationAttrTypes},
		locationsOut,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Info(ctx, "locationsListVal types.ListValueFrom locationsOut - "+fmt.Sprintf("%v", reflect.TypeOf(locationsOut)))
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
