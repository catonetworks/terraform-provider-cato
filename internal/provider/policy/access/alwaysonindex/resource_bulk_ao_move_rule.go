// Package alwaysonindex implements cato_bulk_ao_move_rule: bulk ordering and
// publishing for the ZTNA Always-On policy. The policy is flat (sections +
// rules, no sub-policies) and the API has no reorderPolicy mutation, so
// ordering is applied as a deterministic sequence of MoveSection/MoveRule
// calls, followed by the Always-On stack's ONLY publish (the rule and section
// resources deliberately leave all changes in the draft).
package alwaysonindex

import (
	"context"
	"fmt"
	"sort"

	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

var (
	_ resource.Resource              = &aoRulesIndexResource{}
	_ resource.ResourceWithConfigure = &aoRulesIndexResource{}
)

func NewAoRulesIndexResource() resource.Resource {
	return &aoRulesIndexResource{}
}

type aoRulesIndexResource struct {
	client *catoclient.Service

	apiClient AoMoveResourceClient
}

// AoRulesIndex is the Terraform model.
type AoRulesIndex struct {
	SectionData types.Map `tfsdk:"section_data"`
	RuleData    types.Map `tfsdk:"rule_data"`
}

type AoSectionDataEntry struct {
	SectionIndex types.Int64 `tfsdk:"section_index"`
}

type AoRuleDataEntry struct {
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
}

var aoSectionDataAttrTypes = map[string]attr.Type{
	"section_index": types.Int64Type,
}

var aoRuleDataAttrTypes = map[string]attr.Type{
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
}

func (r *aoRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_ao_move_rule"
}

func (r *aoRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *aoRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_bulk_ao_move_rule` resource orders ZTNA Always-On sections and rules and PUBLISHES " +
			"the policy revision - it is the Always-On stack's only publisher (`cato_always_on_rule` and " +
			"`cato_always_on_section` leave their changes in the draft). Sections are ordered by `section_index`; " +
			"rules by `index_in_section` within their `section_name`. This resource is not importable; it is " +
			"created on the first apply.",
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

// aoMoveOp is one planned move operation (section or rule), in execution order.
type aoMoveOp struct {
	IsSection   bool
	Name        string
	SectionName string // for rules: the target section
}

// BuildAoMoveSequence turns the ordering inputs into a deterministic move
// sequence: sections in section_index order (each moved LAST_IN_POLICY, so
// the sequence yields ascending order), then per section (same order) its
// rules in index_in_section order (each moved LAST_IN_SECTION). Rules whose
// section is not in sectionIndexes are grouped after, in name order, so the
// sequence is total and deterministic.
func BuildAoMoveSequence(sectionIndexes map[string]int64, ruleRows map[string]AoRuleRow) []aoMoveOp {
	ops := make([]aoMoveOp, 0, len(sectionIndexes)+len(ruleRows))

	sectionNames := make([]string, 0, len(sectionIndexes))
	for name := range sectionIndexes {
		sectionNames = append(sectionNames, name)
	}
	sort.Slice(sectionNames, func(i, j int) bool {
		if sectionIndexes[sectionNames[i]] != sectionIndexes[sectionNames[j]] {
			return sectionIndexes[sectionNames[i]] < sectionIndexes[sectionNames[j]]
		}
		return sectionNames[i] < sectionNames[j]
	})
	for _, name := range sectionNames {
		ops = append(ops, aoMoveOp{IsSection: true, Name: name})
	}

	bySection := map[string][]string{}
	for ruleName := range ruleRows {
		bySection[ruleRows[ruleName].SectionName] = append(bySection[ruleRows[ruleName].SectionName], ruleName)
	}
	orderedSections := append([]string{}, sectionNames...)
	var extraSections []string
	for sectionName := range bySection {
		if _, ok := sectionIndexes[sectionName]; !ok {
			extraSections = append(extraSections, sectionName)
		}
	}
	sort.Strings(extraSections)
	orderedSections = append(orderedSections, extraSections...)

	for _, sectionName := range orderedSections {
		ruleNames := bySection[sectionName]
		sort.Slice(ruleNames, func(i, j int) bool {
			ri, rj := ruleRows[ruleNames[i]], ruleRows[ruleNames[j]]
			if ri.IndexInSection != rj.IndexInSection {
				return ri.IndexInSection < rj.IndexInSection
			}
			return ruleNames[i] < ruleNames[j]
		})
		for _, ruleName := range ruleNames {
			ops = append(ops, aoMoveOp{IsSection: false, Name: ruleName, SectionName: sectionName})
		}
	}
	return ops
}

// AoRuleRow is one rule ordering row.
type AoRuleRow struct {
	IndexInSection int64
	SectionName    string
}

func (r *aoRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AoRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API always-on bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *aoRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Ordering inputs are config-only (the API exposes order as rule/section
	// indexes, not as this resource's shape); keep state as-is.
	var state AoRulesIndex
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aoRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AoRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API always-on bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *aoRulesIndexResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Ordering resource: nothing to delete server-side.
	resp.State.RemoveResource(ctx)
}

//nolint:gocyclo
func (r *aoRulesIndexResource) applyOrderingAndPublish(ctx context.Context, plan AoRulesIndex) error {
	sectionIndexes, ruleRows, decodeErr := decodeAoOrderingInputs(ctx, plan)
	if decodeErr != nil {
		return decodeErr
	}

	if len(sectionIndexes) > 0 || len(ruleRows) > 0 {
		// Snapshot the policy to resolve names to ids.
		var body *aoPolicySnapshot
		readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyZtnaAlwaysOnPolicy bulk-move snapshot", func() error {
			var callErr error
			body, callErr = r.snapshot(ctx)
			return callErr
		})
		if readErr != nil {
			return readErr
		}

		for _, op := range BuildAoMoveSequence(sectionIndexes, ruleRows) {
			if op.IsSection {
				sectionID, ok := body.SectionIDByName[op.Name]
				if !ok {
					return fmt.Errorf("section %q from section_data not found in the policy", op.Name)
				}
				if err := r.moveSectionLast(ctx, sectionID); err != nil {
					return err
				}
				continue
			}
			ruleID, ok := body.RuleIDByName[op.Name]
			if !ok {
				return fmt.Errorf("rule %q from rule_data not found in the policy", op.Name)
			}
			sectionID, ok := body.SectionIDByName[op.SectionName]
			if !ok {
				return fmt.Errorf("rule %q references section %q which does not exist in the policy", op.Name, op.SectionName)
			}
			if err := r.moveRuleLastInSection(ctx, ruleID, sectionID); err != nil {
				return err
			}
		}
	}

	// The Always-On stack's only publish.
	return sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicyZtnaAlwaysOnPublishPolicyRevision", func() error {
		pub, err := r.resourceAPI().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
		if err != nil {
			return err
		}
		if errs := pub.GetPolicy().GetZtnaAlwaysOn().GetPublishPolicyRevision().GetErrors(); len(errs) > 0 {
			return fmt.Errorf("always-on publishPolicyRevision errors: %s", formatPolicyMutationErrors(errs))
		}
		return nil
	})
}

func (r *aoRulesIndexResource) moveSectionLast(ctx context.Context, sectionID string) error {
	moveResp, err := r.resourceAPI().PolicyZtnaAlwaysOnMoveSection(ctx, cato_models.PolicyMoveSectionInput{
		ID: sectionID,
		To: &cato_models.PolicySectionPositionInput{
			Position: cato_models.PolicySectionPositionEnumLastInPolicy,
		},
	}, r.client.AccountId)
	if err != nil {
		return err
	}
	ms := moveResp.GetPolicy().GetZtnaAlwaysOn().GetMoveSection()
	if errs := ms.GetErrors(); len(errs) > 0 {
		return fmt.Errorf("always-on moveSection errors: %s", formatPolicyMutationErrors(errs))
	}
	if ms.Status != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("always-on moveSection status %q", string(ms.Status))
	}
	return nil
}

func (r *aoRulesIndexResource) moveRuleLastInSection(ctx context.Context, ruleID, sectionID string) error {
	pos := cato_models.PolicyRulePositionEnumLastInSection
	moveResp, err := r.resourceAPI().PolicyZtnaAlwaysOnMoveRule(ctx, cato_models.PolicyMoveRuleInput{
		ID: ruleID,
		To: &cato_models.PolicyRulePositionInput{
			Position: &pos,
			Ref:      &sectionID,
		},
	}, r.client.AccountId)
	if err != nil {
		return err
	}
	mr := moveResp.GetPolicy().GetZtnaAlwaysOn().GetMoveRule()
	if errs := mr.GetErrors(); len(errs) > 0 {
		return fmt.Errorf("always-on moveRule errors: %s", formatPolicyMutationErrors(errs))
	}
	if mr.Status != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("always-on moveRule status %q", string(mr.Status))
	}
	tflog.Debug(ctx, "always-on moveRule applied", map[string]any{"rule_id": ruleID, "section_id": sectionID})
	return nil
}

// aoPolicySnapshot resolves names to ids from the live policy.
type aoPolicySnapshot struct {
	SectionIDByName map[string]string
	RuleIDByName    map[string]string
}

func (r *aoRulesIndexResource) snapshot(ctx context.Context) (*aoPolicySnapshot, error) {
	body, err := r.resourceAPI().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		return nil, err
	}
	snap := &aoPolicySnapshot{
		SectionIDByName: map[string]string{},
		RuleIDByName:    map[string]string{},
	}
	pol := body.GetPolicy().GetZtnaAlwaysOn().GetPolicy()
	for _, s := range pol.GetSections() {
		snap.SectionIDByName[s.GetSection().GetName()] = s.GetSection().ID
	}
	for _, rw := range pol.GetRules() {
		snap.RuleIDByName[rw.GetRule().GetName()] = rw.GetRule().GetID()
	}
	return snap, nil
}

type policyMutationErrorLike interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

func formatPolicyMutationErrors[T policyMutationErrorLike](errs []T) string {
	out := ""
	for _, e := range errs {
		if out != "" {
			out += "; "
		}
		code, msg := "", ""
		if e.GetErrorCode() != nil {
			code = *e.GetErrorCode()
		}
		if e.GetErrorMessage() != nil {
			msg = *e.GetErrorMessage()
		}
		out += fmt.Sprintf("%s : %s", code, msg)
	}
	return out
}

var _ = basetypes.ObjectAsOptions{}
var _ = aoSectionDataAttrTypes
var _ = aoRuleDataAttrTypes
