//nolint:lll
package acsection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &applicationControlSectionResource{}
	_ resource.ResourceWithConfigure   = &applicationControlSectionResource{}
	_ resource.ResourceWithImportState = &applicationControlSectionResource{}
)

func NewApplicationControlSectionResource() resource.Resource {
	return &applicationControlSectionResource{}
}

type applicationControlSectionResource struct {
	client *catoclient.Service
}

func (r *applicationControlSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_control_section"
}

func (r *applicationControlSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *applicationControlSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *applicationControlSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationControlSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PolicyAddSectionInput{}
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		resp.Diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		input.At.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}
	if !plan.Section.IsNull() {
		input.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := idname.Ref{}
		resp.Diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		input.Section.Name = sectionInput.Name.ValueString()
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyApplicationControlAddSection", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	policyChange, err := r.client.Catov2.PolicyApplicationControlAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlAddSection error", err.Error())
		return
	}
	add := policyChange.GetPolicy().GetApplicationControl().GetAddSection()
	for _, e := range add.GetErrors() {
		if e != nil && e.ErrorCode != nil {
			resp.Diagnostics.AddError("API error: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}
	if add.GetStatus() == nil || *add.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		st := ""
		if add.GetStatus() != nil {
			st = string(*add.GetStatus())
		}
		resp.Diagnostics.AddError("Application Control addSection failed", st)
		return
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := add.GetSection().GetSection().GetID()
	plan.ID = types.StringValue(sid)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.State.SetAttribute(ctx, path.Root("section").AtName("id"), sid)
}

func (r *applicationControlSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationControlSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Catov2.ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		tflog.Warn(ctx, "application control policy missing in API response")
		resp.State.RemoveResource(ctx)
		return
	}
	found := false
	for _, item := range pol.GetSections() {
		if item == nil || item.GetSection() == nil {
			continue
		}
		if item.GetSection().GetID() != section.ID.ValueString() {
			continue
		}
		found = true
		state.ID = types.StringValue(item.GetSection().GetID())
		cur, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(item.GetSection().GetID()),
			"name": types.StringValue(item.GetSection().GetName()),
		})
		resp.Diagnostics.Append(d...)
		state.Section = cur
		break
	}

	// Preserve the at block from existing state; the API does not return positioning
	// info on read. Only default to LAST_IN_POLICY when state has nothing (e.g. import).
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		resp.Diagnostics.Append(d...)
		state.At = atObj
	}

	if !found {
		tflog.Warn(ctx, "application control section not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:gocyclo
func (r *applicationControlSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationControlSection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdate := cato_models.PolicyUpdateSectionInput{}
	inputMove := cato_models.PolicyMoveSectionInput{}
	if !plan.Section.IsNull() {
		sectionInput := idname.Ref{}
		resp.Diagnostics.Append(plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})...)
		inputUpdate.ID = sectionInput.ID.ValueString()
		inputUpdate.Section = &cato_models.PolicyUpdateSectionInfoInput{
			Name: sectionInput.Name.ValueStringPointer(),
		}
	}
	if !plan.At.IsNull() {
		positionInput := policyrule.Position{}
		resp.Diagnostics.Append(plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)
		inputMove.To = &cato_models.PolicySectionPositionInput{
			Position: cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		}
		inputMove.ID = inputUpdate.ID
	}
	if resp.Diagnostics.HasError() {
		return
	}

	moveSection, err := r.client.Catov2.PolicyApplicationControlMoveSection(ctx, inputMove, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlMoveSection error", err.Error())
		return
	}
	if moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetStatus() == nil ||
		*moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.GetPolicy().GetApplicationControl().GetMoveSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error moving section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	updateSection, err := r.client.Catov2.PolicyApplicationControlUpdateSection(ctx, inputUpdate, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlUpdateSection error", err.Error())
		return
	}
	if updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetStatus() == nil ||
		*updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.GetPolicy().GetApplicationControl().GetUpdateSection().GetErrors() {
			if item != nil && item.ErrorCode != nil {
				resp.Diagnostics.AddError("API error updating section", fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage))
				return
			}
		}
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationControlSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationControlSection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	section := idname.Ref{}
	resp.Diagnostics.Append(state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	remove := cato_models.PolicyRemoveSectionInput{ID: section.ID.ValueString()}
	if _, err := r.client.Catov2.PolicyApplicationControlRemoveSection(ctx, remove, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlRemoveSection error", err.Error())
		return
	}
	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
