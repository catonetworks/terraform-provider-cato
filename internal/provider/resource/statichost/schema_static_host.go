package statichost

import (
	"context"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *staticHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_static_host` resource contains the configuration parameters necessary to add a static host. " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.addStaticHost()](https://api.catonetworks.com/documentation/#mutation-site.addStaticHost).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID (Host's parent). Exactly one of `site_id` or `site_name` must be set. " +
					"When `site_name` is used, `site_id` is resolved via entityLookup and written to state (Computed).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIf(
						siteIDRequiresReplaceIf,
						"Replace only when the parent site id changes to a different known value.",
						"Replace only when the parent site id changes to a different known value.",
					),
				},
			},
			"site_name": schema.StringAttribute{
				Description: "Site name (Host's parent), an alternative to `site_id`. Resolved to the site's ID " +
					"via entityLookup at apply (site names are unique). Exactly one of `site_id` or `site_name` must be set.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						siteNameRequiresReplaceIf,
						"Replace only when site_name changes between two concrete sites.",
						"Replace only when site_name changes between two concrete sites.",
					),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("site_id"), path.MatchRoot("site_name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Host name",
				Required:    true,
			},
			"ip": schema.StringAttribute{
				Description: "Host IP address",
				Required:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "Host MAC address (for DHCP reservervation)",
				Optional:    true,
			},
		},
	}
}

// siteIDRequiresReplaceIf forces replacement only when the parent site id
// actually changes to a different KNOWN value. During site_name adoption the
// planned site_id is carried from prior state (UseStateForUnknown), so old ==
// new and no replacement occurs -- the host is updated in place instead. A null
// or unknown planned id (e.g. still to be resolved from site_name at apply)
// never triggers a replace.
func siteIDRequiresReplaceIf(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		resp.RequiresReplace = false
		return
	}
	resp.RequiresReplace = !req.PlanValue.Equal(req.StateValue)
}

// siteNameRequiresReplaceIf forces replacement only when site_name changes from
// one concrete site to a DIFFERENT concrete site. First-time adoption (prior
// state site_name null -> value, for a host whose state only carried site_id)
// and switching back to the site_id form (value -> null) keep the host on the
// same site, so they update in place rather than destroy-and-recreate (which
// the API refuses while the host is referenced by a policy rule).
func siteNameRequiresReplaceIf(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		resp.RequiresReplace = false
		return
	}
	resp.RequiresReplace = !req.PlanValue.Equal(req.StateValue)
}
