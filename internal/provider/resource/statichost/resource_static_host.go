package statichost

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &staticHostResource{}
	_ resource.ResourceWithConfigure   = &staticHostResource{}
	_ resource.ResourceWithImportState = &staticHostResource{}
)

func NewStaticHostResource() resource.Resource {
	return &staticHostResource{}
}

type staticHostResource struct {
	client           *catoclient.Service
	staticHostClient Client

	apiClient ResourceClient
}

func (r *staticHostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_host"
}

func (r *staticHostResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *staticHostResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	siteID, hostID, ok := strings.Cut(req.ID, "/")
	if !ok || siteID == "" || hostID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"expected \"<site-id>/<host-id>\"",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), hostID)...)
}

func (r *staticHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticHost
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := convertCreateInput(plan)

	tflog.Debug(ctx, "Create.SiteAddStaticHost.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.resourceAPI().SiteAddStaticHost(ctx, plan.SiteID.ValueString(), input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with static host id
	resp.State.SetAttribute(
		ctx,
		path.Empty().AtName("id"),
		body.Site.GetAddStaticHost().HostID,
	)
}

func (r *staticHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticHost
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	staticHost, err := r.resourceAPI().SiteStaticHost(
		ctx,
		r.client.AccountId,
		staticHostSiteRefFromTerraform(state),
		state.ID.ValueString(),
	)
	tflog.Debug(ctx, "Read.SiteStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(staticHost),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	if staticHost == nil || staticHost.GetSite().GetStaticHost() == nil {
		tflog.Warn(ctx, "static host not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	convertReadState(&state, staticHost)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *staticHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticHost
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := convertUpdateInput(plan)

	tflog.Debug(ctx, "static_host update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	tflog.Debug(ctx, "Update.SiteUpdateStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	siteUpdateStaticHostResponse, err := r.resourceAPI().SiteUpdateStaticHost(ctx, plan.ID.ValueString(), input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateStaticHostResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *staticHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticHost
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.resourceAPI().EntityLookup(
		ctx,
		r.client.AccountId,
		cato_models.EntityType("site"),
		nil,
		nil,
		nil,
		nil,
		[]string{state.SiteID.ValueString()},
		nil,
		nil,
		nil,
	)
	tflog.Debug(ctx, "Delete.EntityLookup.site.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		queryHostResult, err := r.resourceAPI().EntityLookup(
			ctx,
			r.client.AccountId,
			cato_models.EntityType("host"),
			nil,
			nil,
			nil,
			nil,
			[]string{state.ID.ValueString()},
			nil,
			nil,
			nil,
		)
		tflog.Debug(ctx, "Delete.EntityLookup.host.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(queryHostResult),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}

		// check if host exist before removing
		if len(queryHostResult.EntityLookup.GetItems()) == 1 {
			siteRemoveStaticHostResponse, err := r.resourceAPI().SiteRemoveStaticHost(ctx, state.ID.ValueString(), r.client.AccountId)
			tflog.Debug(ctx, "Delete.SiteRemoveStaticHost.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(siteRemoveStaticHostResponse),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteRemoveStaticHost error",
					err.Error(),
				)
				return
			}
		}
	}
}
