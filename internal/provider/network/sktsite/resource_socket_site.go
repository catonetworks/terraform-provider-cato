package sktsite

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	errorsapp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/client"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/adapter"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

var (
	_ resource.Resource                = &socketSiteResource{}
	_ resource.ResourceWithConfigure   = &socketSiteResource{}
	_ resource.ResourceWithImportState = &socketSiteResource{}
)

func NewSocketSiteResource() resource.Resource {
	return &socketSiteResource{}
}

type SocketSiteClient = adapter.SocketSiteClient
type socketSiteResource struct {
	client           *client.CatoClientData
	socketSiteClient SocketSiteClient
	retry            adapter.Retry
}

func (r *socketSiteResource) getSocketSiteClient() SocketSiteClient {
	return adapter.ResolveClient(r.socketSiteClient, r.client)
}
func (r *socketSiteResource) sdk() adapter.SDK {
	return adapter.SDK{
		Client:    r.getSocketSiteClient(),
		AccountID: r.client.AccountId,
	}
}
func (r *socketSiteResource) service(ctx context.Context, cfg *SocketSite, state SocketSite) application.Service {
	return application.Service{
		Port:  r.sdk(),
		Retry: r.retry,
		ValidateSnapshot: func(snapshot *application.Snapshot) error {
			var diags diag.Diagnostics
			r.projectState(ctx, cfg, state, snapshot, &diags)
			if diags.HasError() {
				return &projectionError{
					Diagnostics: diags,
				}
			}
			return nil
		},
	}
}

type projectionError struct{ Diagnostics diag.Diagnostics }

func (e *projectionError) Error() string { return fmt.Sprint(e.Diagnostics) }
func (r *socketSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.resourceSchema()
}
func (r *socketSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site"
}

func (r *socketSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.CatoClientData)
}

func (r *socketSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// TODO: check if this is really necessary
	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	readResp := resource.ReadResponse{
		State:       resp.State,
		Diagnostics: resp.Diagnostics,
	}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists := r.hydrateSocketSiteState(ctx, nil, state, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *socketSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var cfg, plan SocketSite
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)

	if resp.Diagnostics.HasError() {
		return
	}
	input := r.prepareInput(ctx, &cfg, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.service(ctx, &cfg, plan).Create(ctx, input)
	appendSocketSiteError(&resp.Diagnostics, err)
	if resp.Diagnostics.HasError() {
		return
	}
	if result.Pending {
		cfg.ID = types.StringValue(result.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}
	if !result.Found {
		resp.State.RemoveResource(ctx)
		return
	}
	state := r.projectState(ctx, &cfg, plan, result.Snapshot, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *socketSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan SocketSite
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	var prior SocketSite
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)

	if resp.Diagnostics.HasError() {
		return
	}
	input := r.prepareInput(ctx, &cfg, &plan, &prior, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.service(ctx, &cfg, plan).Update(ctx, input)
	appendSocketSiteError(&resp.Diagnostics, err)
	if resp.Diagnostics.HasError() {
		return
	}
	if result.Pending {
		cfg.ID = types.StringValue(result.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}
	if !result.Found {
		resp.State.RemoveResource(ctx)
		return
	}
	state := r.projectState(ctx, &cfg, plan, result.Snapshot, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *socketSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketSite
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	appendSocketSiteError(
		&resp.Diagnostics,
		r.service(ctx, nil, state).Delete(ctx, state.ID.ValueString()),
	)
}
func (r *socketSiteResource) hydrateSocketSiteState(
	ctx context.Context,
	cfg *SocketSite,
	state SocketSite,
	id string,
	diags *diag.Diagnostics,
) (
	SocketSite,
	bool,
) {
	input := r.readInput(ctx, cfg, state, id, diags)
	if diags.HasError() {
		return state, false
	}
	result, err := r.service(ctx, cfg, state).Read(ctx, input)
	appendSocketSiteError(diags, err)
	if diags.HasError() || !result.Found {
		return state, false
	}
	return r.projectState(ctx, cfg, state, result.Snapshot, diags), true
}
func (r *socketSiteResource) fetchSocketConfiguration(ctx context.Context, id string, diags *diag.Diagnostics) *application.Configuration {
	value, err := r.sdk().Configuration(ctx, id)
	appendSocketSiteError(diags, err)
	return value
}
func appendSocketSiteError(diags *diag.Diagnostics, err error) {
	if err == nil {
		return
	}
	var projection *projectionError
	if errors.As(err, &projection) {
		diags.Append(projection.Diagnostics...)
		return
	}
	var operation *errorsapp.Error
	if errors.As(err, &operation) {
		diags.AddError(operation.Summary, err.Error())
		return
	}
	diags.AddError("Cato API error", err.Error())
}
