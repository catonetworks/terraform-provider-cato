package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func LicensingInfoDataSource() datasource.DataSource {
	return &licensingInfoDataSource{}
}

type licensingInfoDataSource struct {
	client *catoClientData
}

func (d *licensingInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_licensingInfo"
}

func (d *licensingInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents licensing information from the Cato API.",
		Attributes: map[string]schema.Attribute{
			"sku": schema.StringAttribute{
				Description: "License SKU to filter for",
				Required:    false,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("CATO_ANTI_MALWARE", "CATO_ANTI_MALWARE_NG", "CATO_CASB", "CATO_DATALAKE", "CATO_DATALAKE_12M", "CATO_DATALAKE_3M", "CATO_DATALAKE_6M", "CATO_DEM", "CATO_DLP", "CATO_EPP", "CATO_ILMM", "CATO_IOT_OT", "CATO_IP_ADD", "CATO_IPS", "CATO_MANAGED_XDR", "CATO_MDR", "CATO_NOCAAS_HF", "CATO_PB", "CATO_PB_SSE", "CATO_RBI", "CATO_SAAS", "CATO_SAAS_SECURITY_API", "CATO_SAAS_SECURITY_API_ALL_APPS", "CATO_SAAS_SECURITY_API_ONE_APP", "CATO_SAAS_SECURITY_API_TWO_APPS", "CATO_SITE", "CATO_SSE_SITE", "CATO_THREAT_PREVENTION", "CATO_THREAT_PREVENTION_ADV", "CATO_XDR_PRO", "CATO_ZTNA_USERS", "MOBILE_USERS"),
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Boolean indicating if license is active",
				Required:    false,
				Optional:    true,
			},
			"is_assigned": schema.BoolAttribute{
				Description: "Boolean indicating if site(s) are assigned to the license",
				Required:    false,
				Optional:    true,
			},
			"global_license_allocations": schema.SingleNestedAttribute{
				Computed: true,
				Required: false,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"public_ips": schema.SingleNestedAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"total": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
							"allocated": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
							"available": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
						},
					},
					"ztna_users": schema.SingleNestedAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"total": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
							"allocated": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
							"available": schema.Int64Attribute{
								Computed: true,
								Required: false,
								Optional: true,
							},
						},
					},
				},
			},
			"licenses": schema.ListNestedAttribute{
				Computed: true,
				Required: false,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"sku": schema.StringAttribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						"plan": schema.StringAttribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						"expiration_date": schema.StringAttribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						"start_date": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						//  License specific attributes that vary by SKU
						"total": schema.Int64Attribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						// "site_license_group": schema.StringAttribute{
						// 	Computed: true,
						// 	Optional: true,
						// 	Required: false,
						// },
						// "site_license_type": schema.StringAttribute{
						// 	Computed: true,
						// 	Optional: true,
						// 	Required: false,
						// },
						// "allocated_bandwidth": schema.Int64Attribute{
						// 	Computed: true,
						// 	Optional: true,
						// 	Required: false,
						// },
						// "site": schema.SingleNestedAttribute{
						// 	Optional: true,
						// 	Computed: true,
						// 	Required: false,
						// 	Attributes: map[string]schema.Attribute{
						// 		"name": schema.StringAttribute{
						// 			Required: false,
						// 			Optional: true,
						// 			Computed: true,
						// 		},
						// 		"id": schema.StringAttribute{
						// 			Required: false,
						// 			Optional: true,
						// 			Computed: true,
						// 		},
						// 	},
						// },
						"sites": schema.ListNestedAttribute{
							Optional: true,
							Computed: true,
							Required: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"site": schema.SingleNestedAttribute{
										Computed: true,
										Required: false,
										Optional: true,
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Required: false,
												Optional: true,
												Computed: true,
											},
											"id": schema.StringAttribute{
												Required: false,
												Optional: true,
												Computed: true,
											},
										},
									},
									"allocated_bandwidth": schema.Int64Attribute{
										Computed: true,
										Required: false,
										Optional: true,
									},
								},
							},
						},
						// "accounts": schema.ListAttribute{
						// 	ElementType: types.MapType{ElemType: types.StringType},
						// 	Computed:    true,
						// 	Required:    false,
						// 	Optional:    true,
						// },
					},
				},
			},
		},
	}
}

func (d *licensingInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *licensingInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LicenseDataSource
	if diags := req.Config.Get(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	licensingInfoResponse, err := d.client.catov2.Licensing(ctx, d.client.AccountId)
	tflog.Debug(ctx, "Read.Licensing.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(licensingInfoResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	if len(licenses) == 0 {
		resp.Diagnostics.AddError(
			"LICENSE API ERROR",
			"No licenses were found on this account.",
		)
		return
	}
	gla := licensingInfoResponse.Licensing.LicensingInfo.GlobalLicenseAllocations

	publicIps, diags := types.ObjectValue(
		LicenseCountPublicIpsAttrTypes,
		map[string]attr.Value{
			"total":     types.Int64Value(int64(gla.PublicIps.Total)),
			"allocated": types.Int64Value(int64(gla.PublicIps.Allocated)),
			"available": types.Int64Value(int64(gla.PublicIps.Available)),
		},
	)
	resp.Diagnostics.Append(diags...)
	ztnaUsers, diags := types.ObjectValue(
		LicenseCountZtnaUsersAttrTypes,
		map[string]attr.Value{
			"total":     types.Int64Value(int64(gla.ZtnaUsers.Total)),
			"allocated": types.Int64Value(int64(gla.ZtnaUsers.Allocated)),
			"available": types.Int64Value(int64(gla.ZtnaUsers.Available)),
		},
	)
	resp.Diagnostics.Append(diags...)
	globalLicenseAllocationsObj, diags := types.ObjectValue(
		GlobalLicenseAllocationsDataSourceAttrTypes,
		map[string]attr.Value{
			"public_ips": publicIps,
			"ztna_users": ztnaUsers,
		},
	)
	resp.Diagnostics.Append(diags...)
	state.GlobalLicenseAllocations = globalLicenseAllocationsObj

	filteredLicenses := make([]types.Object, 0)
	for _, license := range licenses {
		matches := true
		tflog.Debug(ctx, "state.Filters.IsActive value and type", map[string]interface{}{
			"value": state.IsActive,
			"type":  fmt.Sprintf("%T", state.IsActive),
		})
		tflog.Debug(ctx, "license.Status", map[string]interface{}{
			"val":    (license.Status),
			"status": license.Status.IsValid(),
			"value":  string(license.Status),
		})
		tflog.Debug(ctx, "license.Sku", map[string]interface{}{
			"val":           (license.Sku),
			"state.sku":     fmt.Sprintf("%T", state.SKU),
			"state.sku.val": fmt.Sprintf("%v", state.SKU),
			"matches":       fmt.Sprintf("%v", state.SKU.ValueString() == string(license.Sku)),
		})
		if !state.IsActive.IsNull() && !state.IsActive.IsUnknown() {
			matches = state.IsActive.ValueBool() && string(license.Status) == "ACTIVE"
		}
		if !state.SKU.IsNull() && !state.SKU.IsUnknown() {
			matches = (state.SKU.ValueString() == string(license.Sku))
		}
		if matches {
			curTotal := types.Int64Null()
			curSitesListType := types.ListNull(SiteAllocationObjectType)
			resp.Diagnostics.Append(diags...)
			tflog.Debug(ctx, "curTotal", map[string]interface{}{
				"curTotal": (curTotal),
			})

			// // If looking for site assigned licenses, and sku type has no site attribute, exclude
			skuList := []string{"CATO_ANTI_MALWARE", "CATO_DLP", "CATO_SAAS", "CATO_DATALAKE_3M", "CATO_ILMM", "CATO_IP_ADD", "CATO_ZTNA_USERS"}
			skuToCheck := string(license.Sku)
			isInList := false
			for _, sku := range skuList {
				if sku == skuToCheck {
					isInList = true
					break
				}
			}
			if !state.IsAssigned.IsNull() && state.IsAssigned.ValueBool() == true && isInList {
				matches = false
			} else {
				switch license.Sku {
				case "CATO_ANTI_MALWARE", "CATO_ANTI_MALWARE_NG", "CATO_CASB", "CATO_DLP", "CATO_IOT_OT", "CATO_IPS", "CATO_MANAGED_XDR", "CATO_MDR", "CATO_NOCAAS_HF", "CATO_RBI", "CATO_SAAS", "CATO_THREAT_PREVENTION", "CATO_THREAT_PREVENTION_ADV":
					// Handle licenses with no specific attributes
				case "CATO_DATALAKE", "CATO_DATALAKE_12M", "CATO_DATALAKE_3M", "CATO_DATALAKE_6M":
					total := int(license.DataLakeLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_DEM":
					total := int(license.DemLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_EPP":
					total := int(license.EndpointProtectionLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_ILMM":
					total := int(license.IlmmLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_IP_ADD":
					total := int(license.PublicIpsLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_PB", "CATO_PB_SSE":
					total := int(license.PooledBandwidthLicense.Total)
					curTotal = types.Int64Value(int64(total))
					if len(license.PooledBandwidthLicense.Sites) > 0 {
						if !state.IsAssigned.IsNull() {
							matches = state.IsAssigned.ValueBool() == true
						}
						var curSites []attr.Value
						for _, site := range license.PooledBandwidthLicense.Sites {
							curSite, _ := types.ObjectValue(
								NameIDAttrTypes,
								map[string]attr.Value{
									"name": types.StringValue(site.SitePooledBandwidthLicenseSite.Name),
									"id":   types.StringValue(site.SitePooledBandwidthLicenseSite.ID),
								},
							)
							curSiteItem, _ := types.ObjectValue(
								SiteAllocationAttrTypes,
								map[string]attr.Value{
									"site":                curSite,
									"allocated_bandwidth": types.Int64Value(int64(site.AllocatedBandwidth)),
								},
							)
							curSites = append(curSites, curSiteItem)
							// site
						}
						curSitesListType, _ = types.ListValue(SiteAllocationObjectType, curSites)
					} else {
						if !state.IsAssigned.IsNull() {
							matches = state.IsAssigned.ValueBool() == false
						}
					}
				case "CATO_SAAS_SECURITY_API", "CATO_SAAS_SECURITY_API_ALL_APPS", "CATO_SAAS_SECURITY_API_ONE_APP", "CATO_SAAS_SECURITY_API_TWO_APPS":
					total := int(license.SaasSecurityAPILicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_SITE", "CATO_SSE_SITE":
					total := int(license.SiteLicense.Total)
					curTotal = types.Int64Value(int64(total))
					var curSites []attr.Value
					if license.SiteLicense.Site != nil {
						if !state.IsAssigned.IsNull() {
							matches = state.IsAssigned.ValueBool() == true
						}
						curSite, _ := types.ObjectValue(
							NameIDAttrTypes,
							map[string]attr.Value{
								"name": types.StringValue(license.SiteLicense.Site.Name),
								"id":   types.StringValue(license.SiteLicense.Site.ID),
							},
						)
						curSiteItem, _ := types.ObjectValue(
							SiteAllocationAttrTypes,
							map[string]attr.Value{
								"site":                curSite,
								"allocated_bandwidth": types.Int64Null(),
							},
						)
						curSites = append(curSites, curSiteItem)
					}
					curSitesListType, _ = types.ListValue(SiteAllocationObjectType, curSites)
				case "CATO_XDR_PRO":
					total := int(license.XdrProLicense.Total)
					curTotal = types.Int64Value(int64(total))
				case "CATO_ZTNA_USERS", "MOBILE_USERS":
					total := int(license.ZtnaUsersLicense.Total)
					curTotal = types.Int64Value(int64(total))
				default:
					tflog.Warn(ctx, "Unhandled license SKU", map[string]interface{}{
						"sku": license.Sku,
					})
				}
			}
			if matches {
				licenseObj, diagstmp := types.ObjectValue(
					LicenseAttrTypes,
					map[string]attr.Value{
						"id":              types.StringPointerValue(license.ID),
						"sku":             types.StringValue(string(license.Sku)),
						"plan":            types.StringValue(string(license.Plan)),
						"status":          types.StringValue(string(license.Status)),
						"expiration_date": types.StringValue(parseTime(license.ExpirationDate)),
						"start_date":      types.StringValue(parseTimeP(license.StartDate)),
						"last_updated":    types.StringValue(parseTimeP(license.LastUpdated)),
						// Conditional fields dependent on SKU
						"total": curTotal,
						"sites": curSitesListType,
					},
				)
				diags = append(diags, diagstmp...)
				filteredLicenses = append(filteredLicenses, licenseObj)
			}
		}
	}
	state.Licenses = filteredLicenses
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

type attrType interface{}
type typesObjectType struct {
	AttrTypes map[string]attrType
}
type typesListType struct {
	ElemType typesObjectType
}

func parseTimeP(t *string) string {
	if t == nil {
		return types.StringNull().ValueString()
	}
	parsedTime, _ := time.Parse(time.RFC3339, *t)
	return parsedTime.Format(time.RFC3339)
}

func parseTime(t string) string {
	parsedTime, _ := time.Parse(time.RFC3339, t)
	return parsedTime.Format(time.RFC3339)
}
