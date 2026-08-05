package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource                = &lfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &lfSubPolicyResource{}
	_ resource.ResourceWithImportState = &lfSubPolicyResource{}
)

func NewLfSubPolicyResource() resource.Resource {
	return &lfSubPolicyResource{}
}

type lfSubPolicyResource struct {
	client *catoClientData
}

func (r *lfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lf_sub_policy"
}

func (r *lfSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "The `cato_lf_sub_policy` resource manages a LAN Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Sub-policy ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "Sub-policy name. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Description:   "Sub-policy description. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope_rule_id": schema.StringAttribute{
				Description:   "ID of the underlying SUB_POLICY_SCOPE rule.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"at": schema.SingleNestedAttribute{
				Description:   "Position of the sub-policy scope within the LAN Firewall policy.",
				Required:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule.",
						Required:    true,
					},
					"ref": schema.StringAttribute{
						Description: "Identifier of the object relative to which the position is defined.",
						Optional:    true,
					},
				},
			},
			"scope": r.policyScopeSchema(),
		},
	}
}

func (r *lfSubPolicyResource) policyScopeSchema() schema.SingleNestedAttribute {
	var ruleSchema resource.SchemaResponse
	(&socketLanNetworkRuleResource{}).Schema(context.Background(), resource.SchemaRequest{}, &ruleSchema)

	scopeAttr := ruleSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	scopeAttr.Description = "Policy scope attributes"
	scopeAttr.Required = true
	delete(scopeAttr.Attributes, "transport")
	scopeAttr.Attributes["description"] = schema.StringAttribute{
		Description:   "API-managed scope description, synchronized with the sub-policy description.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	scopeAttr.Attributes["name"] = schema.StringAttribute{
		Description:   "API-managed scope name, synchronized with the sub-policy name.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}

	return scopeAttr
}

func (r *lfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *lfSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *lfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *lfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *lfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *lfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
