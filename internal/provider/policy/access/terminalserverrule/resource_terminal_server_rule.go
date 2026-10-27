package terminalserverrule

import (
	"context"
	"fmt"
	"strconv"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = (*policyResource)(nil)
	_ resource.ResourceWithConfigure   = (*policyResource)(nil)
	_ resource.ResourceWithImportState = (*policyResource)(nil)
)

type policyResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the cato_terminal_server_rule resource.
func NewResource() resource.Resource { return &policyResource{} }

func (r *policyResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminal_server_rule"
}

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *policyResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.getClient() == nil {
		diags.AddError("Resource not configured", "The Cato client must be configured before managing this resource.")
		return false
	}
	return true
}

func (r *policyResource) ready(id types.String, diags *diag.Diagnostics) bool {
	return !diags.HasError() && validID(id, diags) && r.configured(diags)
}

// Create retains identity before publishing so a failed publication can be recovered.
func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	object := buildCreateObject(ctx, plan.Rule, &resp.Diagnostics)
	at := buildPosition(ctx, plan.At, &resp.Diagnostics)
	validatePosition(at, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	if at == nil {
		position := cato_models.PolicyRulePositionEnumLastInPolicy
		at = &cato_models.PolicyRulePositionInput{Position: &position}
	}
	revision := r.beginRevision(ctx, "", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	retained := false
	defer func() {
		if !retained {
			r.discard(ctx, revision, &resp.Diagnostics)
		}
	}()
	if resp.Private != nil {
		rememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	result, err := r.getClient().PolicyTerminalServerAddRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.TerminalServerAddRuleInput{At: at, Rule: object})
	payload := result.GetPolicy().GetTerminalServer().GetAddRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to create rule", &resp.Diagnostics) {
		return
	}
	if !checkStatus(payload.GetStatus(), &resp.Diagnostics) {
		return
	}
	returned := payload.GetRule().GetRule()
	if returned.GetID() == "" {
		resp.Diagnostics.AddError("Invalid API response", "Create did not return a resource ID.")
		return
	}
	plan.Rule = parseAdd(ctx, returned, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retained = true
	if r.publish(ctx, revision, &resp.Diagnostics) && resp.Private != nil {
		rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id)...)
	if !r.ready(id, &resp.Diagnostics) {
		return
	}
	// A previous apply already authorized this revision; refresh retries its failed publication.
	if req.Private != nil {
		revision := pendingRevision(ctx, req.Private, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if revision != "" {
			if !r.publish(ctx, revision, &resp.Diagnostics) {
				return
			}
			if resp.Private != nil {
				rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
			}
		}
	}
	policy, err := r.readPolicy(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read rule", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Rule = parseRead(ctx, object, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id)...)
	if resp.Diagnostics.HasError() || !validID(id, &resp.Diagnostics) {
		return
	}
	object := buildUpdateObject(ctx, plan.Rule, &resp.Diagnostics)
	at := buildPosition(ctx, plan.At, &resp.Diagnostics)
	validatePosition(at, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	revision := r.updateRevision(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = r.updateObject(ctx, id.ValueString(), object, revision, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	requestedAt := plan.At
	plan.At = state.At
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if at != nil && !requestedAt.Equal(state.At) {
		plan.Rule = r.move(ctx, id.ValueString(), at, revision, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	plan.At = requestedAt
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.publish(ctx, revision, &resp.Diagnostics) && resp.Private != nil {
		rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id)...)
	if !r.ready(id, &resp.Diagnostics) {
		return
	}
	pending := ""
	if req.Private != nil {
		pending = pendingRevision(ctx, req.Private, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := r.readPolicy(ctx, pending)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read rule before deletion", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil && pending == "" {
		return
	}

	revision := r.beginRevision(ctx, pending, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Private != nil {
		rememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	if object != nil && !r.removeObject(ctx, id.ValueString(), revision, &resp.Diagnostics) {
		return
	}
	if r.publish(ctx, revision, &resp.Diagnostics) && resp.Private != nil {
		rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
}

func (r *policyResource) updateRevision(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) string {
	pending := ""
	if req.Private != nil {
		pending = pendingRevision(ctx, req.Private, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return ""
	}
	revision := r.beginRevision(ctx, pending, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return ""
	}
	if resp.Private != nil {
		rememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	return revision
}

func (r *policyResource) updateObject(ctx context.Context, id string, object *cato_models.TerminalServerUpdateRuleDataInput,
	revision string, diags *diag.Diagnostics,
) types.Object {
	result, err := r.getClient().PolicyTerminalServerUpdateRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.TerminalServerUpdateRuleInput{ID: id, Rule: object})
	payload := result.GetPolicy().GetTerminalServer().GetUpdateRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to update rule", diags) {
		return types.ObjectNull(ruleAttrTypes)
	}
	if !checkStatus(payload.GetStatus(), diags) {
		return types.ObjectNull(ruleAttrTypes)
	}
	returned := payload.GetRule().GetRule()
	if returned.GetID() != id {
		diags.AddError("Invalid API response", "Update returned a different resource ID.")
		return types.ObjectNull(ruleAttrTypes)
	}
	return parseUpdate(ctx, returned, diags)
}

func (r *policyResource) move(ctx context.Context, id string, at *cato_models.PolicyRulePositionInput,
	revision string, diags *diag.Diagnostics,
) types.Object {
	result, err := r.getClient().PolicyTerminalServerMoveRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyMoveRuleInput{ID: id, To: at})
	payload := result.GetPolicy().GetTerminalServer().GetMoveRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to move rule", diags) || !checkStatus(payload.GetStatus(), diags) {
		return types.ObjectNull(ruleAttrTypes)
	}
	object := payload.GetRule().GetRule()
	if object.GetID() != id {
		diags.AddError("Invalid API response", "Move returned a different resource ID.")
		return types.ObjectNull(ruleAttrTypes)
	}
	return parseMove(ctx, object, diags)
}

func mutationInput(revision string) *cato_models.TerminalServerPolicyMutationInput {
	return &cato_models.TerminalServerPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
}

func (r *policyResource) beginRevision(ctx context.Context, pending string, diags *diag.Diagnostics) string {
	if pending != "" {
		return pending
	}
	result, err := r.getClient().PolicyTerminalServerCreatePolicyRevision(ctx, r.client.AccountId, nil,
		cato_models.PolicyCreateRevisionInput{Name: "Terraform terminal_server_rule", Description: "Managed by Terraform"})
	payload := result.GetPolicy().GetTerminalServer().GetCreatePolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to create policy revision", diags) || !checkStatus(payload.GetStatus(), diags) {
		return ""
	}
	revision := payload.GetPolicy().GetRevision().GetID()
	if revision == "" {
		diags.AddError("Invalid API response", "CreatePolicyRevision did not return a revision ID.")
	}
	return revision
}

func (r *policyResource) publish(ctx context.Context, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicyTerminalServerPublishPolicyRevision(
		ctx, r.client.AccountId, mutationInput(revision), &cato_models.PolicyPublishRevisionInput{},
	)
	payload := result.GetPolicy().GetTerminalServer().GetPublishPolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to publish policy revision "+revision, diags) {
		return false
	}
	return checkStatus(payload.GetStatus(), diags)
}

func (r *policyResource) discard(ctx context.Context, revision string, diags *diag.Diagnostics) {
	result, err := r.getClient().PolicyTerminalServerDiscardPolicyRevision(ctx, r.client.AccountId, mutationInput(revision),
		&cato_models.PolicyDiscardRevisionInput{ID: &revision})
	payload := result.GetPolicy().GetTerminalServer().GetDiscardPolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to discard the resource's policy revision "+revision, diags) {
		return
	}
	checkStatus(payload.GetStatus(), diags)
}

func checkStatus(status *cato_models.PolicyMutationStatus, diags *diag.Diagnostics) bool {
	if status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		diags.AddError("Policy mutation failed", "API did not return SUCCESS.")
		return false
	}
	return true
}

// revisionState is satisfied by Terraform's private state without exposing internal framework types.
type revisionState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

func rememberRevision(ctx context.Context, store revisionState, revision string, diags *diag.Diagnostics) {
	var data []byte
	if revision != "" {
		data = []byte(strconv.Quote(revision))
	}
	diags.Append(store.SetKey(ctx, "pending_revision", data)...)
}

func pendingRevision(ctx context.Context, store revisionState, diags *diag.Diagnostics) string {
	data, d := store.GetKey(ctx, "pending_revision")
	diags.Append(d...)
	if len(data) == 0 || diags.HasError() {
		return ""
	}
	revision, err := strconv.Unquote(string(data))
	if err != nil {
		diags.AddError("Invalid private state", "Cannot decode the pending policy revision ID.")
	}
	return revision
}

func (r *policyResource) removeObject(ctx context.Context, id, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicyTerminalServerRemoveRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.TerminalServerRemoveRuleInput{ID: id})
	payload := result.GetPolicy().GetTerminalServer().GetRemoveRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to delete rule", diags) {
		return false
	}
	return checkStatus(payload.GetStatus(), diags)
}
