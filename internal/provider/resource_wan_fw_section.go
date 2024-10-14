package provider

import (
	"context"
	"fmt"

	cato_models "github.com/cato-networks/cato-go-sdk/models"
	"github.com/cato-networks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &wanFwSectionResource{}
	_ resource.ResourceWithConfigure = &wanFwSectionResource{}
)

func NewWanFwSectionResource() resource.Resource {
	return &wanFwSectionResource{}
}

type wanFwSectionResource struct {
	client *catoClientData
}

func (r *wanFwSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_section"
}

func (r *wanFwSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wf_section` resource contains the configuration parameters necessary to WAN firewall section (https://support.catonetworks.com/hc/en-us/articles/5590037900701-Adding-Sections-to-the-WAN-and-Internet-Firewalls). Documentation for the underlying API used in this resource can be found at[mutation.policy.wanFirewall.addSection()](https://api.catonetworks.com/documentation/#mutation-policy.wanFirewall.addSection).",
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

func (r *wanFwSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanFwSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan WanFirewallSection
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

	tflog.Debug(ctx, "wan_fw_section create", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	//creating new section
	policyChange, err := r.client.catov2.PolicyWanFirewallAddSection(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddSection error",
			err.Error(),
		)
		return
	}

	//publishing new section
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
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
		policyChange.GetPolicy().GetWanFirewall().GetAddSection().Section.GetSection().ID,
	)

}

func (r *wanFwSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state WanFirewallSection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
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

	sectionList := body.GetPolicy().WanFirewall.Policy.GetSections()
	sectionExist := false
	for _, sectionListItem := range sectionList {
		if sectionListItem.GetSection().ID == section.Id.ValueString() {
			sectionExist = true

			// Need to refresh STATE
		}
	}

	// remove resource if it doesn't exist anymore
	if !sectionExist {
		tflog.Warn(ctx, "wan section not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *wanFwSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan WanFirewallSection
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

	tflog.Debug(ctx, "wab_fw_section update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputUpdateSection),
	})

	tflog.Debug(ctx, "wan_fw_section move", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputMoveSection),
	})

	//move section
	moveSection, err := r.client.catov2.PolicyWanFirewallMoveSection(ctx, inputMoveSection, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveSection.Policy.WanFirewall.MoveSection.Status != "SUCCESS" {
		for _, item := range moveSection.Policy.WanFirewall.MoveSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//update section
	updateSection, err := r.client.catov2.PolicyWanFirewallUpdateSection(ctx, inputUpdateSection, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddSection error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateSection.Policy.WanFirewall.UpdateSection.Status != "SUCCESS" {
		for _, item := range updateSection.Policy.WanFirewall.UpdateSection.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//publishing new section
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
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

func (r *wanFwSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state WanFirewallSection
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

	_, err := r.client.catov2.PolicyWanFirewallRemoveSection(ctx, removeSection, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

}
