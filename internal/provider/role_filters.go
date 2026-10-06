package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var roleStringFilterAttrTypes = map[string]attr.Type{
	"eq":  types.StringType,
	"neq": types.StringType,
	"in": types.SetType{
		ElemType: types.StringType,
	},
	"nin": types.SetType{
		ElemType: types.StringType,
	},
}
var roleBooleanFilterAttrTypes = map[string]attr.Type{"eq": types.BoolType, "neq": types.BoolType}
var roleFilterAttrTypes = map[string]attr.Type{
	"id": types.ObjectType{
		AttrTypes: roleStringFilterAttrTypes,
	},
	"name": types.ObjectType{
		AttrTypes: roleStringFilterAttrTypes,
	},
	"predefined": types.ObjectType{
		AttrTypes: roleBooleanFilterAttrTypes,
	},
}

func roleStringFilter(ctx context.Context, object types.Object) (*cato_models.StringFilterInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	if object.IsNull() {
		return nil, diags
	}
	if object.IsUnknown() {
		diags.AddError("Unknown role filter", "Role filters must be known before reading roles.")
		return nil, diags
	}
	attrs := object.Attributes()
	input := &cato_models.StringFilterInput{}
	for _, key := range []string{"eq", "neq"} {
		value := attrs[key].(types.String)
		if value.IsUnknown() {
			diags.AddError("Unknown role filter", "Role filter values must be known.")
			continue
		}
		if value.IsNull() {
			continue
		}
		if key == "eq" {
			input.Eq = value.ValueStringPointer()
		} else {
			input.Neq = value.ValueStringPointer()
		}
	}
	for _, key := range []string{"in", "nin"} {
		value := attrs[key].(types.Set)
		if value.IsUnknown() {
			diags.AddError("Unknown role filter", "Role filter sets must be known.")
			continue
		}
		if value.IsNull() {
			continue
		}
		var values []string
		diags.Append(value.ElementsAs(ctx, &values, false)...)
		if key == "in" {
			input.In = values
		} else {
			input.Nin = values
		}
	}
	return input, diags
}

func roleListInput(ctx context.Context, filter types.Object, direction types.String) (
	cato_models.RoleManagementRoleListInput, diag.Diagnostics,
) {
	var diags diag.Diagnostics
	order := cato_models.SortOrder(direction.ValueString())
	if direction.IsNull() {
		order = cato_models.SortOrderAsc
	}
	if direction.IsUnknown() || (order != cato_models.SortOrderAsc && order != cato_models.SortOrderDesc) {
		diags.AddError("Invalid role sort direction", "sort_direction must be ASC or DESC.")
	}
	input := cato_models.RoleManagementRoleListInput{
		Sort: &cato_models.RoleManagementRoleSortInput{
			Name: &cato_models.SortOrderInput{
				Direction: order,
			},
		},
		Paging: &cato_models.PagingInput{
			From:  0,
			Limit: rolePageSize,
		},
	}
	if filter.IsNull() {
		return input, diags
	}
	if filter.IsUnknown() {
		diags.AddError("Unknown role filter", "filter must be known before reading roles.")
		return input, diags
	}
	attrs := filter.Attributes()
	input.Filter = &cato_models.RoleManagementRoleFilterInput{}
	id, d := roleStringFilter(ctx, attrs["id"].(types.Object))
	diags.Append(d...)
	if id != nil {
		input.Filter.ID = &cato_models.IDFilterInput{Eq: id.Eq, Neq: id.Neq, In: id.In, Nin: id.Nin}
	}
	input.Filter.Name, d = roleStringFilter(ctx, attrs["name"].(types.Object))
	diags.Append(d...)
	predefined := attrs["predefined"].(types.Object)
	if predefined.IsUnknown() {
		diags.AddError("Unknown role filter", "predefined filter must be known.")
	} else if !predefined.IsNull() {
		values := predefined.Attributes()
		input.Filter.Predefined = &cato_models.BooleanFilterInput{}
		for _, key := range []string{"eq", "neq"} {
			value := values[key].(types.Bool)
			if value.IsUnknown() {
				diags.AddError("Unknown role filter", "predefined filter values must be known.")
				continue
			}
			if value.IsNull() {
				continue
			}
			if key == "eq" {
				input.Filter.Predefined.Eq = value.ValueBoolPointer()
			} else {
				input.Filter.Predefined.Neq = value.ValueBoolPointer()
			}
		}
	}
	return input, diags
}
