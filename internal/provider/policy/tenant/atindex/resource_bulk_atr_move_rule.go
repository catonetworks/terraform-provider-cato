// Package atindex implements cato_bulk_atr_move_rule: bulk ordering and
// publishing for the App Tenant Restriction policy. The policy is flat (sections +
// rules, no sub-policies) and the API has no reorderPolicy mutation, so
// ordering is applied as a deterministic sequence of MoveSection/MoveRule
// calls, followed by a policy revision publish.
package atindex

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

var (
	_ resource.Resource              = &atrRulesIndexResource{}
	_ resource.ResourceWithConfigure = &atrRulesIndexResource{}
)

func NewAtrRulesIndexResource() resource.Resource {
	return &atrRulesIndexResource{}
}

type atrRulesIndexResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

// AtrRulesIndex is the Terraform model.
type AtrRulesIndex struct {
	SectionData types.Map `tfsdk:"section_data"`
	RuleData    types.Map `tfsdk:"rule_data"`
}

type AtrRulesIndexSectionEntry struct {
	SectionIndex types.Int64 `tfsdk:"section_index"`
}

type AtrRulesIndexRuleEntry struct {
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
}

func (r *atrRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_atr_move_rule"
}

func (r *atrRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *atrRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_bulk_atr_move_rule` resource orders App Tenant Restriction sections and rules and publishes " +
			"the policy revision. Sections are ordered by `section_index`; rules by `index_in_section` within " +
			"their `section_name`. This resource is not importable; it is created on the first apply.",
		Attributes: map[string]schema.Attribute{
			"section_data": schema.MapNestedAttribute{
				Description: "Map of section name to its ordering data.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"section_index": schema.Int64Attribute{
							Description: "1-based position of the section in the policy.",
							Required:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of rule name to its ordering data.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index_in_section": schema.Int64Attribute{
							Description: "1-based position of the rule within its section.",
							Required:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "Name of the section the rule is ordered within.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *atrRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AtrRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API App Tenant Restriction bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *atrRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Ordering inputs are config-only; keep state as-is.
	var state AtrRulesIndex
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *atrRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AtrRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API App Tenant Restriction bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *atrRulesIndexResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

//nolint:gocyclo
func (r *atrRulesIndexResource) applyOrderingAndPublish(ctx context.Context, plan AtrRulesIndex) error {
	sectionIndexes, ruleRows, decodeErr := decodeAtrOrderingInputs(ctx, plan)
	if decodeErr != nil {
		return decodeErr
	}

	if len(sectionIndexes) > 0 || len(ruleRows) > 0 {
		sectionIDByName := map[string]string{}
		ruleIDByName := map[string]string{}
		readErr := sharedpolicy.WithTransientReadRetry(ctx, "AppTenantRestrictionPolicy bulk-move snapshot", func() error {
			body, callErr := r.resourceAPI().AppTenantRestrictionPolicy(ctx, r.client.AccountId)
			if callErr != nil {
				return callErr
			}
			pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
			for _, s := range pol.GetSections() {
				sectionIDByName[s.GetSection().GetName()] = s.GetSection().ID
			}
			for _, rw := range pol.GetRules() {
				ruleIDByName[rw.GetRule().GetName()] = rw.GetRule().GetID()
			}
			return nil
		})
		if readErr != nil {
			return readErr
		}

		for _, op := range sharedpolicy.BuildFlatMoveSequence(sectionIndexes, ruleRows) {
			if op.IsSection {
				sectionID, ok := sectionIDByName[op.Name]
				if !ok {
					return fmt.Errorf("section %q from section_data not found in the policy", op.Name)
				}
				moveResp, err := r.resourceAPI().PolicyAppTenantRestrictionMoveSection(ctx, cato_models.PolicyMoveSectionInput{
					ID: sectionID,
					To: &cato_models.PolicySectionPositionInput{
						Position: cato_models.PolicySectionPositionEnumLastInPolicy,
					},
				}, r.client.AccountId)
				if err != nil {
					return err
				}
				ms := moveResp.GetPolicy().GetAppTenantRestriction().GetMoveSection()
				if err := sharedpolicy.MutationResult(ms.GetStatus(), ms.GetErrors()); err != nil {
					return err
				}
				continue
			}
			ruleID, ok := ruleIDByName[op.Name]
			if !ok {
				return fmt.Errorf("rule %q from rule_data not found in the policy", op.Name)
			}
			sectionID, ok := sectionIDByName[op.SectionName]
			if !ok {
				return fmt.Errorf("rule %q references section %q which does not exist in the policy", op.Name, op.SectionName)
			}
			pos := cato_models.PolicyRulePositionEnumLastInSection
			moveResp, err := r.resourceAPI().PolicyAppTenantRestrictionMoveRule(ctx, cato_models.PolicyMoveRuleInput{
				ID: ruleID,
				To: &cato_models.PolicyRulePositionInput{
					Position: &pos,
					Ref:      &sectionID,
				},
			}, r.client.AccountId)
			if err != nil {
				return err
			}
			mr := moveResp.GetPolicy().GetAppTenantRestriction().GetMoveRule()
			if err := sharedpolicy.MutationResult(mr.GetStatus(), mr.GetErrors()); err != nil {
				return err
			}
		}
	}

	return sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicyAppTenantRestrictionPublishPolicyRevision", func() error {
		pub, err := r.resourceAPI().PolicyAppTenantRestrictionPublishPolicyRevision(ctx, r.client.AccountId)
		if err != nil {
			return err
		}
		payload := pub.GetPolicy().GetAppTenantRestriction().GetPublishPolicyRevision()
		return sharedpolicy.PublishResult(payload.GetStatus(), payload.GetErrors())
	})
}
