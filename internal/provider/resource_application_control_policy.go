package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &applicationControlPolicyResource{}
	_ resource.ResourceWithConfigure   = &applicationControlPolicyResource{}
	_ resource.ResourceWithImportState = &applicationControlPolicyResource{}
)

func NewApplicationControlPolicyResource() resource.Resource {
	return &applicationControlPolicyResource{}
}

type applicationControlPolicyResource struct {
	client *catoClientData
}

func (r *applicationControlPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_control_policy"
}

func (r *applicationControlPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Account-level Application Control (App & Data Inline Protection) policy toggles. " +
			"Underlying GraphQL is marked @beta.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed identifier for import",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the Application Control policy is enabled",
				Required:    true,
			},
			"data_control_enabled": schema.StringAttribute{
				Description: "Data control feature toggle (ENABLED or DISABLED)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(cato_models.PolicyToggleStateEnabled), string(cato_models.PolicyToggleStateDisabled)),
				},
			},
			"published_time": schema.StringAttribute{
				Description: "Last published time from audit",
				Computed:    true,
			},
			"published_by": schema.StringAttribute{
				Description: "Last published by from audit",
				Computed:    true,
			},
			"revision_id": schema.StringAttribute{
				Description: "Active draft revision ID when present",
				Computed:    true,
			},
			"revision_name": schema.StringAttribute{
				Description: "Active draft revision name when present",
				Computed:    true,
			},
		},
	}
}

func (r *applicationControlPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *applicationControlPolicyResource) ImportState(
	ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "application_control")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data_control_enabled"), string(cato_models.PolicyToggleStateDisabled))...)
}

func (r *applicationControlPolicyResource) readState(ctx context.Context) (ApplicationControlPolicyModel, error) {
	out := ApplicationControlPolicyModel{ID: types.StringValue("application_control")}
	body, err := r.client.catov2.ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		return out, err
	}
	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		return out, nil
	}
	out.Enabled = types.BoolValue(pol.Enabled)
	if pol.GetAdditionalAttributes() != nil && pol.GetAdditionalAttributes().GetDataControlEnabled() != nil {
		out.DataControlEnabled = types.StringValue(string(*pol.GetAdditionalAttributes().GetDataControlEnabled()))
	} else {
		out.DataControlEnabled = types.StringNull()
	}
	if pol.GetAudit() != nil {
		out.PublishedTime = types.StringValue(pol.GetAudit().GetPublishedTime())
		out.PublishedBy = types.StringValue(pol.GetAudit().GetPublishedBy())
	}
	if pol.GetRevision() != nil {
		out.RevisionID = types.StringValue(pol.GetRevision().GetID())
		out.RevisionName = types.StringValue(pol.GetRevision().GetName())
	} else {
		out.RevisionID = types.StringNull()
		out.RevisionName = types.StringNull()
	}
	return out, nil
}

func (r *applicationControlPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationControlPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationControlPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.updatePolicy(ctx, plan)...)
	st, err := r.readState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("read policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_application_control_policy delete is a no-op; policy always exists")
}

func (r *applicationControlPolicyResource) updatePolicy(ctx context.Context, plan ApplicationControlPolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics
	dc := cato_models.PolicyToggleState(plan.DataControlEnabled.ValueString())
	st := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		st = cato_models.PolicyToggleStateEnabled
	}
	in := cato_models.ApplicationControlPolicyUpdateInput{
		State: &st,
		AdditionalAttributes: &cato_models.ApplicationControlConfigInput{
			DataControlEnabled: dc,
		},
	}
	res, err := r.client.catov2.PolicyApplicationControlUpdatePolicy(ctx, in, r.client.AccountId)
	if err != nil {
		diags.AddError("PolicyApplicationControlUpdatePolicy", err.Error())
		return diags
	}
	up := res.GetPolicy().GetApplicationControl().GetUpdatePolicy()
	if up.GetStatus() == nil || *up.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range up.GetErrors() {
			if e != nil && e.ErrorCode != nil {
				diags.AddError("API error", *e.ErrorMessage)
			}
		}
		return diags
	}
	diags.Append(publishApplicationControlPolicyRevision(ctx, r.client)...)
	return diags
}
