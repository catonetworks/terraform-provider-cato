package idname

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

const refErrSummary = "Configuration error: ID/Name reference"

// SetPlanModifier returns a plan modifier for a set of ID/Name reference objects
func SetPlanModifier() planmodifier.Set {
	return idNameSetPlanModifier{}
}

// idNameSetPlanModifier implements the plan modifier.
type idNameSetPlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m idNameSetPlanModifier) Description(_ context.Context) string {
	return "Keep ID/Name from state, if one is changed, set the other to unknown"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m idNameSetPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifySet implements the plan modification logic.
//
//nolint:gocyclo,funlen
func (m idNameSetPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Preserve resource destruction and unresolved configuration plans.
	if req.Plan.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var cfg, state []*Ref
	var plan []Ref

	planSetEmpty, setDiag := types.SetValueFrom(ctx, RefObjectType, []Ref{})
	if setDiag.HasError() {
		resp.Diagnostics.Append(setDiag...)
		return
	}

	if req.ConfigValue.IsNull() {
		resp.PlanValue = planSetEmpty
		return
	}
	// Whole objects can be unknown even when the containing set is known.
	for _, element := range req.ConfigValue.Elements() {
		if element.IsUnknown() {
			return // TF config not ready yet
		}
	}
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.ElementsAs(ctx, &cfg, false)) {
		return
	}

	if len(cfg) == 0 {
		resp.PlanValue = planSetEmpty
		return
	}

	// get current state ID/Name maps, if available
	var names, ids map[string]string
	if utils.HasValue(req.StateValue) {
		if utils.CheckErr(&resp.Diagnostics, req.StateValue.ElementsAs(ctx, &state, false)) {
			return
		}
		names, ids = m.makeNameIDMaps(state, req.Path.String(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	seen := make(map[Ref]struct{})
	for _, cfgItem := range cfg {
		if cfgItem == nil {
			continue
		}
		if cfgItem.Name.IsUnknown() || cfgItem.ID.IsUnknown() {
			return // TF config is not ready yet
		}
		switch {
		case !cfgItem.Name.IsNull():
			planItem := m.handleName(cfgItem, names, req.Path.String(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			if !m.checkDuplicate(planItem, seen, req.Path.String(), &resp.Diagnostics) {
				return
			}
			plan = append(plan, planItem)
		case !cfgItem.ID.IsNull():
			planItem := m.handleID(cfgItem, ids, req.Path.String(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			if !m.checkDuplicate(planItem, seen, req.Path.String(), &resp.Diagnostics) {
				return
			}
			plan = append(plan, planItem)
		default:
			resp.Diagnostics.AddError(refErrSummary, "'name' or 'id' must be defined in the config "+req.Path.String())
			return
		}
	}
	planSet, setDiag := types.SetValueFrom(ctx, RefObjectType, plan)
	resp.Diagnostics.Append(setDiag...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = planSet
}
func (m idNameSetPlanModifier) checkDuplicate(ref Ref, seen map[Ref]struct{}, path string, diags *diag.Diagnostics) bool {
	if _, ok := seen[ref]; ok {
		diags.AddError(refErrSummary, fmt.Sprintf("Duplicate entry in the config path=%s, id=%s, name=%s",
			path, ref.ID.ValueString(), ref.Name.ValueString()))
		return false
	}
	seen[ref] = struct{}{}
	return true
}

// handleName returns a Ref object containing Name and optionally ID if we have it in the state already
func (m idNameSetPlanModifier) handleName(cfgItem *Ref, nameIDMap map[string]string,
	path string, diags *diag.Diagnostics,
) Ref {
	if cfgItem == nil {
		diags.AddError("internal error", "set item is nil"+path)
		return Ref{}
	}
	name := cfgItem.Name.ValueString()
	if name == "" {
		diags.AddError(refErrSummary, "'name' must not be empty string in the config "+path)
		return Ref{}
	}
	if utils.HasValue(cfgItem.ID) {
		diags.AddError(refErrSummary,
			fmt.Sprintf("only one of 'name' or 'id' can be specified in the config, [id:%q, name:%q]",
				cfgItem.ID.ValueString(), name))
	}
	ref := Ref{
		ID:   types.StringUnknown(),
		Name: types.StringValue(name),
	}
	if id, ok := nameIDMap[name]; ok {
		ref.ID = types.StringValue(id)
	}
	return ref
}

// handleID returns a Ref object containing ID and optionally Name if we have it in the state already
func (m idNameSetPlanModifier) handleID(cfgItem *Ref, idNameMap map[string]string,
	path string, diags *diag.Diagnostics,
) Ref {
	if cfgItem == nil {
		diags.AddError("internal error", "set item is nil"+path)
		return Ref{}
	}
	id := cfgItem.ID.ValueString()
	if id == "" {
		diags.AddError(refErrSummary, "'id' must not be empty string in the config "+path)
		return Ref{}
	}
	if utils.HasValue(cfgItem.Name) {
		diags.AddError(refErrSummary,
			fmt.Sprintf("only one of 'id' or 'name' can be specified in the config, [name:%q, id:%q]",
				cfgItem.Name.ValueString(), id))
	}
	ref := Ref{
		Name: types.StringUnknown(),
		ID:   types.StringValue(id),
	}
	if name, ok := idNameMap[id]; ok {
		ref.Name = types.StringValue(name)
	}
	return ref
}

func (m idNameSetPlanModifier) makeNameIDMaps(state []*Ref, path string, diags *diag.Diagnostics,
) (names, ids map[string]string) {
	names = make(map[string]string)
	ids = make(map[string]string)
	for _, item := range state {
		if item == nil || !utils.HasValue(item.Name) || !utils.HasValue(item.ID) {
			continue
		}
		name, id := item.Name.ValueString(), item.ID.ValueString()
		if _, ok := names[name]; ok {
			diags.AddError(refErrSummary, fmt.Sprintf("Duplicate name '%s' in the state %s", name, path))
			return
		}
		names[name] = id
		if _, ok := ids[id]; ok {
			diags.AddError(refErrSummary, fmt.Sprintf("Duplicate id '%s' in the state %s", id, path))
			return
		}
		ids[id] = name
	}
	return names, ids
}
