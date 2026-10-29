package cloudinterconnectphysicalconnection

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var qinqAttrTypes = map[string]attr.Type{"c_vlan": types.Int64Type, "s_vlan": types.Int64Type}

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one manual Cloud Interconnect physical connection. " +
		"Site and HA role changes require replacement. Import using the opaque connection ID.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"ha_role": schema.StringAttribute{Description: "HA role",
			Required:      true,
			Validators:    []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllHaRole)...)},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"encapsulation_method": schema.StringAttribute{
			Description: "Encapsulation method. Changes require explicit replacement because the API cannot clear incompatible previous tags.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllTaggingMethod)...)}},
		"downstream_bw_limit": schema.StringAttribute{Description: "Downstream bandwidth as a canonical non-negative decimal integer.",
			Required: true},
		"upstream_bw_limit": schema.StringAttribute{Description: "Upstream bandwidth as a canonical non-negative decimal integer.",
			Required: true},
		"private_cato_ip": schema.StringAttribute{Description: "Cato IPv4 address within subnet",
			Required: true},
		"private_site_ip": schema.StringAttribute{Description: "Site IPv4 address within subnet, distinct from the Cato address.",
			Required: true},
		"subnet": schema.StringAttribute{Description: "Canonical IPv4 subnet in CIDR notation",
			Required: true},
		"service_provider_name": schema.StringAttribute{Description: "Service provider supported by the selected PoP",
			Required: true},
		"site": schema.SingleNestedAttribute{Description: "Parent Cloud Interconnect site",
			Required: true, Attributes: idname.SchemaNameID("Site"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()}},
		"pop_location": schema.SingleNestedAttribute{Description: "PoP location",
			Required: true, Attributes: idname.SchemaNameID("PoP"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()}},
		"vlan": schema.Int64Attribute{Description: "DOT1Q VLAN. Omit for API allocation; custom tags require the account feature.",
			Optional: true, Computed: true,
			Validators: []validator.Int64{int64validator.Between(minVLAN, maxVLAN)}},
		"qinq_vlan_configuration": schema.SingleNestedAttribute{
			Description: "QINQ tags. Omit for API allocation; custom tags require the account feature.",
			Optional:    true, Computed: true,
			Attributes: map[string]schema.Attribute{
				"c_vlan": schema.Int64Attribute{Description: "Customer VLAN",
					Required: true, Validators: []validator.Int64{int64validator.Between(minVLAN, maxVLAN)}},
				"s_vlan": schema.Int64Attribute{Description: "Service VLAN",
					Required: true, Validators: []validator.Int64{int64validator.Between(minVLAN, maxVLAN)}},
			}},
	}}
}
