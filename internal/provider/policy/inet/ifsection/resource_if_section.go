package ifsection

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

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &internetFwSectionResource{}
	_ resource.ResourceWithConfigure   = &internetFwSectionResource{}
	_ resource.ResourceWithImportState = &internetFwSectionResource{}
)

func NewInternetFwSectionResource() resource.Resource {
	return &internetFwSectionResource{}
}

type internetFwSectionResource struct {
	client *catoclient.Service
}

func (r *internetFwSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_section"
}

func (r *internetFwSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *internetFwSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

//nolint:funlen
func (r *internetFwSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InternetFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PolicyAddSectionInput{}

	// Setting at
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.At.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// Setting section
	if !plan.Section.IsNull() {
		input.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := PolicyAddSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.Section.Name = sectionInput.Name.ValueString()
	}

	sectionIndexAPIData, err := r.client.Catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexAPIData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}
	if len(sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			}}
		sectionCreateAPIData, err := r.client.Catov2.PolicyInternetFirewallAddSection(
			ctx,
			&cato_models.InternetFirewallPolicyMutationInput{},
			input,
			r.client.AccountId,
		)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinSection.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateAPIData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Create.PolicyInternetFirewallAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.client.Catov2.PolicyInternetFirewallAddSection(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		input,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Create.PolicyInternetFirewallAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddSection error",
			err.Error(),
		)
		return
	}
	if len(policyChange.Policy.InternetFirewall.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.InternetFirewall.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// Publishing new section
	tflog.Info(ctx, "Create.publishing-rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Create.PolicyInternetFirewallPublishPolicyRevision.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(policyChange.GetPolicy().GetInternetFirewall().GetAddSection().Section.GetSection().ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with rule id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetInternetFirewall().GetAddSection().Section.GetSection().ID,
	)
}

func (r *internetFwSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InternetFirewallSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.Catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewall.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// Retrieve section ID
	section := PolicyUpdateSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionList := body.GetPolicy().InternetFirewall.Policy.GetSections()
	sectionExist := false
	for _, sectionListItem := range sectionList {
		if sectionListItem.GetSection().ID != section.ID.ValueString() {
			continue
		}

		sectionExist = true
		state.ID = types.StringValue(sectionListItem.GetSection().ID)
		curSectionObj, diagstmp := types.ObjectValue(
			idname.RefAttrTypes,
			map[string]attr.Value{
				"id":   types.StringValue(sectionListItem.GetSection().ID),
				"name": types.StringValue(sectionListItem.GetSection().GetName()),
			},
		)
		resp.Diagnostics.Append(diagstmp...)
		state.Section = curSectionObj
	}

	// The API does not return section position. Preserve configuration state to
	// avoid drift, defaulting imports without position state to LAST_IN_POLICY.
	curAtObj, diagstmp := convert.DefaultSectionPosition(ctx, state.At)
	state.At = curAtObj
	resp.Diagnostics.Append(diagstmp...)

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "internet section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:funlen
func (r *internetFwSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InternetFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection := cato_models.PolicyUpdateSectionInput{}
	inputMoveSection := cato_models.PolicyMoveSectionInput{}

	// Setting section
	if !plan.Section.IsNull() {
		inputUpdateSection.Section = &cato_models.PolicyUpdateSectionInfoInput{}
		sectionInput := PolicyUpdateSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputUpdateSection.Section.Name = sectionInput.Name.ValueStringPointer()
		inputUpdateSection.ID = sectionInput.ID.ValueString()
	}

	// Setting at
	if !plan.At.IsNull() {
		inputMoveSection.To = &cato_models.PolicySectionPositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveSection.To.Position = cato_models.PolicySectionPositionEnum(positionInput.Position.ValueString())
		inputMoveSection.To.Ref = positionInput.Ref.ValueStringPointer()
		inputMoveSection.ID = inputUpdateSection.ID
	}

	// Move section
	tflog.Debug(ctx, "Update.PolicyInternetFirewallMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.client.Catov2.PolicyInternetFirewallMoveSection(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		inputMoveSection,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Update.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.InternetFirewall.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.InternetFirewall.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Update section
	tflog.Debug(ctx, "Update.PolicyInternetFirewallUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.client.Catov2.PolicyInternetFirewallUpdateSection(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		inputUpdateSection,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Update.PolicyInternetFirewallUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.InternetFirewall.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.InternetFirewall.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing new section
	tflog.Info(ctx, "Update.publishing-rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.client.AccountId,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
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

func (r *internetFwSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InternetFirewallSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve section ID
	section := PolicyAddSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := cato_models.PolicyRemoveSectionInput{
		ID: section.ID.ValueString(),
	}

	policyInternetFirewallRemoveSectionResponse, err := r.client.Catov2.PolicyInternetFirewallRemoveSection(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		removeSection,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Delete.PolicyInternetFirewallRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyInternetFirewallRemoveSectionResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Delete.publishing-rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
