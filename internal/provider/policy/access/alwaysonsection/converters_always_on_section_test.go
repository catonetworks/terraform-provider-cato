package alwaysonsection

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestBuildSectionPosition(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name         string
		at           types.Object
		invalid, ref bool
	}{
		{name: "last", at: sectionModel("", "", "LAST_IN_POLICY", "").At},
		{name: "before", at: sectionModel("", "", "BEFORE_SECTION", "other").At, ref: true},
		{name: "after", at: sectionModel("", "", "AFTER_SECTION", "other").At, ref: true},
		{name: "missing ref", at: sectionModel("", "", "BEFORE_SECTION", "").At, invalid: true},
		{name: "unsupported enum", at: sectionModel("", "", "FIRST_IN_POLICY", "").At, invalid: true},
		{name: "null", at: types.ObjectNull(policyrule.PositionAttrTypes), invalid: true},
		{name: "unknown", at: types.ObjectUnknown(policyrule.PositionAttrTypes), invalid: true},
		{name: "unknown ref", at: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("BEFORE_SECTION"), "ref": types.StringUnknown(),
		}), invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			input := buildSectionPosition(context.Background(), tc.at, &diags)
			require.Equal(t, tc.invalid, diags.HasError(), "%v", diags)
			if !tc.invalid {
				require.Equal(t, tc.ref, input.Ref != nil)
			}
		})
	}
}

func TestSectionInvalidInputStopsBeforeAPI(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"null ID", "unknown ID", "empty ID", "unknown section", "unknown at", "unconfigured"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			model := sectionModel("section-123", "Example", "LAST_IN_POLICY", "")
			stateModel := model
			client := mocks.NewAlwaysOnSectionClient(t)
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			switch mode {
			case "null ID", "unknown ID", "empty ID":
				attrs := stateModel.Section.Attributes()
				attrs["id"] = map[string]types.String{"null ID": types.StringNull(), "unknown ID": types.StringUnknown(), "empty ID": types.StringValue("")}[mode]
				stateModel.Section = types.ObjectValueMust(SectionDataAttrTypes, attrs)
			case "unknown section":
				model.Section = types.ObjectUnknown(SectionDataAttrTypes)
			case "unknown at":
				model.At = types.ObjectUnknown(policyrule.PositionAttrTypes)
			case "unconfigured":
				r.client, r.apiClient = nil, nil
			}
			state := sectionState(t, stateModel)
			update := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: sectionPlan(t, model), State: state}, &update)
			require.True(t, update.Diagnostics.HasError())
			require.True(t, state.Raw.Equal(update.State.Raw))
			if mode == "unknown section" || mode == "unknown at" {
				return
			}
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			require.True(t, read.Diagnostics.HasError())
			remove := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &remove)
			require.True(t, remove.Diagnostics.HasError())
		})
	}
}
