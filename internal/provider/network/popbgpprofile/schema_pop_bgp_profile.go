package popbgpprofile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

var communityAttrTypes = map[string]attr.Type{"from": types.StringType, "to": types.StringType}
var communityObjectType = types.ObjectType{AttrTypes: communityAttrTypes}

func (r *bgpProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one account-scoped PoP BGP community profile. " +
		"This resource does not attach the profile to a BGP session. Import using its ID.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name": schema.StringAttribute{Description: "Name", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"description": schema.StringAttribute{Description: "Description. Removing the value clears it.", Optional: true, Computed: true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()}},
		"communities": schema.SetNestedAttribute{
			Description: "Unordered community pairs. Values are canonical decimal strings from 0 through 65535.",
			Required:    true, Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{Description: "From", Required: true},
				"to":   schema.StringAttribute{Description: "To", Required: true},
			}},
		},
	}}
}
