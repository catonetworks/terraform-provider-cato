package provider

import (
	"context"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &appConnectorResource{}
	_ resource.ResourceWithConfigure   = &appConnectorResource{}
	_ resource.ResourceWithImportState = &appConnectorResource{}
)

const optional = true

func NewAppConnectorResource() resource.Resource {
	return &appConnectorResource{}
}

type appConnectorResource struct {
	client *catoClientData
}

func (r *appConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_connector"
}

func (r *appConnectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_app_connector` resource contains the configuration parameters necessary to manage a group. Groups can contain various member types including sites, hosts, network ranges, and more. Documentation for the underlying API used in this resource can be found at [mutation.groups.createGroup()](https://api.catonetworks.com/documentation/#mutation-groups.createGroup).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the ZTNA App Connector",
				Computed:    true,
			},

			"name": schema.StringAttribute{
				Description: "The unique name of the ZTNA App Connector",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the ZTNA App Connector (max 250 characters)",
				Optional:    true,
			},
			"serial_number": schema.StringAttribute{
				Description: "Unique serial number of the ZTNA App Connector",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Connector type (virtual, physical)",
				Required:    true,
				Validators:  []validator.String{appConTypeValidator{}},
			},
			"socket_model": schema.StringAttribute{
				Description: "Socket model of the ZTNA App Connector",
				Optional:    true,
			},
			"group": schema.StringAttribute{
				Description: "Name of the ZTNA App Connector group",
				Required:    true,
			},
			"address": resSchemaPostalAddress("Physical location of the connector"),
			"timezone": schema.StringAttribute{
				Description: "Time zone",
				Required:    true,
			},
			"preferred_pop_location": schema.SingleNestedAttribute{
				Description: "Preferred PoP locations settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"preferred_only": schema.BoolAttribute{
						Description: "Restrict connector attachment exclusively to the configured pop locations",
						Required:    true,
					},
					"automatic": schema.BoolAttribute{
						Description: "Automatic PoP location",
						Required:    true,
					},
					"primary": schema.SingleNestedAttribute{
						Description: "Physical location of the pripary Pop",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "Location name",
								Optional:    true,
								Computed:    true,
							},
							"id": schema.StringAttribute{
								Description: "Location ID",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"secondary": schema.SingleNestedAttribute{
						Description: "Physical location of the secondary Pop",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "Location name",
								Optional:    true,
								Computed:    true,
							},
							"id": schema.StringAttribute{
								Description: "Location ID",
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

func resSchemaPostalAddress(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"address_validated": schema.StringAttribute{
				Description: "Address validation status",
				Computed:    true,
			},
			"city": schema.StringAttribute{
				Description: "City name",
				Required:    true,
			},
			"country": schema.SingleNestedAttribute{
				Description: "Country name or id",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Country name",
						Optional:    true,
						Computed:    true,
					},
					"id": schema.StringAttribute{
						Description: "Country code",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"state": schema.StringAttribute{
				Description: "State name (required for the USA)",
				Optional:    true,
			},
			"street": schema.StringAttribute{
				Description: "Street name and number",
				Optional:    true,
			},
			"zip_code": schema.StringAttribute{
				Description: "Zip code",
				Optional:    true,
			},
		},
	}
}

func (r *appConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *appConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *appConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppConnectorModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.AddZtnaAppConnectorInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		GroupName:   plan.GroupName.ValueString(),
		Type:        cato_models.ZtnaAppConnectorType(plan.Type.ValueString()),
		Address: &cato_models.PostalAddressInput{
			CityName:  plan.PostalAddress.City.ValueStringPointer(),
			StateName: plan.PostalAddress.State.ValueStringPointer(),
			Street:    plan.PostalAddress.Street.ValueStringPointer(),
			ZipCode:   plan.PostalAddress.ZipCode.ValueStringPointer(),
		},
		Timezone: plan.Timezone.ValueString(),
		PreferredPopLocation: &cato_models.ZtnaAppConnectorPreferredPopLocationInput{
			PreferredOnly: plan.PreferredPopLocation.PreferredOnly.ValueBool(),
			Automatic:     plan.PreferredPopLocation.Automatic.ValueBool(),
		},
	}

	// Country
	refBy, refInput, _ := prepareIdName(plan.PostalAddress.Country.ID, plan.PostalAddress.Country.Name, &resp.Diagnostics, "address.country")
	if resp.Diagnostics.HasError() {
		return
	}
	input.Address.Country = &cato_models.CountryRefInput{By: refBy, Input: refInput}

	// Primary pop location
	refBy, refInput, idNameSet := prepareIdName(plan.PreferredPopLocation.Primary.ID, plan.PreferredPopLocation.Primary.Name,
		&resp.Diagnostics, "primary preferred_pop_location", optional)
	if resp.Diagnostics.HasError() {
		return
	}
	if idNameSet {
		input.PreferredPopLocation.Primary = &cato_models.PopLocationRefInput{By: refBy, Input: refInput}
	}

	// Secondary pop location
	refBy, refInput, idNameSet = prepareIdName(plan.PreferredPopLocation.Secondary.ID, plan.PreferredPopLocation.Secondary.Name,
		&resp.Diagnostics, "secondary preferred_pop_location", optional)
	if resp.Diagnostics.HasError() {
		return
	}
	if idNameSet {
		input.PreferredPopLocation.Secondary = &cato_models.PopLocationRefInput{By: refBy, Input: refInput}
	}

	// Call Cato API to create a new connector
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.AppConnectorCreateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorCreateConnector error", err.Error())
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(result.GetZtnaAppConnector().GetAddZtnaAppConnector().GetZtnaAppConnector().GetID())

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateAppConnectorState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating appConnector state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func IdNameInput(id, name types.String) (isSet bool, by cato_models.ObjectRefBy, input string, err error) {
	if id.IsUnknown() && name.IsUnknown() {
		return false, "", "", nil // not set
	}
	if !id.IsUnknown() && !name.IsUnknown() {
		return false, "", "", fmt.Errorf("Only one of 'id' or 'name' can be specified")
	}

	if !id.IsUnknown() {
		return true, cato_models.ObjectRefByID, id.ValueString(), nil
	}
	return true, cato_models.ObjectRefByName, name.ValueString(), nil
}

// prepareIdName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func prepareIdName(id, name types.String, diags *diag.Diagnostics, fieldName string, optional ...bool) (by cato_models.ObjectRefBy, input string, isSet bool) {
	idNameSet, by, input, err := IdNameInput(id, name)
	if err != nil {
		diags.AddError("invalid configuration of "+fieldName, err.Error())
		return
	}

	if idNameSet {
		return by, input, true
	}

	// not set and it is mandatory
	if len(optional) == 0 || (!optional[0]) {
		diags.AddError("missing configuration of "+fieldName, "id or name must be set on "+fieldName)
	}

	return by, input, false
}

func (r *appConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppConnectorModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AppConnectorModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := strings.Trim(state.ID.String(), `"`)
	if id == "" {
		resp.Diagnostics.AddError("AppConnectorUpdateConnector: ID in unknown", "AppConnector ID is not set in TF state")
		return
	}
	input := cato_models.UpdateZtnaAppConnectorInput{
		ID:          id,
		Name:        plan.Name.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		GroupName:   plan.GroupName.ValueStringPointer(),
		// Address:
		Timezone:             plan.Timezone.ValueStringPointer(),
		PreferredPopLocation: nil,
	}

	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.AppConnectorUpdateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorUpdateConnector error", err.Error())
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateAppConnectorState(ctx, id, plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating app-connector state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppConnectorModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, hydrateErr := r.hydrateAppConnectorState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		// Check if app-connector not found
		if hydrateErr.Error() == "app_connector not found" { // TODO: check the actual error
			tflog.Warn(ctx, "app_connector not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppConnectorModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := cato_models.RemoveZtnaAppConnectorInput{
		ZtnaAppConnector: &cato_models.ZtnaAppConnectorRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.AppConnectorDeleteConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorDeleteConnector error", err.Error())
		return
	}
}

// hydrateAppConnectorState fetches the current state of a appConnector from the API
// It takes a plan parameter to match config members with API members correctly
func (r *appConnectorResource) hydrateAppConnectorState(ctx context.Context, appConnectorID string, plan AppConnectorModel) (*AppConnectorModel, error) {
	input := cato_models.ZtnaAppConnectorRefInput{
		By:    cato_models.ObjectRefByID,
		Input: appConnectorID,
	}

	// Call Cato API to get a connector
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.AppConnectorReadConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}

	// Map API response to AppConnectorModel
	con := result.GetZtnaAppConnector().GetZtnaAppConnector()
	addr := con.Address
	pref := con.PreferredPopLocation
	state := &AppConnectorModel{
		ID:   types.StringValue(con.ID),
		Name: types.StringValue(con.Name),
		// Description:  types.StringPointerValue(c.Description), TODO: add description
		SerialNumber: types.StringPointerValue(con.SerialNumber),
		Type:         types.StringValue(con.Type.String()),
		// SocketModel:  types.StringPointerValue(c.SocketModel), // TODO: add socket model
		GroupName: types.StringValue(con.GroupName),
		PostalAddress: &PostalAddressModel{
			AddressValidated: types.StringValue(addr.AddressValidated.String()),
			City:             types.StringPointerValue(addr.GetCityName()),
			Country: IdNameRefModel{
				ID:   types.StringValue(addr.Country.ID),
				Name: types.StringValue(addr.Country.Name),
			},
			State:   types.StringPointerValue(addr.StateName),
			Street:  types.StringPointerValue(addr.Street),
			ZipCode: types.StringPointerValue(addr.ZipCode),
		},
		Timezone: types.StringValue(con.Timezone),
		PreferredPopLocation: &PreferredPopLocationModel{
			PreferredOnly: types.BoolValue(pref.PreferredOnly),
			Automatic:     types.BoolValue(pref.Automatic),
		},
	}
	if pref.Primary != nil {
		state.PreferredPopLocation.Primary = IdNameRefModel{
			ID:   types.StringValue(pref.Primary.ID),
			Name: types.StringValue(pref.Primary.Name),
		}
	}
	if pref.Secondary != nil {
		state.PreferredPopLocation.Secondary = IdNameRefModel{
			ID:   types.StringValue(pref.Secondary.ID),
			Name: types.StringValue(pref.Secondary.Name),
		}
	}

	return state, nil
}

func ptr[T any](x T) *T { return &x }

type appConTypeValidator struct{}

func (v appConTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := strings.Trim(req.ConfigValue.String(), `"`)
	connType := cato_models.ZtnaAppConnectorType(value)
	if !connType.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid connector type (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllZtnaAppConnectorType))
		return
	}
}
func (v appConTypeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("AppConnector type must be one of: %v", cato_models.AllZtnaAppConnectorType)
}
func (v appConTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
