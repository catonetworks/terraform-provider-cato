package roles

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func queryInput(ctx context.Context, config Model) (m.RoleManagementRoleListInput, diag.Diagnostics) {
	var d diag.Diagnostics
	direction := config.SortDirection.ValueString()
	if config.SortDirection.IsNull() || config.SortDirection.IsUnknown() {
		direction = "ASC"
	}
	if direction != "ASC" && direction != "DESC" {
		d.AddError("Invalid sort direction", "Use ASC or DESC.")
	}
	input := m.RoleManagementRoleListInput{Sort: &m.RoleManagementRoleSortInput{Name: &m.SortOrderInput{Direction: m.SortOrder(direction),
		Priority: 1}}}
	if config.Filter.IsUnknown() {
		d.AddError("Invalid role filter", "Filter must be known.")
		return input, d
	}
	if config.Filter.IsNull() {
		return input, d
	}
	a := config.Filter.Attributes()
	input.Filter = &m.RoleManagementRoleFilterInput{}
	for _, name := range []string{"id", "name"} {
		object := a[name].(types.Object)
		if object.IsUnknown() {
			d.AddError("Invalid role filter", "Filter fields must be known.")
			continue
		}
		if object.IsNull() {
			continue
		}
		eq, neq, in, nin, diags := stringFilter(ctx, object)
		d.Append(diags...)
		if name == "id" {
			input.Filter.ID = &m.IDFilterInput{Eq: eq, Neq: neq, In: in, Nin: nin}
		} else {
			input.Filter.Name = &m.StringFilterInput{Eq: eq, Neq: neq, In: in, Nin: nin}
		}
	}
	predefined, predefinedDiags := predefinedInput(a["predefined"].(types.Object))
	input.Filter.Predefined = predefined
	d.Append(predefinedDiags...)
	return input, d
}
func stringFilter(ctx context.Context, object types.Object) (eq, neq *string, in, nin []string, d diag.Diagnostics) {
	for name, value := range object.Attributes() {
		switch v := value.(type) {
		case types.String:
			if v.IsUnknown() {
				d.AddError("Invalid filter", "Filter strings must be known.")
				continue
			}
			if v.IsNull() {
				continue
			}
			s := v.ValueString()
			if name == "eq" {
				eq = &s
			} else {
				neq = &s
			}
		case types.Set:
			if v.IsUnknown() {
				d.AddError("Invalid filter", "Filter sets must be known.")
				continue
			}
			if v.IsNull() {
				continue
			}
			var values []types.String
			d.Append(v.ElementsAs(ctx, &values, false)...)
			output := make([]string, 0, len(values))
			for _, item := range values {
				if item.IsNull() || item.IsUnknown() {
					d.AddError("Invalid filter", "Filter entries must be known strings.")
					continue
				}
				output = append(output, item.ValueString())
			}
			if name == "in" {
				in = output
			} else {
				nin = output
			}
		}
	}
	return eq, neq, in, nin, d
}

func predefinedInput(predefined types.Object) (*m.BooleanFilterInput, diag.Diagnostics) {
	var result *m.BooleanFilterInput
	var d diag.Diagnostics
	if predefined.IsUnknown() {
		d.AddError("Invalid role filter", "Predefined filter must be known.")
	} else if !predefined.IsNull() {
		a := predefined.Attributes()
		result = &m.BooleanFilterInput{}
		for _, name := range []string{"eq", "neq"} {
			v := a[name].(types.Bool)
			if v.IsUnknown() {
				d.AddError("Invalid role filter", "Boolean filters must be known.")
				continue
			}
			if v.IsNull() {
				continue
			}
			b := v.ValueBool()
			if name == "eq" {
				result.Eq = &b
			} else {
				result.Neq = &b
			}
		}
	}

	return result, d
}
