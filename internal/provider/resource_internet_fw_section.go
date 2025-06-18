package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	client *catoClientData
}

func (r *internetFwSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_section"
}

func (r *internetFwSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_section` resource contains the configuration parameters necessary to Internet firewall section (https://support.catonetworks.com/hc/en-us/articles/5590037900701-Adding-Sections-to-the-WAN-and-Internet-Firewalls). Documentation for the underlying API used in this resource can be found at[mutation.policy.internetFirewall.addSection()](https://api.catonetworks.com/documentation/#mutation-policy.internetFirewall.addSection).",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule",
						Required:    true,
						Optional:    false,
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the object (e.g. a rule, a section) relative to which the position of the added rule is defined",
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

func (r *internetFwSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *internetFwSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("section").AtName("id"), req, resp)
}

func (r *internetFwSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan InternetFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PolicyAddSectionInput{}

	//setting at
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicySectionPositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.At.Position = (cato_models.PolicySectionPositionEnum)(positionInput.Position.ValueString())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	//setting section
	if !plan.Section.IsNull() {
		input.Section = &cato_models.PolicyAddSectionInfoInput{}
		sectionInput := PolicyAddSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.Section.Name = sectionInput.Name.ValueString()
	}

	sectionIndexApiData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}
	if len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			}}
		sectionCreateApiData, err := r.client.catov2.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinSection.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateApiData),
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
	policyChange, err := r.client.catov2.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
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

	//publishing new section
	tflog.Info(ctx, "Create.publishing-rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
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
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
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

	//retrieve section ID
	section := PolicyUpdateSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionList := body.GetPolicy().InternetFirewall.Policy.GetSections()
	sectionExist := false
	for _, sectionListItem := range sectionList {
		if sectionListItem.GetSection().ID == section.Id.ValueString() {
			sectionExist = true

			// Need to refresh STATE
			resp.State.SetAttribute(
				ctx,
				path.Root("section").AtName("id"),
				sectionListItem.GetSection().ID,
			)
		}
	}

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

func (r *internetFwSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan InternetFirewallSection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputUpdateSection := cato_models.PolicyUpdateSectionInput{}
	inputMoveSection := cato_models.PolicyMoveSectionInput{}

	//setting section
	if !plan.Section.IsNull() {
		inputUpdateSection.Section = &cato_models.PolicyUpdateSectionInfoInput{}
		sectionInput := PolicyUpdateSectionInfoInput{}
		diags = plan.Section.As(ctx, &sectionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputUpdateSection.Section.Name = sectionInput.Name.ValueStringPointer()
		inputUpdateSection.ID = sectionInput.Id.ValueString()
	}

	//setting at
	if !plan.At.IsNull() {
		inputMoveSection.To = &cato_models.PolicySectionPositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveSection.To.Position = (cato_models.PolicySectionPositionEnum)(positionInput.Position.ValueString())
		inputMoveSection.To.Ref = positionInput.Ref.ValueStringPointer()
		inputMoveSection.ID = inputUpdateSection.ID
	}

	//move section
	tflog.Debug(ctx, "Update.PolicyInternetFirewallMoveSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputMoveSection),
	})
	moveSection, err := r.client.catov2.PolicyInternetFirewallMoveSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, inputMoveSection, r.client.AccountId)
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
	if moveSection.Policy.InternetFirewall.MoveSection.Status != "SUCCESS" {
		for _, item := range moveSection.Policy.InternetFirewall.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Section Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//update section
	tflog.Debug(ctx, "Update.PolicyInternetFirewallUpdateSection.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSection),
	})
	updateSection, err := r.client.catov2.PolicyInternetFirewallUpdateSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, inputUpdateSection, r.client.AccountId)
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
	if updateSection.Policy.InternetFirewall.UpdateSection.Status != "SUCCESS" {
		for _, item := range updateSection.Policy.InternetFirewall.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//publishing new section
	tflog.Info(ctx, "Update.publishing-rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)

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

	//retrieve section ID
	section := PolicyAddSectionInfoInput{}
	diags = state.Section.As(ctx, &section, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeSection := cato_models.PolicyRemoveSectionInput{
		ID: section.Id.ValueString(),
	}

	PolicyInternetFirewallRemoveSectionResponse, err := r.client.catov2.PolicyInternetFirewallRemoveSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, removeSection, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyInternetFirewallRemoveSection.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(PolicyInternetFirewallRemoveSectionResponse),
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
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

}
