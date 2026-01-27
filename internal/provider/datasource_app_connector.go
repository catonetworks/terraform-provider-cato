package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppConnectorDataSourceModel struct {
	NameFilter types.List            `tfsdk:"name_filter"`
	Items      []appConnectorDSModel `tfsdk:"items"`
}

type appConnectorDSModel struct {
	ID                   types.String              `tfsdk:"id"`
	Name                 types.String              `tfsdk:"name"`
	Description          types.String              `tfsdk:"description"`
	SerialNumber         types.String              `tfsdk:"serial_number"`
	Type                 types.String              `tfsdk:"type"`
	SocketModel          types.String              `tfsdk:"socket_model"`
	GroupName            types.String              `tfsdk:"group"`
	PostalAddress        postalAddressModel        `tfsdk:"address"`
	Timezone             types.String              `tfsdk:"timezone"`
	PreferredPopLocation preferredPopLocationModel `tfsdk:"preferred_pop_location"`
}

type postalAddressModel struct {
	AddressValidated types.String   `tfsdk:"address_validated"`
	CityName         types.String   `tfsdk:"city"`
	CountryRef       idNameRefModel `tfsdk:"country"`
	StateName        types.String   `tfsdk:"state"`
	Street           types.String   `tfsdk:"street"`
	ZipCode          types.String   `tfsdk:"zip_code"`
}

type idNameRefModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type preferredPopLocationModel struct {
	PreferredOnly types.Bool     `tfsdk:"preferred_only"`
	Automatic     types.Bool     `tfsdk:"automatic"`
	Primary       idNameRefModel `tfsdk:"primary"`
	Secondary     idNameRefModel `tfsdk:"secondary"`
}

func dsSchemaRefIDName(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name",
				Computed:    true,
			},
		},
	}
}

func dsSchemaPostalAddress(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"address_validated": schema.StringAttribute{
				Description: "Address validation status",
				Computed:    true,
			},
			"city": schema.StringAttribute{
				Description: "City",
				Computed:    true,
			},
			"country": schema.SingleNestedAttribute{
				Description: "Country name and code",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Country name",
						Computed:    true,
					},
					"id": schema.StringAttribute{
						Description: "Country code",
						Computed:    true,
					},
				},
			},
			// TODO: change to reference ID/Name
			"state": schema.StringAttribute{
				Description: "State name",
				Computed:    true,
			},
			"street": schema.StringAttribute{
				Description: "Street name and number",
				Computed:    true,
			},
			"zip_code": schema.StringAttribute{
				Description: "Zip code",
				Computed:    true,
			},
		},
	}
}

func AppConnectorDataSource() datasource.DataSource {
	return &appConnectorDataSource{}
}

type appConnectorDataSource struct {
	client *catoClientData
}

func (d *appConnectorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_connector"
}

func (d *appConnectorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "AppConnector schema definition",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of AppConnector names to filter",
				Optional:    true,
			},

			"items": schema.ListNestedAttribute{
				Description: "List of AppConnectors",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique ID of the ZTNA App Connector",
							Computed:    true,
						},

						"name": schema.StringAttribute{
							Description: "The unique name of the ZTNA App Connector",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Optional description of the ZTNA App Connector (max 250 characters)",
							Computed:    true,
						},
						"serial_number": schema.StringAttribute{
							Description: "Unique serial number of the ZTNA App Connector",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Connector type (virtual, physical)",
							Computed:    true,
						},
						"socket_model": schema.StringAttribute{
							Description: "Socket model of the ZTNA App Connector",
							Computed:    true,
						},
						"group": schema.StringAttribute{
							Description: "Name of the ZTNA App Connector group",
							Computed:    true,
						},
						"address": dsSchemaPostalAddress("Physical location of the connector"),

						"timezone": schema.StringAttribute{
							Description: "Time zone",
							Computed:    true,
						},
						"preferred_pop_location": schema.SingleNestedAttribute{
							Description: "Preferred PoP locations settings",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"preferred_only": schema.BoolAttribute{
									Description: "Restrict connector attachment exclusively to the configured pop locations",
									Computed:    true,
								},
								"automatic": schema.BoolAttribute{
									Description: "Automatic PoP location",
									Computed:    true,
								},
								"primary":   dsSchemaRefIDName("Physical location of the pripary Pop"),
								"secondary": dsSchemaRefIDName("Physical location of the secondary Pop"),
							},
						},
					},
				},
			},
		},
	}
}

func (d *appConnectorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoClientData)
}

func (d *appConnectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var connectors AppConnectorDataSourceModel
	if diags := req.Config.Get(ctx, &connectors); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	acs, err := fetchConnectors()
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch app-connectors", err.Error())
		return
	}

	for _, c := range acs {
		con := appConnectorDSModel{
			ID:           types.StringValue(c.ID),
			Name:         types.StringValue(c.Name),
			Description:  types.StringValue(""),
			SerialNumber: types.StringPointerValue(c.SerialNumber),
			Type:         types.StringValue(c.Type),
			// SocketModel:  types.StringValue(*c.SocketModel),
			GroupName: types.StringValue(c.GroupName),
			PostalAddress: postalAddressModel{
				AddressValidated: types.StringValue(c.Address.AddressValidated),
				CityName:         types.StringValue(c.Address.CityName),
				CountryRef: idNameRefModel{
					ID:   types.StringValue(c.Address.Country.ID),
					Name: types.StringValue(c.Address.Country.Name),
				},
				StateName: types.StringPointerValue(c.Address.StateName),
				// Street:    types.StringValue(*c.Address.Street),
				// ZipCode:   types.StringValue(*c.Address.ZipCode),
			},
			Timezone: types.StringValue(c.Timezone),
			PreferredPopLocation: preferredPopLocationModel{
				PreferredOnly: types.BoolValue(c.PreferredPopLocation.PreferredOnly),
				Automatic:     types.BoolValue(false),
			},
		}
		if c.PreferredPopLocation.Primary != nil {
			con.PreferredPopLocation.Primary = idNameRefModel{
				ID:   types.StringValue(c.PreferredPopLocation.Primary.ID),
				Name: types.StringValue(c.PreferredPopLocation.Primary.Name),
			}
		}
		if c.PreferredPopLocation.Secondary != nil {
			con.PreferredPopLocation.Secondary = idNameRefModel{
				ID:   types.StringValue(c.PreferredPopLocation.Secondary.ID),
				Name: types.StringValue(c.PreferredPopLocation.Secondary.Name),
			}
		}
		connectors.Items = append(connectors.Items, con)
	}

	if diags := resp.State.Set(ctx, &connectors); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
