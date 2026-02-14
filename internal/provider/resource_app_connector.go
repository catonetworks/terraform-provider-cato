package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &appConnectorResource{}
	_ resource.ResourceWithConfigure   = &appConnectorResource{}
	_ resource.ResourceWithImportState = &appConnectorResource{}

	ErrAppConnectorNotFound = errors.New("app-connector not found")
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

func (r *appConnectorResource) prepareAddress(ctx context.Context, addr types.Object, diags *diag.Diagnostics) *cato_models.PostalAddressInput {
	if !hasValue(addr) {
		return nil
	}

	var tfAddress PostalAddressModel
	diags.Append(addr.As(ctx, &tfAddress, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkAddress := cato_models.PostalAddressInput{
		CityName:  tfAddress.City.ValueStringPointer(),
		Country:   prepareIDRef[cato_models.CountryRefInput](ctx, tfAddress.Country, diags, "address.country"),
		StateName: tfAddress.State.ValueStringPointer(),
		Street:    tfAddress.Street.ValueStringPointer(),
		ZipCode:   tfAddress.ZipCode.ValueStringPointer(),
	}

	return &sdkAddress
}

func (r *appConnectorResource) preparePopLocation(ctx context.Context, loc types.Object, diags *diag.Diagnostics,
) *cato_models.ZtnaAppConnectorPreferredPopLocationInput {
	if !hasValue(loc) {
		return nil
	}

	var tfLocation PreferredPopLocationModel
	diags.Append(loc.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkLocation := cato_models.ZtnaAppConnectorPreferredPopLocationInput{
		PreferredOnly: tfLocation.PreferredOnly.ValueBool(),
		Automatic:     tfLocation.Automatic.ValueBool(),
		Primary:       prepareIDRef[cato_models.PopLocationRefInput](ctx, tfLocation.Primary, diags, "preferred_pop_location.primary"),
		Secondary:     prepareIDRef[cato_models.PopLocationRefInput](ctx, tfLocation.Secondary, diags, "preferred_pop_location.secondary"),
	}

	return &sdkLocation
}

func (r *appConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppConnectorModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.AddZtnaAppConnectorInput{
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueStringPointer(),
		GroupName:            plan.GroupName.ValueString(),
		Type:                 cato_models.ZtnaAppConnectorType(plan.Type.ValueString()),
		Address:              r.prepareAddress(ctx, plan.PostalAddress, &diags),
		Timezone:             plan.Timezone.ValueString(),
		PreferredPopLocation: r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
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
	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating appConnector state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, id, plan)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
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

	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		// Check if app-connector not found
		if errors.Is(hydrateErr, ErrAppConnectorNotFound) {
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

func (r *appConnectorResource) parsePostalAddress(ctx context.Context, addr cato_go_sdk.AppConnectorReadConnector_ZtnaAppConnector_ZtnaAppConnector_Address,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare PostalAddressModel object
	tfAddress := PostalAddressModel{
		AddressValidated: types.StringValue(string(addr.AddressValidated)),
		City:             types.StringPointerValue(addr.CityName),
		Country:          parseIDRef(ctx, addr.Country, diags),
		State:            types.StringPointerValue(addr.StateName),
		Street:           types.StringPointerValue(addr.Street),
		ZipCode:          types.StringPointerValue(addr.ZipCode),
	}
	addrObj, diag := types.ObjectValueFrom(ctx, PostalAddressModelTypes, tfAddress)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyScheduleTypes)
	}

	return addrObj
}

func (r *appConnectorResource) parsePopLocation(ctx context.Context, loc *cato_go_sdk.AppConnectorReadConnector_ZtnaAppConnector_ZtnaAppConnector_PreferredPopLocation,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	if loc == nil {
		return types.ObjectNull(PreferredPopLocationModelTypes)
	}

	// Prepare PreferredPopLocationModel object
	tfLocation := PreferredPopLocationModel{
		PreferredOnly: types.BoolValue(loc.PreferredOnly),
		Automatic:     types.BoolValue(loc.Automatic),
		Primary:       types.ObjectNull(IdNameRefModelTypes),
		Secondary:     types.ObjectNull(IdNameRefModelTypes),
	}
	if loc.Primary != nil {
		tfLocation.Primary = parseIDRef(ctx, *loc.Primary, diags)
	}
	if loc.Secondary != nil {
		tfLocation.Secondary = parseIDRef(ctx, *loc.Secondary, diags)
	}

	locObj, diag := types.ObjectValueFrom(ctx, PreferredPopLocationModelTypes, tfLocation)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PreferredPopLocationModelTypes)
	}

	return locObj
}

// hydrateAppConnectorState fetches the current state of a appConnector from the API
// It takes a plan parameter to match config members with API members correctly
func (r *appConnectorResource) hydrateAppConnectorState(ctx context.Context, appConnectorID string, plan AppConnectorModel) (*AppConnectorModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

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
		return nil, diags, err
	}

	// Map API response to AppConnectorModel
	con := result.GetZtnaAppConnector().GetZtnaAppConnector()
	if con == nil {
		return nil, diags, ErrAppConnectorNotFound
	}

	state := &AppConnectorModel{
		ID:   types.StringValue(con.ID),
		Name: types.StringValue(con.Name),
		// Description:  types.StringPointerValue(c.Description), TODO: add description
		SerialNumber: types.StringPointerValue(con.SerialNumber),
		Type:         types.StringValue(con.Type.String()),
		// SocketModel:  types.StringPointerValue(c.SocketModel), // TODO: add socket model
		GroupName:            types.StringValue(con.GroupName),
		PostalAddress:        r.parsePostalAddress(ctx, con.Address, &diags),
		Timezone:             types.StringValue(con.Timezone),
		PreferredPopLocation: r.parsePopLocation(ctx, con.PreferredPopLocation, &diags),
	}

	if diags.HasError() {
		return nil, diags, ErrAPIResponseParse
	}

	return state, nil, nil
}

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
