package cloudinterconnectsite

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var locationAttrTypes = map[string]attr.Type{
	"country_code": types.StringType, "timezone": types.StringType, "city": types.StringType,
	"state_code": types.StringType, "address": types.StringType,
}

func (r *cloudSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages Cloud Interconnect site metadata. " +
		"Physical connections are managed separately; deletion is refused while any " +
		"connection exists. Import using its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Description: "Name", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"description": schema.StringAttribute{Description: "Description. Removing the value clears it.", Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()}},
			"site_type": schema.StringAttribute{Description: "Site type", Required: true,
				Validators: []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllSiteType)...)}},
			"site_location": schema.SingleNestedAttribute{Description: "Site location", Required: true, Attributes: schemaLocation()},
		}}
}
func schemaLocation() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"country_code": schema.StringAttribute{Description: "Country code", Required: true},
		"timezone":     schema.StringAttribute{Description: "Timezone", Required: true},
		"city": schema.StringAttribute{Description: "City. Omission uses the API default or current value.",
			Optional: true, Computed: true},
		"state_code": schema.StringAttribute{Description: "State code. Omission uses the API default or current value.",
			Optional: true, Computed: true},
		"address": schema.StringAttribute{Description: "Address. Omission uses the API default or current value.",
			Optional: true, Computed: true},
	}
}
