package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	client *catoClientData
}

func (r *staticHostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_host"
}

func (r *staticHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_static_host` resource contains the configuration parameters necessary to add a static host. Documentation for the underlying API used in this resource can be found at [mutation.addStaticHost()](https://api.catonetworks.com/documentation/#mutation-site.addStaticHost).",
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

func (r *staticHostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *staticHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
		IP:         plan.Ip.ValueString(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}

	tflog.Debug(ctx, "Create.SiteAddStaticHost.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.client.catov2.SiteAddStaticHost(ctx, plan.SiteId.ValueString(), input, r.client.AccountId)
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

	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteId.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.site.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
		tflog.Warn(ctx, "site not found, static host resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// check if host exist before removing
	queryHostResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("host"), nil, nil, nil, nil, []string{state.Id.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Read.EntityLookup.host.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryHostResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// read in the ipsec site entries
	for _, v := range queryHostResult.EntityLookup.Items {
		if v.Entity.ID == state.Id.ValueString() {
			resp.State.SetAttribute(
				ctx,
				path.Root("id"),
				v.Entity.ID,
			)
		}
	}

	if len(queryHostResult.EntityLookup.GetItems()) != 1 {
		tflog.Warn(ctx, "static host found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
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
		IP:         plan.Ip.ValueStringPointer(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}

	tflog.Debug(ctx, "static_host update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	tflog.Debug(ctx, "Update.SiteUpdateStaticHost.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	siteUpdateStaticHostResponse, err := r.client.catov2.SiteUpdateStaticHost(ctx, plan.Id.ValueString(), input, r.client.AccountId)
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

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.SiteId.ValueString()}, nil, nil, nil)
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
		queryHostResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("host"), nil, nil, nil, nil, []string{state.Id.ValueString()}, nil, nil, nil)
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
			siteRemoveStaticHostResponse, err := r.client.catov2.SiteRemoveStaticHost(ctx, state.Id.ValueString(), r.client.AccountId)
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
