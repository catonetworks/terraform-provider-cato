package idname

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// ListPlanModifier preserves ID/Name references in configuration order, discarding null elements.
func ListPlanModifier() planmodifier.List {
	return idNameListPlanModifier{}
}

// idNameListPlanModifier implements the plan modifier.
type idNameListPlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m idNameListPlanModifier) Description(_ context.Context) string {
	return "Preserves matching IDs and names and discards null references."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m idNameListPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyList implements the plan modification logic.
//
//nolint:gocyclo,funlen
func (m idNameListPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Preserve resource destruction and unresolved configuration plans.
	if req.Plan.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var cfg, state []*Ref

	planListEmpty, setDiag := types.ListValueFrom(ctx, RefObjectType, []Ref{})
	if setDiag.HasError() {
		resp.Diagnostics.Append(setDiag...)
		return
	}

	if req.ConfigValue.IsNull() {
		resp.PlanValue = planListEmpty
		return
	}
	// Whole objects can be unknown even when the containing list is known.
	for _, element := range req.ConfigValue.Elements() {
		if element.IsUnknown() {
			return // TF config not ready yet
		}
	}
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.ElementsAs(ctx, &cfg, false)) {
		return
	}
	cfg = slices.DeleteFunc(cfg, func(item *Ref) bool { return item == nil })

	if len(cfg) == 0 {
		resp.PlanValue = planListEmpty
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

	plan := make([]Ref, 0, len(cfg))
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
			plan = append(plan, planItem)
		case !cfgItem.ID.IsNull():
			planItem := m.handleID(cfgItem, ids, req.Path.String(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			plan = append(plan, planItem)
		default:
			resp.Diagnostics.AddError(refErrSummary, "'name' or 'id' must be defined in the config "+req.Path.String())
			return
		}
	}
	planList, setDiag := types.ListValueFrom(ctx, RefObjectType, plan)
	resp.Diagnostics.Append(setDiag...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = planList
}

// handleName returns a Ref object containing Name and optionally ID if we have it in the state already
func (m idNameListPlanModifier) handleName(cfgItem *Ref, nameIDMap map[string]string,
	path string, diags *diag.Diagnostics,
) Ref {
	if cfgItem == nil {
		diags.AddError("internal error", "list item is nil "+path)
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
func (m idNameListPlanModifier) handleID(cfgItem *Ref, idNameMap map[string]string,
	path string, diags *diag.Diagnostics,
) Ref {
	if cfgItem == nil {
		diags.AddError("internal error", "list item is nil "+path)
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

func (m idNameListPlanModifier) makeNameIDMaps(state []*Ref, path string, diags *diag.Diagnostics,
) (names, ids map[string]string) {
	names = make(map[string]string)
	ids = make(map[string]string)
	for _, item := range state {
		if item == nil || !utils.HasValue(item.Name) || !utils.HasValue(item.ID) {
			continue
		}
		name, id := item.Name.ValueString(), item.ID.ValueString()
		if previousID, ok := names[name]; ok && previousID != id {
			diags.AddError(refErrSummary, fmt.Sprintf("Duplicate name '%s' in the state %s", name, path))
			return
		}
		names[name] = id
		if previousName, ok := ids[id]; ok && previousName != name {
			diags.AddError(refErrSummary, fmt.Sprintf("Duplicate id '%s' in the state %s", id, path))
			return
		}
		ids[id] = name
	}
	return names, ids
}
