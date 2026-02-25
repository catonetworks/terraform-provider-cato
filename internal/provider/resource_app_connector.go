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
		Description: "The `cato_app_connector` resource contains the configuration parameters necessary to manage an app connector.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Description: "Optional description of the ZTNA App Connector (max 250 characters)",
				Optional:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "Name of the ZTNA App Connector group",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique ID of the ZTNA App Connector",
				Computed:    true,
			},
			"location": r.schemaLocation(),
			"name": schema.StringAttribute{
				Description: "The unique name of the ZTNA App Connector",
				Required:    true,
			},
			"preferred_pop_location": r.schemaPreferredPopLocation(),
			"private_apps": schema.ListNestedAttribute{
				Description:  "List of private applications",
				Computed:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Private app")},
			},
			"serial_number": schema.StringAttribute{
				Description: "Unique serial number of the ZTNA App Connector",
				Computed:    true,
			},
			"socket_id": schema.StringAttribute{
				Description: "Connector type (virtual, physical)",
				Computed:    true,
			},
			"socket_model": schema.StringAttribute{
				Description: "Socket model of the ZTNA App Connector",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Connector type (virtual, physical)",
				Required:    true,
				Validators:  []validator.String{appConTypeValidator{}},
			},
		},
	}
}

func (r *appConnectorResource) schemaLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "App connector location",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "Street, number",
				Optional:    true,
			},
			"city_name": schema.StringAttribute{
				Description: "City name",
				Required:    true,
			},
			"country_code": schema.StringAttribute{
				Description: "Country code",
				Required:    true,
			},
			"state_code": schema.StringAttribute{
				Description: "State code",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone",
				Required:    true,
			},
		},
	}
}

func (r *appConnectorResource) schemaPreferredPopLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Preferred PoP locations settings",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"automatic": schema.BoolAttribute{
				Description: "Automatic PoP location",
				Required:    true,
			},
			"preferred_only": schema.BoolAttribute{
				Description: "Restrict connector attachment exclusively to the configured pop locations",
				Required:    true,
			},
			"primary": schema.SingleNestedAttribute{
				Description: "Physical location of the pripary Pop",
				Optional:    true,
				Attributes:  schemaNameID("Primary location"),
			},
			"secondary": schema.SingleNestedAttribute{
				Description: "Physical location of the secondary Pop",
				Optional:    true,
				Attributes:  schemaNameID("Secondary location"),
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

// Create a new app connector
func (r *appConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppConnectorModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.AddZtnaAppConnectorInput{
		Description:          knownStringPointer(plan.Description),
		GroupName:            plan.GroupName.ValueString(),
		Location:             r.prepareLocation(ctx, plan.Location, &diags),
		Name:                 plan.Name.ValueString(),
		PreferredPopLocation: r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
		Type:                 cato_models.ZtnaAppConnectorType(plan.Type.ValueString()),
	}

	// Call Cato API to create a new connector
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.AppConnectorCreateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
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

// Read app connector data from Cato API
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

// Update app connector configuration
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
		Description:          knownStringPointer(plan.Description),
		GroupName:            knownStringPointer(plan.GroupName),
		ID:                   id,
		Location:             r.prepareLocation(ctx, plan.Location, &diags),
		Name:                 knownStringPointer(plan.Name),
		PreferredPopLocation: r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
	}

	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.AppConnectorUpdateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})

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

// Delete app connector
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

func (r *appConnectorResource) prepareLocation(ctx context.Context, addr types.Object, diags *diag.Diagnostics) *cato_models.ZtnaAppConnectorLocationInput {
	if !hasValue(addr) {
		return nil
	}

	var tfLocation AppConnectorLocation
	diags.Append(addr.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkLocation := cato_models.ZtnaAppConnectorLocationInput{
		Address:     knownStringPointer(tfLocation.Address),
		City:        tfLocation.CityName.ValueString(),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   knownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}

	return &sdkLocation
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

func (r *appConnectorResource) parseLocation(ctx context.Context, addr cato_go_sdk.AppConnectorReadConnector_ZtnaAppConnector_ZtnaAppConnector_Location,
	diags *diag.Diagnostics,
) types.Object {
	// Prepare AppConnectorLocation object
	tfLocation := AppConnectorLocation{
		Address:     types.StringPointerValue(addr.Address),
		CityName:    types.StringValue(addr.CityName),
		CountryCode: types.StringValue(addr.CountryCode),
		StateCode:   types.StringPointerValue(addr.StateCode),
		Timezone:    types.StringValue(addr.Timezone),
	}

	locObj, diag := types.ObjectValueFrom(ctx, AppConnectorLocationTypes, tfLocation)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(AppConnectorLocationTypes)
	}

	return locObj
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
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.AppConnectorReadConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, diags, err
	}

	// Map API response to AppConnectorModel
	con := result.GetZtnaAppConnector().GetZtnaAppConnector()
	if con == nil {
		return nil, diags, ErrAppConnectorNotFound
	}

	state := &AppConnectorModel{
		Description:          types.StringPointerValue(con.Description),
		GroupName:            types.StringValue(con.GroupName),
		ID:                   types.StringValue(con.ID),
		Location:             r.parseLocation(ctx, con.Location, &diags),
		Name:                 types.StringValue(con.Name),
		PreferredPopLocation: r.parsePopLocation(ctx, con.PreferredPopLocation, &diags),
		PrivateAppRef:        parseIDRefList(ctx, con.PrivateAppRef, &diags),
		SerialNumber:         types.StringPointerValue(con.SerialNumber),
		SocketID:             types.StringPointerValue(con.SocketID),
		SocketModel:          types.StringPointerValue((*string)(con.SocketModel)),
		Type:                 types.StringValue(con.Type.String()),
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
