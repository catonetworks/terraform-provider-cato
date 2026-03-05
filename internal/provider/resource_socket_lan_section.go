package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &socketLanSectionResource{}
	_ resource.ResourceWithConfigure   = &socketLanSectionResource{}
	_ resource.ResourceWithImportState = &socketLanSectionResource{}
)

func NewSocketLanSectionResource() resource.Resource {
	return &socketLanSectionResource{}
}

type socketLanSectionResource struct {
	client *catoClientData
}

func (r *socketLanSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_section"
}

func (r *socketLanSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_section` resource contains the configuration parameters necessary to add a Socket LAN firewall section. Documentation for the underlying API used in this resource can be found at [mutation.policy.socketLan.addSection()](https://api.catonetworks.com/documentation/#mutation-policy.socketLan.addSection).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the section in the policy",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy or another section (AFTER_SECTION, BEFORE_SECTION, LAST_IN_POLICY)",
						Required:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_SECTION", "BEFORE_SECTION", "LAST_IN_POLICY"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the section relative to which the position of the added section is defined",
						Required:    false,
						Optional:    true,
					},
				},
			},
			"section": schema.SingleNestedAttribute{
				Description: "Section parameters",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Section ID",
						Computed:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Section Name",
						Required:    true,
						Optional:    false,
					},
				},
			},
		},
	}
}

func (r *socketLanSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketLanSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *socketLanSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan SocketLanSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PolicyAddSectionInput{}

	// setting at
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicySectionPositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.At.Position = (cato_models.PolicySectionPositionEnum)(positionInput.Position.ValueString())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting section
	if !plan.Section.IsNull() {
		input.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := PolicyAddSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.Section.Name = sectionInput.Name.ValueString()
	}

	tflog.Debug(ctx, "Create.PolicySocketLanAddSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	policyChange, err := r.client.catov2.PolicySocketLanAddSection(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanAddSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanAddSection error",
			err.Error(),
		)
		return
	}
	if len(policyChange.Policy.SocketLan.AddSection.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.AddSection.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
			return
		}
	}

	// publishing new section
	tflog.Info(ctx, "Create.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanPublishPolicyRevision.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	plan.Id = types.StringValue(policyChange.GetPolicy().GetSocketLan().GetAddSection().Section.GetSection().ID)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overriding state with section id
	resp.State.SetAttribute(
		ctx,
		path.Root("section").AtName("id"),
		policyChange.GetPolicy().GetSocketLan().GetAddSection().Section.GetSection().ID,
	)
}

func (r *socketLanSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state SocketLanSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	tflog.Debug(ctx, "Read.PolicySocketLanPolicy.response", map[string]interface{}{
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
	section := PolicyUpdateSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionList := body.GetPolicy().SocketLan.Policy.GetSections()
	sectionExist := false
	for _, sectionListItem := range sectionList {
		if sectionListItem.GetSection().ID == section.Id.ValueString() {
			sectionExist = true
			state.Id = types.StringValue(sectionListItem.GetSection().ID)
			curSectionObj, diagstmp := types.ObjectValue(
				NameIDAttrTypes,
				map[string]attr.Value{
					"id":   types.StringValue(sectionListItem.GetSection().ID),
					"name": types.StringValue(sectionListItem.GetSection().GetName()),
				},
			)
			diags = append(diags, diagstmp...)
			state.Section = curSectionObj
		}
	}

	// Hard coding LAST_IN_POLICY position as the API does not return any value
	curAtObj, diagstmp := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		},
	)
	state.At = curAtObj
	diags = append(diags, diagstmp...)

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "socket lan section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *socketLanSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan SocketLanSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection := cato_models.PolicyUpdateSectionInput{}
	inputMoveSection := cato_models.PolicyMoveSectionInput{}

	// setting section
	if !plan.Section.IsNull() {
		inputUpdateSection.Section = &cato_models.PolicyUpdateSectionInfoInput{}
		sectionInput := PolicyUpdateSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputUpdateSection.Section.Name = sectionInput.Name.ValueStringPointer()
		inputUpdateSection.ID = sectionInput.Id.ValueString()
	}

	// setting at
	if !plan.At.IsNull() {
		inputMoveSection.To = &cato_models.PolicySectionPositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveSection.To.Position = (cato_models.PolicySectionPositionEnum)(positionInput.Position.ValueString())
		inputMoveSection.To.Ref = positionInput.Ref.ValueStringPointer()
		inputMoveSection.ID = inputUpdateSection.ID
	}

	// move section
	tflog.Debug(ctx, "Update.PolicySocketLanMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.client.catov2.PolicySocketLanMoveSection(ctx, inputMoveSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanMoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanMoveSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.SocketLan.MoveSection.Status != "SUCCESS" {
		for _, item := range moveSection.Policy.SocketLan.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// update section
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.client.catov2.PolicySocketLanUpdateSection(ctx, nil, inputUpdateSection, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateSection),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanUpdateSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.SocketLan.UpdateSection.Status != "SUCCESS" {
		for _, item := range updateSection.Policy.SocketLan.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing section changes
	tflog.Info(ctx, "Update.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
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

func (r *socketLanSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state SocketLanSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve section ID
	section := PolicyAddSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := cato_models.PolicyRemoveSectionInput{
		ID: section.Id.ValueString(),
	}

	PolicySocketLanRemoveSectionResponse, err := r.client.catov2.PolicySocketLanRemoveSection(ctx, nil, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(PolicySocketLanRemoveSectionResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Delete.publishing-section")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
