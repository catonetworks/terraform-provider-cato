package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppConnectorModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	SerialNumber         types.String `tfsdk:"serial_number"`
	Type                 types.String `tfsdk:"type"`
	SocketModel          types.String `tfsdk:"socket_model"`
	GroupName            types.String `tfsdk:"group"`
	PostalAddress        types.Object `tfsdk:"address"` // PostalAddressModel
	Timezone             types.String `tfsdk:"timezone"`
	PreferredPopLocation types.Object `tfsdk:"preferred_pop_location"` // PreferredPopLocationModel
}

type PostalAddressModel struct {
	AddressValidated types.String `tfsdk:"address_validated"`
	City             types.String `tfsdk:"city"`
	Country          types.Object `tfsdk:"country"` // IdNameRefModel
	State            types.String `tfsdk:"state"`
	Street           types.String `tfsdk:"street"`
	ZipCode          types.String `tfsdk:"zip_code"`
}

var PostalAddressModelTypes = map[string]attr.Type{
	"address_validated": types.StringType,
	"city":              types.StringType,
	"country":           types.ObjectType{AttrTypes: IdNameRefModelTypes},
}

type IdNameRefModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var IdNameRefModelTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

type PreferredPopLocationModel struct {
	PreferredOnly types.Bool   `tfsdk:"preferred_only"`
	Automatic     types.Bool   `tfsdk:"automatic"`
	Primary       types.Object `tfsdk:"primary"`   // IdNameRefModel
	Secondary     types.Object `tfsdk:"secondary"` // IdNameRefModel
}

var PreferredPopLocationModelTypes = map[string]attr.Type{
	"preferred_only": types.BoolType,
	"automatic":      types.BoolType,
	"primary":        types.ObjectType{AttrTypes: IdNameRefModelTypes},
	"secondary":      types.ObjectType{AttrTypes: IdNameRefModelTypes},
}
