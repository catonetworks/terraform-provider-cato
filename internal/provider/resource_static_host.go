package provider

import (
	"context"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
	client           *catoClientData
	staticHostClient StaticHostClient
}

type StaticHostClient interface {
	SiteStaticHost(ctx context.Context, accountID string, siteRefInput cato_models.SiteRefInput, hostID string,
		interceptors ...clientv2.RequestInterceptor) (*cato.SiteStaticHost, error)
}

func (r *staticHostResource) getStaticHostClient() StaticHostClient {
	if r.staticHostClient != nil {
		return r.staticHostClient
	}
	if r.client == nil {
		return nil
	}

	return r.client.catov2
}

func (r *staticHostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_host"
}

func (r *staticHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_static_host` resource contains the configuration parameters necessary to add a static host. " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.addStaticHost()](https://api.catonetworks.com/documentation/#mutation-site.addStaticHost).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID (Host's parent)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Host name",
				Required:    true,
			},
			"ip": schema.StringAttribute{
				Description: "Host IP address",
				Required:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "Host MAC address (for DHCP reservervation)",
				Optional:    true,
			},
		},
	}
}

func (r *staticHostResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *staticHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	input := cato_models.AddStaticHostInput{
		Name:       plan.Name.ValueString(),
		IP:         plan.IP.ValueString(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}

	tflog.Debug(ctx, "Create.SiteAddStaticHost.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.client.catov2.SiteAddStaticHost(ctx, plan.SiteID.ValueString(), input, r.client.AccountId)
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

	staticHost, err := r.getStaticHostClient().SiteStaticHost(
		ctx,
		r.client.AccountId,
		cato_models.SiteRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.SiteID.ValueString(),
		},
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

	host := staticHost.GetSite().GetStaticHost().GetHost()
	state.ID = types.StringValue(host.HostID)
	state.Name = types.StringValue(host.Name)
	state.IP = types.StringValue(host.IP)
	if host.MacAddress == nil {
		state.MacAddress = types.StringNull()
	} else {
		state.MacAddress = types.StringValue(*host.MacAddress)
	}

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
	input := cato_models.UpdateStaticHostInput{
		Name:       plan.Name.ValueStringPointer(),
		IP:         plan.IP.ValueStringPointer(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}

	tflog.Debug(ctx, "static_host update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	tflog.Debug(ctx, "Update.SiteUpdateStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	siteUpdateStaticHostResponse, err := r.client.catov2.SiteUpdateStaticHost(ctx, plan.ID.ValueString(), input, r.client.AccountId)
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

	querySiteResult, err := r.client.catov2.EntityLookup(
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
		queryHostResult, err := r.client.catov2.EntityLookup(
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
			siteRemoveStaticHostResponse, err := r.client.catov2.SiteRemoveStaticHost(ctx, state.ID.ValueString(), r.client.AccountId)
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
