package tlssection

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &tlsInspectionSectionResource{}
	_ resource.ResourceWithConfigure   = &tlsInspectionSectionResource{}
	_ resource.ResourceWithImportState = &tlsInspectionSectionResource{}
)

func NewTLSInspectionSectionResource() resource.Resource {
	return &tlsInspectionSectionResource{}
}

type tlsInspectionSectionResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *tlsInspectionSectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_tls_section"
}

func (r *tlsInspectionSectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *tlsInspectionSectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *tlsInspectionSectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan InspectionSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.PolicyTLSInspectAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.resourceAPI().PolicyTLSInspectAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicyTLSInspectAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})

	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyTLSInspectAddSection error", err.Error())
		return
	}
	if len(policyChange.Policy.TLSInspect.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.TLSInspect.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// publishing new section
	tflog.Info(ctx, "publishing new section")
	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	convertCreateInput2(&plan, policyChange)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overriding state with section id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetTLSInspect().GetAddSection().Section.GetSection().ID,
	)
}

func (r *tlsInspectionSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InspectionSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.Tlsinspectpolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// retrieve section ID
	section := convertReadState(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionExist, diagstmp := convertReadState2(&state, body, section, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "tls inspection section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tlsInspectionSectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan InspectionSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection, inputMoveSection := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Update.PolicyTLSInspectMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.resourceAPI().PolicyTLSInspectMoveSection(ctx, inputMoveSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyTLSInspectMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectMoveSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.TLSInspect.MoveSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveSection.Policy.TLSInspect.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Debug(ctx, "Update.PolicyTLSInspectUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.resourceAPI().PolicyTLSInspectUpdateSection(ctx, inputUpdateSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicyTLSInspectUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectUpdateSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.TLSInspect.UpdateSection.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateSection.Policy.TLSInspect.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Section",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing updated section
	tflog.Info(ctx, "publishing updated section")
	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectPublishPolicyRevision error",
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

func (r *tlsInspectionSectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state InspectionSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve section ID
	section := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := convertDeleteInput2(section)

	tflog.Debug(ctx, "Delete.PolicyTLSInspectRemoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeSection),
	})
	policyTLSInspectRemoveSectionResponse, err := r.resourceAPI().PolicyTLSInspectRemoveSection(ctx, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyTLSInspectRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyTLSInspectRemoveSectionResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyTLSInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
