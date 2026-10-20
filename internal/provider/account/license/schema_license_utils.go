package license

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:funlen // Terraform schemas are declarative and lengthy by nature.
func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_license` resource contains the configuration parameters necessary to assign and " +
			"replace licenses for sites. When creating a new license resource, the `site_id` and `license_id` " +
			"attributes are required. The `license_info` attribute is optional and will be populated with the " +
			"license information after the resource is created. If the site has an existing license assigned, " +
			"the license resource will call the mutation.ReplaceSiteBwLicense() operation replacing the existing " +
			"license with the new license_id ensuring there is no interruption in servive of reassigning a license." +
			"\n\n**NOTE** License assignment does not work for Trial accounts, or for accounts that are \"not synced\".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "License ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"license_id": schema.StringAttribute{
				Description: "License ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bw": schema.Int64Attribute{
				Description: "Bandwidth to allocate to site (only used for pooled license model)",
				Required:    false,
				Optional:    true,
				Validators: []validator.Int64{
					customInt64Validator{},
				},
			},
			"license_info": schema.SingleNestedAttribute{
				Description: "",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(), // avoid drift on refresh
				},
				Attributes: map[string]schema.Attribute{
					"sku": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"plan": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"status": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"expiration_date": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"start_date": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"last_updated": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"total": schema.Int64Attribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"site_license_group": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"site_license_type": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"allocated_bandwidth": schema.Int64Attribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}
