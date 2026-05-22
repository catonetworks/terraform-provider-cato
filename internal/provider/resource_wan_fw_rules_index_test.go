package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestNewWanRulesIndexResource(t *testing.T) {
	t.Parallel()

	res := NewWanRulesIndexResource()
	if _, ok := res.(*wanRulesIndexResource); !ok {
		t.Fatalf("expected *wanRulesIndexResource, got %T", res)
	}
}

func TestWanRulesIndexResourceMetadata(t *testing.T) {
	t.Parallel()

	r := &wanRulesIndexResource{}
	req := resource.MetadataRequest{ProviderTypeName: "cato"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "cato_bulk_wf_move_rule" {
		t.Fatalf("unexpected type name %q", resp.TypeName)
	}
}

func TestWanRulesIndexResourceConfigure(t *testing.T) {
	t.Parallel()

	t.Run("nil provider data", func(t *testing.T) {
		t.Parallel()

		r := &wanRulesIndexResource{}
		req := resource.ConfigureRequest{ProviderData: nil}
		resp := &resource.ConfigureResponse{}

		r.Configure(context.Background(), req, resp)

		if r.client != nil {
			t.Fatalf("expected nil client when provider data is nil")
		}
	})

	t.Run("sets provider client", func(t *testing.T) {
		t.Parallel()

		providerData := &catoClientData{AccountId: "account-123"}
		r := &wanRulesIndexResource{}
		req := resource.ConfigureRequest{ProviderData: providerData}
		resp := &resource.ConfigureResponse{}

		r.Configure(context.Background(), req, resp)

		if r.client != providerData {
			t.Fatalf("expected resource client to be set from provider data")
		}
	})
}

func TestWanRulesIndexResourceGetClient(t *testing.T) {
	t.Parallel()

	t.Run("nil without provider client", func(t *testing.T) {
		t.Parallel()

		r := &wanRulesIndexResource{}
		if got := r.getWanRulesIndexClient(); got != nil {
			t.Fatalf("expected nil client, got %T", got)
		}
	})

	t.Run("uses injected client", func(t *testing.T) {
		t.Parallel()

		mockClient := mocks.NewWanRulesIndexClient(t)
		r := &wanRulesIndexResource{wanRulesIndexClient: mockClient}
		if got := r.getWanRulesIndexClient(); got != mockClient {
			t.Fatalf("expected injected client, got %T", got)
		}
	})

	t.Run("falls back to provider client", func(t *testing.T) {
		t.Parallel()

		sdkClient := &cato_go_sdk.Client{}
		r := &wanRulesIndexResource{client: &catoClientData{catov2: sdkClient}}
		if got := r.getWanRulesIndexClient(); got != sdkClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
}

func TestWanRulesIndexCreateReturnsDiagnosticsOnSectionsIndexError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(nil, errors.New("sections index failed")).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newWanRulesIndexPlan(ctx, t)}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWanRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for sections index error")
	}
}

func TestWanRulesIndexUpdateReturnsDiagnosticsOnSectionsIndexError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(nil, errors.New("sections index failed")).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}
	req := resource.UpdateRequest{Plan: newWanRulesIndexPlan(ctx, t)}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getWanRulesIndexSchema(ctx, t)}}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for sections index error")
	}
}

func TestKeepPlannedMapKeys_UsesActualForPlannedKeysOnly(t *testing.T) {
	t.Parallel()

	planMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "plan-id", "managed", 1),
	})
	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed":  mustSectionObject(t, "actual-id", "managed", 10),
		"unmanaged": mustSectionObject(t, "extra-id", "unmanaged", 20),
	})

	got, diags := keepPlannedMapKeys(planMap, actualMap, WanSectionIndexResourceObjectTypes)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %+v", diags)
	}

	gotElements := got.Elements()
	if len(gotElements) != 1 {
		t.Fatalf("expected one managed element, got %d", len(gotElements))
	}

	obj, ok := gotElements["managed"].(types.Object)
	if !ok {
		t.Fatalf("expected managed value to be object, got %T", gotElements["managed"])
	}
	gotSection := asSectionItem(t, context.Background(), obj)
	if gotSection.ID.ValueString() != "actual-id" {
		t.Fatalf("expected managed element from actual state, got id %q", gotSection.ID.ValueString())
	}
	if gotSection.SectionIndex.ValueInt64() != 10 {
		t.Fatalf("expected section_index 10 from actual state, got %d", gotSection.SectionIndex.ValueInt64())
	}
}

func TestKeepPlannedMapKeys_UsesPlanValueWhenActualKeyMissing(t *testing.T) {
	t.Parallel()

	planMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "plan-id", "managed", 1),
	})
	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{})

	got, diags := keepPlannedMapKeys(planMap, actualMap, WanSectionIndexResourceObjectTypes)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %+v", diags)
	}

	gotElements := got.Elements()
	if len(gotElements) != 1 {
		t.Fatalf("expected one managed element, got %d", len(gotElements))
	}

	obj, ok := gotElements["managed"].(types.Object)
	if !ok {
		t.Fatalf("expected managed value to be object, got %T", gotElements["managed"])
	}
	gotSection := asSectionItem(t, context.Background(), obj)
	if gotSection.ID.ValueString() != "plan-id" {
		t.Fatalf("expected managed element to fall back to plan value, got id %q", gotSection.ID.ValueString())
	}
}

func TestKeepPlannedMapKeys_ReturnsActualWhenPlanMapNullOrUnknown(t *testing.T) {
	t.Parallel()

	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "actual-id", "managed", 10),
	})

	testCases := []struct {
		name    string
		planMap types.Map
	}{
		{
			name:    "null plan",
			planMap: types.MapNull(WanSectionIndexResourceObjectTypes),
		},
		{
			name:    "unknown plan",
			planMap: types.MapUnknown(WanSectionIndexResourceObjectTypes),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := keepPlannedMapKeys(tc.planMap, actualMap, WanSectionIndexResourceObjectTypes)
			if diags.HasError() {
				t.Fatalf("expected no diagnostics, got: %+v", diags)
			}
			if !got.Equal(actualMap) {
				t.Fatalf("expected actual map to be returned when plan is %s", tc.name)
			}
		})
	}
}

func mustSectionObject(t *testing.T, id, sectionName string, sectionIndex int64) basetypes.ObjectValue {
	t.Helper()

	value, diags := types.ObjectValue(
		WanSectionIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":            types.StringValue(id),
			"section_name":  types.StringValue(sectionName),
			"section_index": types.Int64Value(sectionIndex),
		},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating section object: %+v", diags)
	}
	return value
}

func mustMapValue(t *testing.T, objectType types.ObjectType, elements map[string]attr.Value) types.Map {
	t.Helper()

	value, diags := types.MapValue(objectType, elements)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating map value: %+v", diags)
	}
	return value
}

func asSectionItem(t *testing.T, ctx context.Context, object types.Object) WanRulesSectionItemIndex {
	t.Helper()

	var item WanRulesSectionItemIndex
	diags := object.As(ctx, &item, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics decoding object: %+v", diags)
	}
	return item
}

func getWanRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &wanRulesIndexResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	return schemaResp.Schema
}

func newWanRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	emptyRuleMap := mustMapValue(t, WanRuleIndexResourceObjectTypes, map[string]attr.Value{})
	emptySectionMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{})

	plan := tfsdk.Plan{Schema: getWanRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, WanRulesIndex{
		SectionToStartAfterID: types.StringNull(),
		RuleData:              emptyRuleMap,
		SectionData:           emptySectionMap,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating plan: %+v", diags)
	}

	return plan
}
