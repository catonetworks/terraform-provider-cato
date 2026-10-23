package privateapp

import (
	"context"
	"errors"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &privateAppResource{}
	_ resource.ResourceWithConfigure   = &privateAppResource{}
	_ resource.ResourceWithImportState = &privateAppResource{}
)

var ErrPrivateAppNotFound = errors.New("private-app not found")

func NewPrivateAppResource() resource.Resource {
	return &privateAppResource{}
}

type privateAppResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *privateAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_app"
}

func (r *privateAppResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *privateAppResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a new private app
func (r *privateAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := privateApplicationInputFromTerraform(ctx, r, &resp.Diagnostics, plan)

	// Call Cato API to create a new private app
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PrivateAppCreatePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppCreatePrivateApp error", err.Error())
		return
	}

	// Set the ID from the response
	hydratedState := convertCreateInput(
		ctx,
		r,
		&plan,
		&diags,
		result,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read the private app
func (r *privateAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags, hydrateErr := r.hydratePrivateAppState(ctx, state.ID.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		// Check if private-app was found
		if errors.Is(hydrateErr, ErrPrivateAppNotFound) {
			tflog.Warn(ctx, "private app not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating group state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update the private app
func (r *privateAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Model
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := strings.Trim(state.ID.String(), `"`)
	if id == "" {
		resp.Diagnostics.AddError("PrivateAppUpdatePrivateApp: ID is unknown", "PrivateApp ID is not set in TF state")
		return
	}

	input := cato_models.UpdatePrivateApplicationInput{
		AllowICMPProtocol:  convert.KnownBoolPointer(plan.AllowIcmpProtocol),
		Description:        convert.KnownStringPointer(plan.Description),
		ID:                 id,
		InternalAppAddress: convert.KnownStringPointer(plan.InternalAppAddress),
		Name:               convert.KnownStringPointer(plan.Name),
		PrivateAppProbing:  r.preparePrivateAppProbing(ctx, plan.PrivateAppProbing, &resp.Diagnostics),
		ProbingEnabled:     convert.KnownBoolPointer(plan.ProbingEnabled),
		ProtocolPorts:      r.prepareProtocolPorts(ctx, plan.ProtocolPorts, &resp.Diagnostics),
		Published:          convert.KnownBoolPointer(plan.Published),
		PublishedAppDomain: r.preparePublishedAppDomain(ctx, plan.PublishedAppDomain, &resp.Diagnostics),
	}

	tflog.Debug(ctx, "PrivateAppUpdatePrivateApp", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PrivateAppUpdatePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppUpdatePrivateApp", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})

	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppUpdatePrivateApp error", err.Error())
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivateAppState(ctx, id)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating private-app state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete the private app
func (r *privateAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := convertDeleteInput(state)

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PrivateAppDeletePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})

	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppDeletePrivateApp error", err.Error())
		return
	}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	r *privateAppResource,
	plan *Model,
	priorDiags *diag.Diagnostics,
	result *cato_go_sdk.PrivateAppCreatePrivateApp,
	diagnostics *diag.Diagnostics,
) (resultHydratedState *Model) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(result.GetPrivateApplication().GetCreatePrivateApplication().GetApplication().GetID())

	// Hydrate state from API
	resultHydratedState, diags, hydrateErr := r.hydratePrivateAppState(ctx, plan.ID.ValueString())
	if hydrateErr != nil {
		diagnostics.Append(diags...)
		diagnostics.AddError("Error hydrating privateApp state", hydrateErr.Error())
		return resultHydratedState
	}
	return resultHydratedState
}

// hydratePrivateAppState fetches the current state of a privateApp from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privateAppResource) hydratePrivateAppState(
	ctx context.Context,
	privateAppID string,
) (*Model, diag.Diagnostics, error) {
	var diags diag.Diagnostics
	input := convertHydratePrivateAppStateValues(privateAppID)

	// Call Cato API to get a private-app
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]any{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PrivateAppReadPrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]any{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, diags, err
	}

	// Map API response to PrivateAppModel
	app := result.GetPrivateApplication().GetPrivateApplication()
	if app == nil {
		return nil, diags, ErrPrivateAppNotFound
	}

	state := &Model{
		AllowIcmpProtocol:  types.BoolValue(app.AllowICMPProtocol),
		CreationTime:       types.StringValue(app.CreationTime),
		Description:        types.StringPointerValue(app.Description),
		ID:                 types.StringValue(app.ID),
		InternalAppAddress: types.StringValue(app.InternalAppAddress),
		Name:               types.StringValue(app.Name),
		PrivateAppProbing:  r.parsePrivateAppProbing(ctx, app.PrivateAppProbing, &diags),
		ProbingEnabled:     types.BoolValue(app.ProbingEnabled),
		ProtocolPorts:      r.parseProtocolPorts(ctx, app.ProtocolPorts, &diags),
		Published:          types.BoolValue(app.Published),
		PublishedAppDomain: r.parsePublishedAppDomain(ctx, app.PublishedAppDomain, &diags),
	}
	return state, diags, nil
}
