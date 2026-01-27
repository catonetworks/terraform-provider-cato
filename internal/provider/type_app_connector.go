package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TODO: merge with datasource models?
type AppConnectorModel struct {
	ID                   types.String               `tfsdk:"id"`
	Name                 types.String               `tfsdk:"name"`
	Description          types.String               `tfsdk:"description"`
	SerialNumber         types.String               `tfsdk:"serial_number"`
	Type                 types.String               `tfsdk:"type"`
	SocketModel          types.String               `tfsdk:"socket_model"`
	GroupName            types.String               `tfsdk:"group"`
	PostalAddress        *PostalAddressModel        `tfsdk:"address"`
	Timezone             types.String               `tfsdk:"timezone"`
	PreferredPopLocation *PreferredPopLocationModel `tfsdk:"preferred_pop_location"`
}

type PostalAddressModel struct {
	AddressValidated types.String   `tfsdk:"address_validated"`
	City             types.String   `tfsdk:"city"`
	Country          IdNameRefModel `tfsdk:"country"`
	State            types.String   `tfsdk:"state"`
	Street           types.String   `tfsdk:"street"`
	ZipCode          types.String   `tfsdk:"zip_code"`
}

type IdNameRefModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type PreferredPopLocationModel struct {
	PreferredOnly types.Bool     `tfsdk:"preferred_only"`
	Automatic     types.Bool     `tfsdk:"automatic"`
	Primary       idNameRefModel `tfsdk:"primary"`
	Secondary     idNameRefModel `tfsdk:"secondary"`
}
