package privateapp

import (
	"context"
	"errors"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

func (r *privateAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	input := cato_models.CreatePrivateApplicationInput{
		AllowICMPProtocol:  plan.AllowIcmpProtocol.ValueBool(),
		Description:        convert.KnownStringPointer(plan.Description),
		InternalAppAddress: plan.InternalAppAddress.ValueString(),
		Name:               plan.Name.ValueString(),
		PrivateAppProbing:  r.preparePrivateAppProbing(ctx, plan.PrivateAppProbing, &resp.Diagnostics),
		ProbingEnabled:     plan.ProbingEnabled.ValueBool(),
		ProtocolPorts:      r.prepareProtocolPorts(ctx, plan.ProtocolPorts, &resp.Diagnostics),
		Published:          plan.Published.ValueBool(),
		PublishedAppDomain: r.preparePublishedAppDomain(ctx, plan.PublishedAppDomain, &resp.Diagnostics),
	}

	// Call Cato API to create a new private app
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.PrivateAppCreatePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppCreatePrivateApp error", err.Error())
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(result.GetPrivateApplication().GetCreatePrivateApplication().GetApplication().GetID())

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivateAppState(ctx, plan.ID.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating privateApp state", hydrateErr.Error())
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
	result, err := r.client.Catov2.PrivateAppUpdatePrivateApp(ctx, r.client.AccountId, input)
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
	input := cato_models.DeletePrivateApplicationInput{
		PrivateApplication: &cato_models.PrivateApplicationRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.PrivateAppDeletePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})

	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppDeletePrivateApp error", err.Error())
		return
	}
}

// hydratePrivateAppState fetches the current state of a privateApp from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privateAppResource) hydratePrivateAppState(ctx context.Context, privateAppID string) (*Model, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	input := cato_models.PrivateApplicationRefInput{
		By:    cato_models.ObjectRefByID,
		Input: privateAppID,
	}

	// Call Cato API to get a private-app
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]any{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.PrivateAppReadPrivateApp(ctx, r.client.AccountId, input)
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

func (r *privateAppResource) parseProtocolPorts(
	ctx context.Context,
	protoPorts []*privateAppProtocolPorts,
	diags *diag.Diagnostics,
) types.Set {
	var objDiags diag.Diagnostics
	setNull := types.SetNull(types.ObjectType{AttrTypes: ProtocolPortTypes})

	if protoPorts == nil {
		return setNull
	}

	protoPortsObjects := make([]types.Object, 0, len(protoPorts))

	for _, pp := range protoPorts {
		if pp == nil {
			continue
		}

		// Ports
		tfPortSet := types.SetNull(types.Int64Type)
		if pp.Port != nil {
			intSlice := make([]types.Int64, 0, len(pp.Port))
			for _, p := range pp.Port {
				intSlice = append(intSlice, types.Int64Value(p.GetInt64()))
			}
			tfPortSet, objDiags = types.SetValueFrom(ctx, types.Int64Type, intSlice)
			diags.Append(objDiags...)
			if diags.HasError() {
				return setNull
			}
		}

		// Port range
		tfPortRangeObj := types.ObjectNull(PortRangeTypes)
		if pp.PortRange != nil {
			tfPortRange := PortRange{
				From: types.Int64Value(pp.PortRange.From.GetInt64()),
				To:   types.Int64Value(pp.PortRange.To.GetInt64()),
			}
			tfPortRangeObj, objDiags = types.ObjectValueFrom(ctx, PortRangeTypes, tfPortRange)
			diags.Append(objDiags...)
			if diags.HasError() {
				return setNull
			}
		}

		// ProtocolPorts item
		tfPPort := ProtocolPort{
			Ports:     tfPortSet,
			PortRange: tfPortRangeObj,
			Protocol:  types.StringValue(string(pp.Protocol)),
		}
		tfPPortObj, objDiags := types.ObjectValueFrom(ctx, ProtocolPortTypes, tfPPort)
		diags.Append(objDiags...)
		if diags.HasError() {
			return setNull
		}
		protoPortsObjects = append(protoPortsObjects, tfPPortObj)
	}

	// convert slice to types.Set
	tfProtoPortSet, objDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ProtocolPortTypes}, protoPortsObjects)
	diags.Append(objDiags...)
	if diags.HasError() {
		return setNull
	}

	return tfProtoPortSet
}

func (r *privateAppResource) parsePublishedAppDomain(
	ctx context.Context,
	domain *privateAppPublishedAppDomain,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	objNull := types.ObjectNull(PublishedAppDomainTypes)

	if domain == nil {
		return objNull
	}

	tfDomain := PublishedAppDomain{
		ConnectorGroupName: types.StringPointerValue(domain.ConnectorGroupName),
		CreationTime:       types.StringValue(domain.CreationTime),
		ID:                 types.StringValue(domain.ID),
		PublishedAppDomain: types.StringValue(domain.PublishedAppDomain),
	}

	domainObj, objDiags := types.ObjectValueFrom(ctx, PublishedAppDomainTypes, tfDomain)
	diags.Append(objDiags...)
	if diags.HasError() {
		return objNull
	}
	return domainObj
}

func (r *privateAppResource) parsePrivateAppProbing(
	ctx context.Context,
	probing *privateAppProbing,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	objNull := types.ObjectNull(ProbingTypes)

	if probing == nil {
		return objNull
	}

	tfProbing := Probing{
		FaultThresholdDown: types.Int64Value(probing.FaultThresholdDown),
		ID:                 types.StringValue(probing.ID),
		Interval:           types.Int64Value(probing.Interval),
		Type:               types.StringValue(probing.Type),
	}

	probingObj, objDiags := types.ObjectValueFrom(ctx, ProbingTypes, tfProbing)
	diags.Append(objDiags...)
	if diags.HasError() {
		return objNull
	}
	return probingObj
}

func (r *privateAppResource) preparePublishedAppDomain(ctx context.Context, appDomain types.Object, diags *diag.Diagnostics,
) *cato_models.PublishedAppDomainInput {
	if !utils.HasValue(appDomain) {
		return nil
	}

	var tfAppDomain PublishedAppDomain
	if utils.CheckErr(diags, appDomain.As(ctx, &tfAppDomain, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PublishedAppDomainInput{
		ConnectorGroupName: convert.KnownStringPointer(tfAppDomain.ConnectorGroupName),
		CreationTime:       convert.KnownStringPointer(tfAppDomain.CreationTime),
		ID:                 convert.KnownStringPointer(tfAppDomain.ID),
		PublishedAppDomain: convert.KnownStringPointer(tfAppDomain.PublishedAppDomain),
	}
}

func (r *privateAppResource) preparePrivateAppProbing(ctx context.Context, probing types.Object, diags *diag.Diagnostics,
) *cato_models.PrivateAppProbingInput {
	if !utils.HasValue(probing) {
		return nil
	}

	var tfProbing Probing
	if utils.CheckErr(diags, probing.As(ctx, &tfProbing, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PrivateAppProbingInput{
		FaultThresholdDown: convert.KnownInt64Pointer(tfProbing.FaultThresholdDown),
		ID:                 convert.KnownStringPointer(tfProbing.ID),
		Interval:           convert.KnownInt64Pointer(tfProbing.Interval),
		Type:               convert.KnownStringPointer(tfProbing.Type),
	}
}

func (r *privateAppResource) prepareProtocolPorts(
	ctx context.Context,
	protPorts types.Set,
	diags *diag.Diagnostics,
) []*cato_models.CustomServiceInput {
	if !utils.HasValue(protPorts) {
		return nil
	}

	var out []*cato_models.CustomServiceInput

	for _, p := range protPorts.Elements() {
		var svcInput cato_models.CustomServiceInput

		if !utils.HasValue(p) {
			continue
		}
		port := p.(types.Object)

		var tfProtoPort ProtocolPort
		if utils.CheckErr(diags, port.As(ctx, &tfProtoPort, basetypes.ObjectAsOptions{})) {
			return nil
		}

		// Port numbers
		if utils.HasValue(tfProtoPort.Ports) {
			var tfPortNumbers []types.Int64
			if utils.CheckErr(diags, tfProtoPort.Ports.ElementsAs(ctx, &tfPortNumbers, false)) {
				return nil
			}
			for _, portNum := range tfPortNumbers {
				if utils.HasValue(portNum) {
					svcInput.Port = append(svcInput.Port, scalars.Port(portNum.String()))
				}
			}
		}

		// Port range
		if utils.HasValue(tfProtoPort.PortRange) {
			var tfProtoRange PortRange
			if utils.CheckErr(diags, tfProtoPort.PortRange.As(ctx, &tfProtoRange, basetypes.ObjectAsOptions{})) {
				return nil
			}
			svcInput.PortRange = &cato_models.PortRangeInput{
				From: scalars.Port(tfProtoRange.From.String()),
				To:   scalars.Port(tfProtoRange.To.String()),
			}
		}

		// Protocol
		svcInput.Protocol = cato_models.IPProtocol(tfProtoPort.Protocol.ValueString())

		out = append(out, &svcInput)
	}

	return out
}
