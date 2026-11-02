package fqdncontainer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	containerhelpers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/container"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

var refType = containerhelpers.ReferenceType
var notificationsAttrTypes = containerhelpers.NotificationAttributeTypes
var syncAttrTypes = containerhelpers.SyncAttributeTypes

func (r *containerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	refs := func(description string) schema.SetNestedAttribute {
		return schema.SetNestedAttribute{Description: description,
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID(description)}}
	}
	resp.Schema = schema.Schema{Description: "Manages an account-scoped FQDN container. INLINE owns all members, including an empty " +
		"set. URL owns synchronization settings, not remotely fetched members. Name and mode " +
		"changes require replacement. Changes are published by the API. Import by exact " +
		"container ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Container ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},

			"name": schema.StringAttribute{Description: "Container name; changing it requires replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},

			"mode": schema.StringAttribute{Description: "INLINE or URL; changing it requires replacement.",
				Required: true,
				Validators: []validator.String{stringvalidator.OneOf("INLINE",
					"URL")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},

			"description": schema.StringAttribute{Description: "Container description", Required: true},
			"size": schema.Int64Attribute{Description: "Number of members reported by the API, which may change after URL synchronization.",
				Computed: true},

			"values": schema.SetAttribute{Description: "Complete set of canonical lowercase FQDNs, required in INLINE mode. Use [] for no " +
				"members. Conflicts with file_type and sync_data.",
				Optional:    true,
				ElementType: types.StringType},

			"file_type": schema.StringAttribute{Description: "CSV or STIX, required only in URL mode.",
				Optional: true,
				Validators: []validator.String{stringvalidator.OneOf("CSV",
					"STIX")}},

			"sync_data": schema.SingleNestedAttribute{Description: "Required only in URL mode. Synchronization and notification settings.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"time_interval": schema.Int64Attribute{Description: "Positive synchronization interval", Required: true},
					"time_unit": schema.StringAttribute{Description: "HOUR or DAY",
						Required: true,
						Validators: []validator.String{stringvalidator.OneOf("HOUR",
							"DAY")}},

					"url": schema.StringAttribute{Description: "HTTPS source URL without user credentials or fragments.", Required: true},
					"notifications": schema.SingleNestedAttribute{Description: "Exactly one destination type must contain references. " +
						"Use [] for the " +
						"other types.",
						Required: true,
						Attributes: map[string]schema.Attribute{"mailing_list": refs("Mailing lists"),
							"subscription_group": refs("Subscription groups"),
							"webhook":            refs("Webhooks")}},
				}}}}
}
