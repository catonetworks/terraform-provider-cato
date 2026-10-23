package alwaysonrule

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &alwaysOnRuleResource{}
	_ resource.ResourceWithConfigure   = &alwaysOnRuleResource{}
	_ resource.ResourceWithImportState = &alwaysOnRuleResource{}
)

type alwaysOnRuleResource struct {
	client     *catoclient.Service
	mockClient Client
}

// NewAlwaysOnRuleResource returns the cato_always_on_rule resource.
func NewAlwaysOnRuleResource() resource.Resource {
	return &alwaysOnRuleResource{}
}

func (r *alwaysOnRuleResource) getClient() Client {
	if r.mockClient != nil {
		return r.mockClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *alwaysOnRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_always_on_rule"
}

func (r *alwaysOnRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *alwaysOnRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *alwaysOnRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlwaysOnRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildAlwaysOnRuleInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil || r.getClient() == nil {
		resp.Diagnostics.AddError("Resource not configured", "The Cato client must be configured before creating an Always On rule.")
		return
	}

	errMsg := fmt.Sprintf("failed to create Always On rule %s", input.Rule.Name)
	result, err := r.getClient().PolicyZtnaAlwaysOnAddRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetAddRule()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, &resp.Diagnostics) {
		return
	}
	rule := payload.GetRule().GetRule()
	if rule.GetID() == "" {
		resp.Diagnostics.AddError(errMsg, "API did not return resource ID")
		return
	}
	if *payload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		resp.Diagnostics.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", *payload.GetStatus()))
		return
	}

	plan.Rule = parseAlwaysOnRule(ctx, rule, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alwaysOnRuleResource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("Read not implemented", "Reading Always On rules is not implemented yet.")
}

func (r *alwaysOnRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not implemented", "Updating Always On rules is not implemented yet.")
}

func (r *alwaysOnRuleResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete not implemented", "Deleting Always On rules is not implemented yet.")
}
