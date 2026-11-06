package convert

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveEmptyValues converts a hydrated model to an object, retaining explicitly
// empty collections and objects from the prior plan/state when the API omits them.
// Populated prior values and unknown values never replace the API response.
func PreserveEmptyValues(
	ctx context.Context,
	attrTypes map[string]attr.Type,
	hydrated, prior any,
) (types.Object, diag.Diagnostics) {
	current, diags := types.ObjectValueFrom(ctx, attrTypes, hydrated)
	previous, d := types.ObjectValueFrom(ctx, attrTypes, prior)
	diags.Append(d...)
	if diags.HasError() {
		return current, diags
	}
	value, d := preserveEmptyValue(ctx, current, previous)
	diags.Append(d...)
	return value.(types.Object), diags
}

func preserveEmptyValue(ctx context.Context, current, prior attr.Value) (attr.Value, diag.Diagnostics) {
	if current.IsNull() && !prior.IsNull() && isEmptyValue(prior) {
		return prior, nil
	}
	currentObj, currentIsObject := current.(types.Object)
	priorObj, priorIsObject := prior.(types.Object)
	if !currentIsObject || !priorIsObject || current.IsNull() || current.IsUnknown() || prior.IsNull() || prior.IsUnknown() {
		return current, nil
	}
	var diags diag.Diagnostics
	attrs := currentObj.Attributes()
	priorAttrs := priorObj.Attributes()
	for name, value := range attrs {
		previous, ok := priorAttrs[name]
		if !ok {
			continue
		}
		merged, d := preserveEmptyValue(ctx, value, previous)
		diags.Append(d...)
		attrs[name] = merged
	}
	object, d := types.ObjectValue(currentObj.AttributeTypes(ctx), attrs)
	diags.Append(d...)
	return object, diags
}

func isEmptyValue(value attr.Value) bool {
	if value.IsUnknown() {
		return false
	}
	if value.IsNull() {
		return true
	}
	switch v := value.(type) {
	case types.List:
		return len(v.Elements()) == 0
	case types.Set:
		return len(v.Elements()) == 0
	case types.Object:
		for _, child := range v.Attributes() {
			if !isEmptyValue(child) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
