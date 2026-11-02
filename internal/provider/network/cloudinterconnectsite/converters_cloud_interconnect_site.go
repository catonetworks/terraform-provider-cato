package cloudinterconnectsite

import (
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

func buildCreate(plan resourceModel, diags *diag.Diagnostics) m.AddCloudInterconnectSiteInput {
	validateFields(plan, diags)
	attrs := locationValues(plan.SiteLocation, diags)
	input := m.AddCloudInterconnectSiteInput{Name: plan.Name.ValueString(), SiteType: m.SiteType(plan.SiteType.ValueString()),
		Description: convert.StringPointerForOptionalInput(plan.Description)}
	if attrs == nil {
		return input
	}
	input.SiteLocation = &m.AddSiteLocationInput{
		CountryCode: attrs["country_code"].(types.String).ValueString(), Timezone: attrs["timezone"].(types.String).ValueString(),
		City: convert.KnownStringPointer(attrs["city"].(types.String)), StateCode: convert.KnownStringPointer(attrs["state_code"].(types.String)),
		Address: convert.KnownStringPointer(attrs["address"].(types.String)),
	}
	return input
}
func buildUpdate(plan resourceModel, diags *diag.Diagnostics) m.UpdateSiteGeneralDetailsInput {
	input := buildCreate(plan, diags)
	description := plan.Description.ValueString()
	result := m.UpdateSiteGeneralDetailsInput{Name: &input.Name, Description: &description, SiteType: &input.SiteType}
	if input.SiteLocation != nil {
		location := input.SiteLocation
		result.SiteLocation = &m.UpdateSiteLocationInput{CountryCode: &location.CountryCode, Timezone: &location.Timezone,
			CityName: location.City, StateCode: location.StateCode, Address: location.Address}
	}
	return result
}
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	if plan.Description.IsNull() {
		plan.Description = types.StringValue("")
	}
	attrs := plan.SiteLocation.Attributes()
	for key, value := range attrs {
		if value.IsUnknown() {
			attrs[key] = types.StringValue("")
		}
	}
	plan.SiteLocation = types.ObjectValueMust(locationAttrTypes, attrs)
	return plan
}
func parseSite(item *siteRecord, diags *diag.Diagnostics) resourceModel {
	location := item.GetSiteLocation()
	object, d := types.ObjectValue(locationAttrTypes, map[string]attr.Value{
		"country_code": types.StringValue(location.GetCountryCode()), "timezone": types.StringValue(location.GetTimezone()),
		"city": convert.ParseOptionalString(location.GetCityName()), "state_code": convert.ParseOptionalString(location.GetStateCode()),
		"address": convert.ParseOptionalString(location.GetAddress()),
	})
	diags.Append(d...)
	result := resourceModel{ID: types.StringValue(item.GetSite().GetID()), Name: types.StringValue(item.GetSite().GetName()),
		Description: convert.ParseOptionalString(item.GetDescription()),
		SiteType:    types.StringValue(string(*item.GetSiteType())), SiteLocation: object}
	validateFields(result, diags)
	locationValues(result.SiteLocation, diags)
	return result
}
