package customapplication

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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var rangeAttrTypes = convert.FromToAttrTypes
var ipAttrTypes = map[string]attr.Type{
	"ip": types.SetType{ElemType: types.StringType}, "subnet": types.SetType{ElemType: types.StringType},
	"ip_range": types.SetType{ElemType: types.ObjectType{AttrTypes: rangeAttrTypes}},
}
var destinationAttrTypes = map[string]attr.Type{
	"domain": types.SetType{ElemType: types.StringType}, "fqdn": types.SetType{ElemType: types.StringType},
	"destination_ip": types.ObjectType{AttrTypes: ipAttrTypes},
}
var criteriaAttrTypes = map[string]attr.Type{
	"protocol": types.StringType, "port": types.SetType{ElemType: types.StringType},
	"port_range":  types.SetType{ElemType: types.ObjectType{AttrTypes: rangeAttrTypes}},
	"destination": types.ObjectType{AttrTypes: destinationAttrTypes},
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages an account-scoped custom application with unordered OR criteria. " +
		"Changes are published by the API. Import using the exact application ID.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name": schema.StringAttribute{Description: "Name", Required: true},
		"description": schema.StringAttribute{Description: "Description", Optional: true, Computed: true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()}},
		"category": schema.SetNestedAttribute{Description: "Category", Optional: true, Computed: true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Category")},
			PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.ObjectType{AttrTypes: idname.RefAttrTypes})}},
		"criteria": schema.SetNestedAttribute{Description: "Unordered OR criteria; at least one meaningful criterion is required.",
			Required: true,

			NestedObject: schema.NestedAttributeObject{Attributes: schemaCriteria()}},
	}}
}
func schemaCriteria() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"protocol": schema.StringAttribute{Description: "Protocol", Required: true,
			Validators: []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllIPProtocol)...)}},
		"port": schema.SetAttribute{Description: "Canonical decimal ports from 0 through 65535; use [] if unused.",
			Required:    true,
			ElementType: types.StringType},

		"port_range": schema.SetNestedAttribute{Description: "Port ranges; use [] if unused.",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaRange()}},

		"destination": schema.SingleNestedAttribute{Description: "Destination", Optional: true, Attributes: schemaDestination()},
	}
}
func schemaRange() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{Description: "From", Required: true}, "to": schema.StringAttribute{Description: "To", Required: true},
	}
}
func schemaStrings(description string) schema.SetAttribute {
	return schema.SetAttribute{Description: description, Optional: true, Computed: true,
		ElementType: types.StringType, PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.StringType)}}
}
func schemaDestination() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"domain": schemaStrings("Domain"), "fqdn": schemaStrings("Fqdn"),
		"destination_ip": schema.SingleNestedAttribute{Description: "Destination IP criteria",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"ip": schemaStrings("IP addresses"), "subnet": schemaStrings("Subnets"),
				"ip_range": schema.SetNestedAttribute{Description: "IP ranges", Optional: true, Computed: true,
					NestedObject:  schema.NestedAttributeObject{Attributes: schemaRange()},
					PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.ObjectType{AttrTypes: rangeAttrTypes})}},
			}},
	}
}
