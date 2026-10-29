package pasection

import (
	"context"
	"crypto/rand"
	"errors"
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

// NewResource returns the cato_private_access_section resource.
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
	resp.TypeName = req.ProviderTypeName + "_private_access_section"
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
	object := buildCreateObject(ctx, plan.Section, &resp.Diagnostics)
	at := buildPosition(ctx, plan.At, &resp.Diagnostics)
	validatePosition(at, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
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
	result, err := r.getClient().PolicyPrivateAccessAddSection(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyAddSectionInput{At: at, Section: object})
	payload := result.GetPolicy().GetPrivateAccess().GetAddSection()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to create section", &resp.Diagnostics) {
		return
	}
	if !checkStatus(payload.GetStatus(), &resp.Diagnostics) {
		return
	}
	returned := payload.GetSection().GetSection()
	if returned.GetID() == "" {
		resp.Diagnostics.AddError("Invalid API response", "Create did not return a resource ID.")
		return
	}
	plan.Section = parseAdd(ctx, returned, &resp.Diagnostics)
	plan.ID = types.StringValue(returned.GetID())
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retained = true
	if r.publishOrReconcile(ctx, revision, returned.GetID(), plan.Section, &resp.Diagnostics, at) && resp.Private != nil {
		rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if !r.ready(id, &resp.Diagnostics) {
		return
	}
	policy, err := r.readPolicy(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read section", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil {
		// A failed create may exist only in its owned draft. Keep its identity for destroy.
		if req.Private != nil {
			revision := pendingRevision(ctx, req.Private, &resp.Diagnostics)
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
	state.Section = parseRead(ctx, object, &resp.Diagnostics)
	state.ID = types.StringValue(object.GetID())
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
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() || !validID(id, &resp.Diagnostics) {
		return
	}
	object := buildUpdateObject(ctx, plan.Section, &resp.Diagnostics)
	at := buildPosition(ctx, plan.At, &resp.Diagnostics)
	validatePosition(at, &resp.Diagnostics)
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
	plan.Section = r.updateObject(ctx, id.ValueString(), object, revision, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = id
	if at != nil && !plan.At.Equal(state.At) {
		plan.Section = r.move(ctx, id.ValueString(), at, revision, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	retained = true
	if !r.publishOrReconcile(ctx, revision, id.ValueString(), plan.Section, &resp.Diagnostics, at) {
		return
	}
	if resp.Private != nil {
		rememberRevision(ctx, resp.Private, "", &resp.Diagnostics)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
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
	// Dispose only this resource's unfinished draft before examining published state.
	if pending != "" {
		if !r.discard(ctx, pending, &resp.Diagnostics) {
			return
		}
		pending = ""
	}
	policy, err := r.readPolicy(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read section before deletion", err.Error())
		return
	}
	object := findObject(policy, id.ValueString())
	if object == nil {
		return
	}
	if !sectionEmpty(policy, id.ValueString(), &resp.Diagnostics) {
		return
	}
	revision := r.beginRevision(ctx, pending, &resp.Diagnostics, id.ValueString())
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
	if !r.removeObject(ctx, id.ValueString(), revision, &resp.Diagnostics) {
		return
	}
	retained = true
	if r.publishOrReconcile(ctx, revision, id.ValueString(), types.ObjectNull(sectionAttrTypes), &resp.Diagnostics) && resp.Private != nil {
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

func (r *policyResource) updateObject(ctx context.Context, id string, object *cato_models.PolicyUpdateSectionInfoInput,
	revision string, diags *diag.Diagnostics,
) types.Object {
	result, err := r.getClient().PolicyPrivateAccessUpdateSection(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyUpdateSectionInput{ID: id, Section: object})
	payload := result.GetPolicy().GetPrivateAccess().GetUpdateSection()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to update section", diags) {
		return types.ObjectNull(sectionAttrTypes)
	}
	if !checkStatus(payload.GetStatus(), diags) {
		return types.ObjectNull(sectionAttrTypes)
	}
	returned := payload.GetSection().GetSection()
	if returned.GetID() != id {
		diags.AddError("Invalid API response", "Update returned a different resource ID.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return parseUpdate(ctx, returned, diags)
}

func (r *policyResource) move(ctx context.Context, id string, at *cato_models.PolicySectionPositionInput,
	revision string, diags *diag.Diagnostics,
) types.Object {
	result, err := r.getClient().PolicyPrivateAccessMoveSection(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyMoveSectionInput{ID: id, To: at})
	payload := result.GetPolicy().GetPrivateAccess().GetMoveSection()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to move section", diags) || !checkStatus(payload.GetStatus(), diags) {
		return types.ObjectNull(sectionAttrTypes)
	}
	object := payload.GetSection().GetSection()
	if object.GetID() != id {
		diags.AddError("Invalid API response", "Move returned a different resource ID.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return parseMove(ctx, object, diags)
}

func mutationInput(revision string) *cato_models.PrivateAccessPolicyMutationInput {
	return &cato_models.PrivateAccessPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
}

func (r *policyResource) beginRevision(ctx context.Context, pending string, diags *diag.Diagnostics, emptySectionIDs ...string) string {
	if pending != "" && !r.discard(ctx, pending, diags) {
		return ""
	}
	result, err := r.getClient().PolicyPrivateAccessCreatePolicyRevision(ctx, r.client.AccountId, nil,
		cato_models.PolicyCreateRevisionInput{Name: "terraform-" + rand.Text(), Description: "Managed by Terraform"})
	payload := result.GetPolicy().GetPrivateAccess().GetCreatePolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to create policy revision", diags) || !checkStatus(payload.GetStatus(), diags) {
		return ""
	}
	revision := payload.GetPolicy().GetRevision().GetID()
	if revision == "" {
		diags.AddError("Invalid API response", "CreatePolicyRevision did not return a revision ID.")
		return ""
	}
	if len(emptySectionIDs) > 0 {
		if payload.GetPolicy().GetRules() == nil {
			diags.AddError("Invalid API response", "The policy revision did not return a complete rule list.")
			r.discard(ctx, revision, diags)
			return ""
		}
		for _, entry := range payload.GetPolicy().GetRules() {
			if entry.GetRule().GetID() == "" || entry.GetRule().GetSection().GetID() == emptySectionIDs[0] {
				diags.AddError("Section is not empty", "The revision contains rules in this section or has incomplete rule data.")
				r.discard(ctx, revision, diags)
				return ""
			}
		}
	}
	return revision
}

func (r *policyResource) publish(ctx context.Context, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicyPrivateAccessPublishPolicyRevision(
		ctx, r.client.AccountId, mutationInput(revision), &cato_models.PolicyPublishRevisionInput{},
	)
	payload := result.GetPolicy().GetPrivateAccess().GetPublishPolicyRevision()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to publish policy revision "+revision, diags) {
		return false
	}
	return checkStatus(payload.GetStatus(), diags)
}

func (r *policyResource) discard(ctx context.Context, revision string, diags *diag.Diagnostics) bool {
	result, err := r.getClient().PolicyPrivateAccessDiscardPolicyRevision(ctx, r.client.AccountId, mutationInput(revision),
		&cato_models.PolicyDiscardRevisionInput{ID: &revision})
	payload := result.GetPolicy().GetPrivateAccess().GetDiscardPolicyRevision()
	if revisionMissing(err) {
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
	return checkStatus(payload.GetStatus(), diags)
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
	result, err := r.getClient().PolicyPrivateAccessRemoveSection(ctx, r.client.AccountId, mutationInput(revision),
		cato_models.PolicyRemoveSectionInput{ID: id})
	payload := result.GetPolicy().GetPrivateAccess().GetRemoveSection()
	if utils.CheckAPIErrors(err, payload.GetErrors(), "Failed to delete section", diags) {
		return false
	}
	return checkStatus(payload.GetStatus(), diags)
}

func sectionEmpty(policy *policyResponse, id string, diags *diag.Diagnostics) bool {
	for _, entry := range policy.GetRules() {
		if entry.GetRule().GetSection().GetID() == id {
			diags.AddError("Section is not empty", "Remove or move all rules before deleting their section.")
			return false
		}
	}
	return true
}

// A publication can commit even when its response is lost. Reconcile only after
// the owned revision is confirmed gone, and only against published state.
func (r *policyResource) publishOrReconcile(
	ctx context.Context, revision, id string, expected types.Object, diags *diag.Diagnostics,
	positions ...*cato_models.PolicySectionPositionInput,
) bool {
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
				if !conversion.HasError() && actual.Equal(expected) && publicationPositionMatches(policy, id, positions) {
					return true
				}
			}
		}
	}
	diags.Append(publication...)
	return false
}

// Section identity does not encode ordering. Require published ordering evidence
// before acknowledging a move whose publication response was lost.
func publicationPositionMatches(policy *policyResponse, id string, positions []*cato_models.PolicySectionPositionInput) bool {
	if len(positions) == 0 || positions[0] == nil {
		return true
	}
	at := positions[0]
	index := publishedSectionIndex(policy, id)
	if index < 0 {
		return false
	}
	ref := ""
	if at.Ref != nil {
		ref = *at.Ref
	}
	refIndex := publishedSectionIndex(policy, ref)
	switch at.Position {
	case cato_models.PolicySectionPositionEnumLastInPolicy:
		// A nonempty ref denotes another scope; this projection cannot establish it.
		return ref == "" && index == len(policy.GetSections())-1
	case cato_models.PolicySectionPositionEnumBeforeSection:
		return refIndex >= 0 && index+1 == refIndex
	case cato_models.PolicySectionPositionEnumAfterSection:
		return refIndex >= 0 && index == refIndex+1
	default:
		return false
	}
}

func publishedSectionIndex(policy *policyResponse, id string) int {
	for index, item := range policy.GetSections() {
		if item.GetSection().GetID() == id {
			return index
		}
	}
	return -1
}
