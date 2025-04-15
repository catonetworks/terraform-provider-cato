package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
							// "total": schema.Int64Attribute{
							// 	Computed: true,
							// 	Required: false,
							// 	Optional: true,
							// },
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
						"total": schema.Int64Attribute{
							Computed: true,
							Required: false,
							Optional: true,
						},
						"dpa_version": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"site_license_group": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"site_license_type": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"regionality": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"ztna_users_license_group": schema.StringAttribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"allocated_bandwidth": schema.Int64Attribute{
							Computed: true,
							Optional: true,
							Required: false,
						},
						"site": schema.SingleNestedAttribute{
							Optional: true,
							Computed: true,
							Required: false,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed: true,
									Required: false,
									Optional: true,
								},
								"name": schema.StringAttribute{
									Computed: true,
									Required: false,
									Optional: true,
								},
							},
						},
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
											"id": schema.StringAttribute{
												Computed: true,
												Required: false,
												Optional: true,
											},
											"name": schema.StringAttribute{
												Computed: true,
												Required: false,
												Optional: true,
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
						"accounts": schema.ListAttribute{
							ElementType: types.MapType{ElemType: types.StringType},
							Computed:    true,
							Required:    false,
							Optional:    true,
						},
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
	var state LicensingInfo
	tflog.Info(ctx, "state - "+fmt.Sprintf("%v", state))
	licensingInfoResponse, err := d.client.catov2.Licensing(ctx, d.client.AccountId)
	tflog.Info(ctx, "licensingInfoResponse - "+fmt.Sprintf("%v", licensingInfoResponse))

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	gla := licensingInfoResponse.Licensing.LicensingInfo.GlobalLicenseAllocations

	state.GlobalLicenseAllocations = GlobalLicenseAllocations{
		PublicIPs: LicenseCountPublicIps{
			Allocated: int(gla.PublicIps.Allocated),
			Available: int(gla.PublicIps.Available),
		},
		ZTNAUsers: LicenseCountZtnaUsers{
			Total:     int(gla.ZtnaUsers.Total),
			Allocated: int(gla.ZtnaUsers.Allocated),
			Available: int(gla.ZtnaUsers.Available),
		},
	}

	li := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	var licenses []License
	for _, license := range li {
		licenses = append(licenses, License{
			ID:             license.ID,
			SKU:            string(license.Sku),
			Plan:           license.Plan.String(),
			Status:         string(license.Status),
			ExpirationDate: parseTime(license.ExpirationDate),
			StartDate:      parseTimeP(license.StartDate),
			LastUpdated:    parseTimeP(license.LastUpdated),
			Total:          int(0),
			// DPAVersion:            types.StringNull().ValueString(),
			// SiteLicenseGroup:      types.StringNull().ValueString(),
			// SiteLicenseType:       types.StringNull().ValueString(),
			// Regionality:           types.StringNull().ValueString(),
			// ZTNAUsersLicenseGroup: types.StringNull().ValueString(),
			// AllocatedBandwidth:    nil,
			// Site:                  nil, // Replace with the appropriate default value or initialization for the expected type
			// Accounts:              types.StringNull().ValueString(),
		})
	}
	state.Licenses = licenses
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
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
