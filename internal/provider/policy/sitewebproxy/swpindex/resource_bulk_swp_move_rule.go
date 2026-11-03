// Package swpindex implements cato_bulk_swp_move_rule: bulk ordering and
// publishing for the Site Web Proxy policy. The policy is flat (sections +
// rules, no sub-policies) and the API has no reorderPolicy mutation, so
// ordering is applied as a deterministic sequence of MoveSection/MoveRule
// calls, followed by a policy revision publish.
package swpindex

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
	_ resource.Resource              = &swpRulesIndexResource{}
	_ resource.ResourceWithConfigure = &swpRulesIndexResource{}
)

func NewSwpRulesIndexResource() resource.Resource {
	return &swpRulesIndexResource{}
}

type swpRulesIndexResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

// SwpRulesIndex is the Terraform model.
type SwpRulesIndex struct {
	SectionData types.Map `tfsdk:"section_data"`
	RuleData    types.Map `tfsdk:"rule_data"`
}

type SwpRulesIndexSectionEntry struct {
	SectionIndex types.Int64 `tfsdk:"section_index"`
}

type SwpRulesIndexRuleEntry struct {
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
}

func (r *swpRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_swp_move_rule"
}

func (r *swpRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *swpRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_bulk_swp_move_rule` resource orders Site Web Proxy sections and rules and publishes " +
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

func (r *swpRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SwpRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API Site Web Proxy bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *swpRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Ordering inputs are config-only; keep state as-is.
	var state SwpRulesIndex
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *swpRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SwpRulesIndex
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyOrderingAndPublish(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Catov2 API Site Web Proxy bulk move error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *swpRulesIndexResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

//nolint:gocyclo
func (r *swpRulesIndexResource) applyOrderingAndPublish(ctx context.Context, plan SwpRulesIndex) error {
	sectionIndexes, ruleRows, decodeErr := decodeSwpOrderingInputs(ctx, plan)
	if decodeErr != nil {
		return decodeErr
	}

	if len(sectionIndexes) > 0 || len(ruleRows) > 0 {
		sectionIDByName := map[string]string{}
		ruleIDByName := map[string]string{}
		readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicySiteWebProxyPolicy bulk-move snapshot", func() error {
			body, callErr := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
			if callErr != nil {
				return callErr
			}
			pol := body.GetPolicy().GetSiteWebProxy().GetPolicy()
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
				moveResp, err := r.resourceAPI().PolicySiteWebProxyMoveSection(ctx, cato_models.PolicyMoveSectionInput{
					ID: sectionID,
					To: &cato_models.PolicySectionPositionInput{
						Position: cato_models.PolicySectionPositionEnumLastInPolicy,
					},
				}, r.client.AccountId)
				if err != nil {
					return err
				}
				ms := moveResp.GetPolicy().GetSiteWebProxy().GetMoveSection()
				if st := ms.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
					return fmt.Errorf("site web proxy moveSection status %q", string(*st))
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
			moveResp, err := r.resourceAPI().PolicySiteWebProxyMoveRule(ctx, cato_models.PolicyMoveRuleInput{
				ID: ruleID,
				To: &cato_models.PolicyRulePositionInput{
					Position: &pos,
					Ref:      &sectionID,
				},
			}, r.client.AccountId)
			if err != nil {
				return err
			}
			mr := moveResp.GetPolicy().GetSiteWebProxy().GetMoveRule()
			if st := mr.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
				return fmt.Errorf("site web proxy moveRule status %q", string(*st))
			}
		}
	}

	return sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySiteWebProxyPublishPolicyRevision", func() error {
		pub, err := r.resourceAPI().PolicySiteWebProxyPublishPolicyRevision(ctx, r.client.AccountId)
		if err != nil {
			return err
		}
		if errs := pub.GetPolicy().GetSiteWebProxy().GetPublishPolicyRevision().GetErrors(); len(errs) > 0 {
			msg := ""
			for _, e := range errs {
				if msg != "" {
					msg += "; "
				}
				code, m := "", ""
				if e.GetErrorCode() != nil {
					code = *e.GetErrorCode()
				}
				if e.GetErrorMessage() != nil {
					m = *e.GetErrorMessage()
				}
				msg += fmt.Sprintf("%s : %s", code, m)
			}
			return fmt.Errorf("site web proxy publishPolicyRevision errors: %s", msg)
		}
		return nil
	})
}
