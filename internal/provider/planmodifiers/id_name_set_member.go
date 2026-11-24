package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// IDNameSetMemberPlanModifier is the element-level companion of
// IDNameSetPlanModifier. The framework hands a set element's modifier the
// state value at the same POSITION, which for a set can be a different
// reference; IDNamePlanModifier then marks the computed id/name unknown and
// overwrites what the set-level modifier already resolved by matching names
// across the whole set. This keeps an element the set-level modifier resolved
// (its configured selector matches and its computed counterpart is known) and
// otherwise delegates to IDNamePlanModifier, which also validates the config.
func IDNameSetMemberPlanModifier() planmodifier.Object {
	return idNameSetMemberModifier{inner: idNamePlanModifier{}}
}

type idNameSetMemberModifier struct{ inner idNamePlanModifier }

func (m idNameSetMemberModifier) Description(ctx context.Context) string {
	return "keeps a set member resolved by IDNameSetPlanModifier; otherwise behaves like IDNamePlanModifier"
}

func (m idNameSetMemberModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m idNameSetMemberModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.PlanValue.IsNull() && !req.PlanValue.IsUnknown() && !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		p, c := req.PlanValue.Attributes(), req.ConfigValue.Attributes()
		pn, pi, cn, ci := p["name"], p["id"], c["name"], c["id"]
		resolved := pn != nil && pi != nil && !pn.IsNull() && !pn.IsUnknown() && !pi.IsNull() && !pi.IsUnknown()
		exactlyOne := cn != nil && ci != nil && (cn.IsNull() != ci.IsNull())
		if resolved && exactlyOne && ((!cn.IsNull() && cn.Equal(pn)) || (!ci.IsNull() && ci.Equal(pi))) {
			return
		}
	}
	m.inner.PlanModifyObject(ctx, req, resp)
}
