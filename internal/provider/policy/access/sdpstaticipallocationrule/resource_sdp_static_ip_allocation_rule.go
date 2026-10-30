package sdpstaticipallocationrule

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
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

// NewResource returns the cato_sdp_static_ip_allocation_rule resource.
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
	resp.TypeName = req.ProviderTypeName + "_sdp_static_ip_allocation_rule"
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
	return !diags.HasError() && policyshared.ValidID(id, diags) && r.configured(diags)
}

// Create retains identity before publishing so a failed publication can be recovered.
func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	object := buildCreateObject(ctx, plan.Rule, &resp.Diagnostics)
	at := policyshared.BuildRulePosition(ctx, plan.At, &resp.Diagnostics)
	policyshared.ValidateRulePosition(at, &resp.Diagnostics)
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
		policyshared.RememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	result, err := r.getClient().PolicySdpStaticIPAllocationAddRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.SdpStaticIPAllocationAddRuleInput{At: at, Rule: object})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetAddRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to create rule", &resp.Diagnostics) {
		return
	}
	if !policyshared.CheckStatus(payload.GetStatus(), &resp.Diagnostics) {
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
	if r.publishOrReconcile(ctx, revision, returned.GetID(), plan.Rule, &resp.Diagnostics) && resp.Private != nil {
		policyshared.RememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
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
	policy, err := r.readPolicy(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read rule", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil {
		// A failed create may exist only in its owned draft. Keep its identity for destroy.
		if req.Private != nil {
			revision := policyshared.PendingRevision(ctx, req.Private, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			if revision != "" {
				draft, draftErr := r.readPolicy(ctx, revision)
				if draftErr != nil && !errors.Is(draftErr, errMissingRevision) {
					resp.Diagnostics.AddError("Failed to inspect pending revision", draftErr.Error())
					return
				}
				if draftErr == nil && findObject(draft, id.ValueString()) != nil {
					return
				}
			}
		}
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
	if resp.Diagnostics.HasError() || !policyshared.ValidID(id, &resp.Diagnostics) {
		return
	}
	object := buildUpdateObject(ctx, plan.Rule, &resp.Diagnostics)
	at := policyshared.BuildRulePosition(ctx, plan.At, &resp.Diagnostics)
	policyshared.ValidateRulePosition(at, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	revision := r.updateRevision(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	retained := false
	defer func() {
		if !retained {
			r.discard(ctx, revision, &resp.Diagnostics)
		}
	}()
	plan.Rule = r.updateObject(ctx, id.ValueString(), object, revision, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if at != nil && !plan.At.Equal(state.At) {
		plan.Rule = r.move(ctx, id.ValueString(), at, revision, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	retained = true
	if !r.publishOrReconcile(ctx, revision, id.ValueString(), plan.Rule, &resp.Diagnostics) {
		return
	}
	if resp.Private != nil {
		policyshared.RememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id)...)
	if !r.ready(id, &resp.Diagnostics) {
		return
	}
	pending := ""
	if req.Private != nil {
		pending = policyshared.PendingRevision(ctx, req.Private, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	// Dispose only this resource's unfinished draft before examining published state.
	if pending != "" {
		if !r.discard(ctx, pending, &resp.Diagnostics) {
			return
		}
		if resp.Private != nil {
			policyshared.RememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
		}
		pending = ""
	}
	policy, err := r.readPolicy(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read rule before deletion", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil {
		return
	}

	revision := r.beginRevision(ctx, pending, &resp.Diagnostics)
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
		policyshared.RememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	if !r.removeObject(ctx, id.ValueString(), revision, &resp.Diagnostics) {
		return
	}
	retained = true
	if r.publishOrReconcile(ctx, revision, id.ValueString(), types.ObjectNull(ruleAttrTypes), &resp.Diagnostics) && resp.Private != nil {
		policyshared.RememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
}

func (r *policyResource) updateRevision(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) string {
	pending := ""
	if req.Private != nil {
		pending = policyshared.PendingRevision(ctx, req.Private, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return ""
	}
	revision := r.beginRevision(ctx, pending, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return ""
	}
	if resp.Private != nil {
		policyshared.RememberRevision(ctx, resp.Private, revision, &resp.Diagnostics)
	}
	return revision
}

func (r *policyResource) updateObject(ctx context.Context, id string, object *cato_models.SdpStaticIPAllocationUpdateRuleDataInput,
	revision string, diags *diag.Diagnostics,
) types.Object {
	result, err := r.getClient().PolicySdpStaticIPAllocationUpdateRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.SdpStaticIPAllocationUpdateRuleInput{ID: id, Rule: object})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetUpdateRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to update rule", diags) {
		return types.ObjectNull(ruleAttrTypes)
	}
	if !policyshared.CheckStatus(payload.GetStatus(), diags) {
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
	result, err := r.getClient().PolicySdpStaticIPAllocationMoveRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyMoveRuleInput{ID: id, To: at})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetMoveRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to move rule", diags) || !policyshared.CheckStatus(payload.GetStatus(), diags) {
		return types.ObjectNull(ruleAttrTypes)
	}
	object := payload.GetRule().GetRule()
	if object.GetID() != id {
		diags.AddError("Invalid API response", "Move returned a different resource ID.")
		return types.ObjectNull(ruleAttrTypes)
	}
	return parseMove(ctx, object, diags)
}

func mutationInput(revision string) *cato_models.SdpStaticIPAllocationPolicyMutationInput {
	return &cato_models.SdpStaticIPAllocationPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
}

func (r *policyResource) beginRevision(ctx context.Context, pending string, diags *diag.Diagnostics) string {
	if pending != "" && !r.discard(ctx, pending, diags) {
		return ""
	}
	result, err := r.getClient().PolicySdpStaticIPAllocationCreatePolicyRevision(ctx, r.client.AccountId, nil,
		cato_models.PolicyCreateRevisionInput{Name: "terraform-" + rand.Text(), Description: "Managed by Terraform"})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetCreatePolicyRevision()
	if utils.CheckAPIErrors(err,
		payload.GetErrors(),
		"Failed to create policy revision",
		diags) ||
		!policyshared.CheckStatus(payload.GetStatus(),
			diags) {
		return ""
	}
	revision := payload.GetPolicy().GetRevision().GetID()
	if revision == "" {
		diags.AddError("Invalid API response", "CreatePolicyRevision did not return a revision ID.")
	}
	return revision
}

func (r *policyResource) publish(ctx context.Context, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx, r.client.AccountId, mutationInput(revision), &cato_models.PolicyPublishRevisionInput{},
	)
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetPublishPolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to publish policy revision "+revision, diags) {
		return false
	}
	return policyshared.CheckStatus(payload.GetStatus(), diags)
}

func (r *policyResource) discard(ctx context.Context, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicySdpStaticIPAllocationDiscardPolicyRevision(ctx, r.client.AccountId, mutationInput(revision),
		&cato_models.PolicyDiscardRevisionInput{ID: &revision})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetDiscardPolicyRevision()
	if policyshared.RevisionMissing(err) {
		return true
	}
	for _, item := range payload.GetErrors() {
		if code := item.GetErrorCode(); code != nil && *code == "PolicyRevisionNotFound" {
			return true
		}
	}
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to discard the resource's policy revision "+revision, diags) {
		return false
	}
	return policyshared.CheckStatus(payload.GetStatus(), diags)
}

func (r *policyResource) removeObject(ctx context.Context, id, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicySdpStaticIPAllocationRemoveRule(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.SdpStaticIPAllocationRemoveRuleInput{ID: id})
	payload := result.GetPolicy().GetSdpStaticIPAllocation().GetRemoveRule()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to delete rule", diags) {
		return false
	}
	return policyshared.CheckStatus(payload.GetStatus(), diags)
}

// A publication can commit even when its response is lost. Reconcile only after
// the owned revision is confirmed gone, and only against published state.
func (r *policyResource) publishOrReconcile(ctx context.Context, revision, id string, expected types.Object, diags *diag.Diagnostics) bool {
	var publication diag.Diagnostics
	if r.publish(ctx, revision, &publication) {
		return true
	}
	_, err := r.readPolicy(ctx, revision)
	if errors.Is(err, errMissingRevision) {
		policy, readErr := r.readPolicy(ctx, "")
		if readErr == nil {
			object := findObject(policy, id)
			if expected.IsNull() && object == nil {
				return true
			}
			if !expected.IsNull() && object != nil {
				var conversion diag.Diagnostics
				actual := parseRead(ctx, object, &conversion)
				if !conversion.HasError() && actual.Equal(expected) {
					return true
				}
			}
		}
	}
	diags.Append(publication...)
	return false
}
