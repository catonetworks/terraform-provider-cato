package provider

import (
	"context"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privateAppResource{}
	_ resource.ResourceWithConfigure   = &privateAppResource{}
	_ resource.ResourceWithImportState = &privateAppResource{}
)

func NewPrivateAppResource() resource.Resource {
	return &privateAppResource{}
}

type privateAppResource struct {
	client *catoClientData
}

func (r *privateAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_app"
}

func (r *privateAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_app` resource contains the configuration parameters necessary to manage a private apps.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the Private App",
				Computed:    true,
			},
			"creation_time": schema.StringAttribute{
				Description: "Creation time",
				Computed:    true,
			},

			"name": schema.StringAttribute{
				Description: "The unique name of the private App",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the private App",
				Optional:    true,
			},
			"internal_app_address": schema.StringAttribute{
				Description: "The local address of the application",
				Required:    true,
			},
			"probing_enabled": schema.BoolAttribute{
				Description: "Is probing is enabled?",
				Required:    true,
			},
			"published": schema.BoolAttribute{
				Description: "Is the private app published?",
				Required:    true,
			},
			"allow_icmp_protocol": schema.BoolAttribute{
				Description: "Is ICMP enabled?",
				Required:    true,
			},
			"published_app_domain": schema.SingleNestedAttribute{
				Description: "Domain information about the published private app",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID of the private app domain",
						Computed:    true,
					},
					"creation_time": schema.StringAttribute{
						Description: "Creation time",
						Computed:    true,
					},
					"published_app_domain": schema.StringAttribute{
						Description: "Published app domain",
						Required:    true,
					},
					"cato_ip": schema.StringAttribute{
						Description: "Cato IP address",
						Optional:    true,
					},
					"connector_group_name": schema.StringAttribute{
						Description: "Connector group name",
						Required:    true,
					},
				},
			},

			"private_app_probing": schema.SingleNestedAttribute{
				Description: "Private app probing settings",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Probing ID",
						Computed:    true,
					},
					"type": schema.StringAttribute{
						Description: "Probing type",
						Required:    true,
					},
					"interval": schema.Int64Attribute{
						Description: "Probing interval",
						Required:    true,
					},
					"fault_threshold_down": schema.Int64Attribute{
						Description: "Fault threshold",
						Required:    true,
					},
				},
			},
			"protocol_ports": schema.ListNestedAttribute{
				Description: "List ports and protocols",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ports": schema.ListAttribute{
							ElementType: types.Int64Type,
							Description: "List of TCP or UDP ports",
							Optional:    true,
						},
						"port_range": schema.SingleNestedAttribute{
							Description: "Port range",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"from": schema.Int64Attribute{
									Description: "From",
									Required:    true,
								},
								"to": schema.Int64Attribute{
									Description: "To",
									Required:    true,
								},
							},
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol; e.g.: TCP, UDP, ICMP",
							Required:    true,
							Validators:  []validator.String{privAppProtocolValidator{}},
						},
					},
				},
			},
		},
	}
}

func (r *privateAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *privateAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privateAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAppModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.CreatePrivateApplicationInput{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueStringPointer(),
		InternalAppAddress: plan.InternalAppAddress.ValueString(),
		ProbingEnabled:     plan.ProbingEnabled.ValueBool(),
		Published:          plan.Published.ValueBool(),
		AllowICMPProtocol:  plan.AllowIcmpProtocol.ValueBool(),
		ProtocolPorts:      r.prepareProtocolPorts(&plan),
		PublishedAppDomain: r.preparePublishedAppDomain(&plan),
		PrivateAppProbing:  r.preparePrivateAppProbing(&plan),
	}

	// Call Cato API to create a new private app
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.PrivateAppCreatePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppCreatePrivateApp", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppCreatePrivateApp error", err.Error())
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(result.GetPrivateApplication().GetCreatePrivateApplication().GetApplication().GetID())

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydratePrivateAppState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateApp state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privateAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PrivateAppModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PrivateAppModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := strings.Trim(state.ID.String(), `"`)
	if id == "" {
		resp.Diagnostics.AddError("PrivateAppUpdatePrivateApp: ID in unknown", "PrivateApp ID is not set in TF state")
		return
	}

	input := cato_models.UpdatePrivateApplicationInput{
		ID:                 id,
		Name:               plan.Name.ValueStringPointer(),
		Description:        plan.Description.ValueStringPointer(),
		InternalAppAddress: plan.InternalAppAddress.ValueStringPointer(),
		ProbingEnabled:     plan.ProbingEnabled.ValueBoolPointer(),
		Published:          plan.Published.ValueBoolPointer(),
		AllowICMPProtocol:  plan.AllowIcmpProtocol.ValueBoolPointer(),
		ProtocolPorts:      r.prepareProtocolPorts(&plan),
		PublishedAppDomain: r.preparePublishedAppDomain(&plan),
		PrivateAppProbing:  r.preparePrivateAppProbing(&plan),
	}

	tflog.Debug(ctx, "PrivateAppUpdatePrivateApp", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.PrivateAppUpdatePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppUpdatePrivateApp", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppUpdatePrivateApp error", err.Error())
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydratePrivateAppState(ctx, id, plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating private-app state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privateAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivateAppModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, hydrateErr := r.hydratePrivateAppState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		// Check if private-app not found
		if hydrateErr.Error() == "private_app not found" { // TODO: check the actual error
			tflog.Warn(ctx, "private_app not found, resource removed")
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

func (r *privateAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PrivateAppModel
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
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.PrivateAppDeletePrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppDeletePrivateApp", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API PrivateAppDeletePrivateApp error", err.Error())
		return
	}
}

// hydratePrivateAppState fetches the current state of a privateApp from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privateAppResource) hydratePrivateAppState(ctx context.Context, privateAppID string, plan PrivateAppModel) (*PrivateAppModel, error) {
	input := cato_models.PrivateApplicationRefInput{
		By:    cato_models.ObjectRefByID,
		Input: privateAppID,
	}

	// Call Cato API to get a private-app
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]any{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.PrivateAppReadPrivateApp(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]any{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}

	// Map API response to PrivateAppModel
	app := result.GetPrivateApplication().GetPrivateApplication()
	state := &PrivateAppModel{
		ID:                 types.StringValue(app.ID),
		CreationTime:       types.StringValue(app.CreationTime),
		Name:               types.StringValue(app.Name),
		Description:        types.StringPointerValue(app.Description),
		InternalAppAddress: types.StringValue(app.InternalAppAddress),
		ProbingEnabled:     types.BoolValue(app.ProbingEnabled),
		Published:          types.BoolValue(app.Published),
		AllowIcmpProtocol:  types.BoolValue(app.AllowICMPProtocol),
	}
	if app.PublishedAppDomain != nil {
		pad := app.PublishedAppDomain
		state.PublishedAppDomain = &PublishedAppDomain{
			ID:                 types.StringValue(pad.ID),
			CreationTime:       types.StringValue(pad.CreationTime),
			CatoIP:             types.StringPointerValue(pad.CatoIP),
			ConnectorGroupName: types.StringPointerValue(pad.ConnectorGroupName),
		}
	}
	if app.PrivateAppProbing != nil {
		pap := app.PrivateAppProbing
		state.PrivateAppProbing = &PrivateAppProbing{
			ID:                 types.StringValue(pap.ID),
			Type:               types.StringValue(pap.Type),
			Interval:           types.Int64Value(pap.Interval),
			FaultThresholdDown: types.Int64Value(pap.FaultThresholdDown),
		}
	}
	if len(app.ProtocolPorts) > 0 {
		for _, pp := range app.ProtocolPorts {
			portDef := ProtocolPort{
				Protocol: types.StringValue(string(pp.Protocol)),
			}
			if pp.PortRange != nil {
				portDef.PortRange = &PortRange{
					From: types.Int64Value(pp.PortRange.From.GetInt64()),
					To:   types.Int64Value(pp.PortRange.To.GetInt64()),
				}
			}
			for _, port := range pp.Port {
				portDef.Ports = append(portDef.Ports, types.Int64Value(port.GetInt64()))
			}
			state.ProtocolPorts = append(state.ProtocolPorts, portDef)
		}
	}
	return state, nil
}

func (r *privateAppResource) preparePublishedAppDomain(plan *PrivateAppModel) *cato_models.PublishedAppDomainInput {
	if plan.PublishedAppDomain == nil {
		return nil
	}
	return &cato_models.PublishedAppDomainInput{
		PublishedAppDomain: plan.PublishedAppDomain.PublishedAppDomain.ValueStringPointer(),
		CatoIP:             plan.PublishedAppDomain.CatoIP.ValueStringPointer(),
		ConnectorGroupName: plan.PublishedAppDomain.ConnectorGroupName.ValueStringPointer(),
	}
}

func (r *privateAppResource) preparePrivateAppProbing(plan *PrivateAppModel) *cato_models.PrivateAppProbingInput {
	if plan.PrivateAppProbing == nil {
		return nil
	}
	return &cato_models.PrivateAppProbingInput{
		ID:                 plan.PrivateAppProbing.ID.ValueStringPointer(),
		Type:               plan.PrivateAppProbing.Type.ValueStringPointer(),
		Interval:           plan.PrivateAppProbing.Interval.ValueInt64Pointer(),
		FaultThresholdDown: plan.PrivateAppProbing.FaultThresholdDown.ValueInt64Pointer(),
	}
}

func (r *privateAppResource) prepareProtocolPorts(plan *PrivateAppModel) []*cato_models.CustomServiceInput {
	if len(plan.ProtocolPorts) == 0 {
		return nil
	}

	protocolPorts := make([]*cato_models.CustomServiceInput, 0, len(plan.ProtocolPorts))

	for _, pp := range plan.ProtocolPorts {
		svcInput := cato_models.CustomServiceInput{
			Protocol: cato_models.IPProtocol(pp.Protocol.ValueString()),
		}
		if pp.PortRange != nil {
			svcInput.PortRange = &cato_models.PortRangeInput{
				From: scalars.Port(pp.PortRange.From.String()),
				To:   scalars.Port(pp.PortRange.To.String()),
			}
		}
		for _, port := range pp.Ports {
			svcInput.Port = append(svcInput.Port, scalars.Port(port.String()))
		}
		protocolPorts = append(protocolPorts, &svcInput)
	}
	return protocolPorts
}

// Validators
type (
	privAppProtocolValidator struct{}
)

func (v privAppProtocolValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := strings.Trim(req.ConfigValue.String(), `"`)
	protoName := cato_models.IPProtocol(value)
	if !protoName.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid protocol (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllIPProtocol))
		return
	}
}
func (v privAppProtocolValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Protocol must be one of: %v", cato_models.AllIPProtocol)
}
func (v privAppProtocolValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
