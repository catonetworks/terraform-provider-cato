package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppConnectorModel struct {
	Description          types.String `tfsdk:"description"`
	GroupName            types.String `tfsdk:"group_name"`
	ID                   types.String `tfsdk:"id"`
	Location             types.Object `tfsdk:"location"` // AppConnectorLocation
	Name                 types.String `tfsdk:"name"`
	PreferredPopLocation types.Object `tfsdk:"preferred_pop_location"` // PreferredPopLocationModel
	PrivateAppRef        types.Set    `tfsdk:"private_apps"`           // []IDNameRefModel
	SerialNumber         types.String `tfsdk:"serial_number"`
	SocketID             types.String `tfsdk:"socket_id"`
	SocketModel          types.String `tfsdk:"socket_model"`
	Type                 types.String `tfsdk:"type"`
}

type AppConnectorLocation struct {
	Address     types.String `tfsdk:"address"`
	CityName    types.String `tfsdk:"city_name"`
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
}

var AppConnectorLocationTypes = map[string]attr.Type{
	"address":      types.StringType,
	"city_name":    types.StringType,
	"country_code": types.StringType,
	"state_code":   types.StringType,
	"timezone":     types.StringType,
}

type PostalAddressModel struct {
	AddressValidated types.String `tfsdk:"address_validated"`
	City             types.String `tfsdk:"city"`
	Country          types.Object `tfsdk:"country"` // IDNameRefModel
	State            types.String `tfsdk:"state"`
	Street           types.String `tfsdk:"street"`
	ZipCode          types.String `tfsdk:"zip_code"`
}
