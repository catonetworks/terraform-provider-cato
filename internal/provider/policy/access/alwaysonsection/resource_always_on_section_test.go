package alwaysonsection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func sectionModel(id, name, position, ref string) Section {
	return Section{
		At: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue(position), "ref": types.StringValue(ref),
		}),
		Section: types.ObjectValueMust(SectionDataAttrTypes, map[string]attr.Value{
			"id": types.StringValue(id), "name": types.StringValue(name),
		}),
	}
}

func sectionPlan(t *testing.T, model Section) tfsdk.Plan {
	t.Helper()
	var schema resource.SchemaResponse
	NewAlwaysOnSectionResource().Schema(context.Background(), resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, plan.Set(context.Background(), model).HasError())
	return plan
}

func sectionState(t *testing.T, model Section) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: sectionPlan(t, model).Schema}
	require.False(t, state.Set(context.Background(), model).HasError())
	return state
}

func decodeResponse[T any](t *testing.T, body string) *T {
	t.Helper()
	if body == "" {
		return nil
	}
	var result T
	require.NoError(t, json.Unmarshal([]byte(body), &result))
	return &result
}

func mutationJSON(operation, id, name string) string {
	return fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{%q:{"status":"SUCCESS","errors":[],"section":{"section":{"id":%q,"name":%q}}}}}}`, operation, id, name)
}

func TestSectionLifecycleWiring(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewAlwaysOnSectionResource().(*alwaysOnSectionResource)
	var metadata resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_always_on_section", metadata.TypeName)
	var configured resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &configured)
	require.Nil(t, r.getClient())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &configured)
	require.True(t, configured.Diagnostics.HasError())
	client := &catoclient.Service{AccountId: "123"}
	configured = resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: client}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	require.Same(t, client, r.client)
	require.Nil(t, r.getClient())
	state := sectionState(t, sectionModel("", "", "LAST_IN_POLICY", ""))
	response := resource.ImportStateResponse{State: state}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "section-123"}, &response)
	require.False(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(ctx, path.Root("section").AtName("id"), &id).HasError())
	require.Equal(t, "section-123", id.ValueString())
}

func TestSectionCreate(t *testing.T) {
	t.Parallel()
	client := mocks.NewAlwaysOnSectionClient(t)
	client.On("PolicyZtnaAlwaysOnAddSection", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
		cato_models.PolicyAddSectionInput{At: &cato_models.PolicySectionPositionInput{Position: "LAST_IN_POLICY"},
			Section: &cato_models.PolicyAddSectionInfoInput{Name: "Example"}}).
		Return(decodeResponse[cato.PolicyZtnaAlwaysOnAddSection](t, mutationJSON("addSection", "section-123", "Example")), nil).Once()
	r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	model := sectionModel("", "Example", "LAST_IN_POLICY", "")
	attrs := model.Section.Attributes()
	attrs["id"] = types.StringUnknown()
	model.Section = types.ObjectValueMust(SectionDataAttrTypes, attrs)
	plan := sectionPlan(t, model)
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var actual Section
	require.False(t, response.State.Get(context.Background(), &actual).HasError())
	require.Equal(t, types.StringValue("section-123"), actual.Section.Attributes()["id"])
	require.True(t, model.At.Equal(actual.At))
}

func TestSectionCreateErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "transport", err: errors.New("request failed"), message: "request failed"},
		{name: "nil response", message: "mutation status"},
		{name: "failure", body: `{"policy":{"ztnaAlwaysOn":{"addSection":{"status":"FAILURE"}}}}`, message: "FAILURE"},
		{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"addSection":{"status":"SUCCESS","errors":[{"errorMessage":"Rejected"}]}}}}`, message: "Rejected"},
		{name: "missing ID", body: mutationJSON("addSection", "", "Example"), message: "resource ID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnSectionClient(t)
			client.On("PolicyZtnaAlwaysOnAddSection", mock.Anything, "123", mock.Anything, mock.Anything).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnAddSection](t, tc.body), tc.err).Once()
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			plan := sectionPlan(t, sectionModel("", "Example", "LAST_IN_POLICY", ""))
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
			require.True(t, response.Diagnostics.HasError())
			require.Contains(t, response.Diagnostics.Errors()[0].Detail(), tc.message)
		})
	}
}

func TestSectionRead(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, position, ref, message string
		removed, imported                  bool
		err                                error
	}{
		{name: "refresh", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[{"section":{"id":"section-123","name":"Remote"}}]}}}}`},
		{name: "missing", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[]}}}}`, removed: true},
		{name: "empty section import", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[{"section":{"id":"section-123","name":"Remote"}},{"section":{"id":"next","name":"Next"}}],"rules":[]}}}}`, imported: true, position: "BEFORE_SECTION", ref: "next"},
		{name: "last section import", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[{"section":{"id":"section-123","name":"Remote"}}]}}}}`, imported: true, position: "LAST_IN_POLICY"},
		{name: "nil response", message: "section list"},
		{name: "missing section list", body: `{"policy":{"ztnaAlwaysOn":{"policy":{}}}}`, message: "section list"},
		{name: "malformed entry", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[null]}}}}`, message: "without an ID"},
		{name: "duplicate IDs", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[{"section":{"id":"section-123"}},{"section":{"id":"section-123"}}]}}}}`, message: "duplicate"},
		{name: "transport", err: errors.New("read failed"), message: "read failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			client := mocks.NewAlwaysOnSectionClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnPolicy](t, tc.body), tc.err).Once()
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			model := sectionModel("section-123", "Example", "AFTER_SECTION", "previous")
			if tc.imported {
				model.At = types.ObjectNull(policyrule.PositionAttrTypes)
			}
			state := sectionState(t, model)
			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			if tc.message != "" {
				require.True(t, response.Diagnostics.HasError())
				require.Contains(t, response.Diagnostics.Errors()[0].Detail(), tc.message)
				require.True(t, state.Raw.Equal(response.State.Raw))
				return
			}
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if tc.removed {
				require.True(t, response.State.Raw.IsNull())
				return
			}
			var actual Section
			require.False(t, response.State.Get(ctx, &actual).HasError())
			require.Equal(t, types.StringValue("Remote"), actual.Section.Attributes()["name"])
			if tc.imported {
				require.Equal(t, types.StringValue(tc.position), actual.At.Attributes()["position"])
				require.Equal(t, types.StringValue(tc.ref), actual.At.Attributes()["ref"])
			} else {
				require.True(t, model.At.Equal(actual.At))
			}
		})
	}
}

func TestSectionUpdate(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"rename", "move", "move fails"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			client := mocks.NewAlwaysOnSectionClient(t)
			name := "Renamed"
			update := client.On("PolicyZtnaAlwaysOnUpdateSection", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
				cato_models.PolicyUpdateSectionInput{ID: "section-123", Section: &cato_models.PolicyUpdateSectionInfoInput{Name: &name}}).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnUpdateSection](t, mutationJSON("updateSection", "section-123", name)), nil).Once()
			model := sectionModel("section-123", name, "LAST_IN_POLICY", "")
			if mode != "rename" {
				ref := "other"
				model = sectionModel("section-123", name, "BEFORE_SECTION", ref)
				var err error
				if mode == "move fails" {
					err = errors.New("move failed")
				}
				client.On("PolicyZtnaAlwaysOnMoveSection", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
					cato_models.PolicyMoveSectionInput{ID: "section-123", To: &cato_models.PolicySectionPositionInput{Position: "BEFORE_SECTION", Ref: &ref}}).
					Return(decodeResponse[cato.PolicyZtnaAlwaysOnMoveSection](t, mutationJSON("moveSection", "section-123", name)), err).Once().NotBefore(update)
			}
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := sectionState(t, sectionModel("section-123", "Example", "LAST_IN_POLICY", ""))
			response := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: sectionPlan(t, model), State: state}, &response)
			require.Equal(t, mode == "move fails", response.Diagnostics.HasError(), "%v", response.Diagnostics)
			var actual Section
			require.False(t, response.State.Get(ctx, &actual).HasError())
			require.Equal(t, types.StringValue(name), actual.Section.Attributes()["name"])
			if mode == "move fails" {
				require.Equal(t, types.StringValue("LAST_IN_POLICY"), actual.At.Attributes()["position"])
			} else {
				require.True(t, model.At.Equal(actual.At))
			}
		})
	}
}

func TestSectionDelete(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "success", body: mutationJSON("removeSection", "section-123", "Example")},
		{name: "transport", err: errors.New("delete failed"), message: "delete failed"},
		{name: "nil response", message: "mutation status"},
		{name: "failure", body: `{"policy":{"ztnaAlwaysOn":{"removeSection":{"status":"FAILURE"}}}}`, message: "FAILURE"},
		{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"removeSection":{"status":"SUCCESS","errors":[{"errorMessage":"Policy locked"}]}}}}`, message: "Policy locked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnSectionClient(t)
			client.On("PolicyZtnaAlwaysOnRemoveSection", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
				cato_models.PolicyRemoveSectionInput{ID: "section-123"}).Return(decodeResponse[cato.PolicyZtnaAlwaysOnRemoveSection](t, tc.body), tc.err).Once()
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := sectionState(t, sectionModel("section-123", "Example", "LAST_IN_POLICY", ""))
			response := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
			if tc.message == "" {
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				require.True(t, response.State.Raw.IsNull())
			} else {
				require.True(t, response.Diagnostics.HasError())
				require.Contains(t, response.Diagnostics.Errors()[0].Detail(), tc.message)
				require.True(t, state.Raw.Equal(response.State.Raw))
			}
		})
	}
}

func TestSectionUpdateMutationErrors(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"updateSection", "moveSection"} {
		for _, failure := range []string{"transport", "nil response", "failure status", "wrong ID", "missing ID", "payload error"} {
			t.Run(stage+"/"+failure, func(t *testing.T) {
				t.Parallel()
				client := mocks.NewAlwaysOnSectionClient(t)
				body := mutationJSON(stage, "section-123", "Renamed")
				var err error
				switch failure {
				case "transport":
					err = errors.New("request failed")
				case "nil response":
					body = ""
				case "failure status":
					body = fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{%q:{"status":"FAILURE"}}}}`, stage)
				case "wrong ID":
					body = mutationJSON(stage, "other", "Renamed")
				case "missing ID":
					body = mutationJSON(stage, "", "Renamed")
				case "payload error":
					body = fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{%q:{"status":"SUCCESS","errors":[{"errorMessage":"Rejected"}]}}}}`, stage)
				}
				if stage == "updateSection" {
					client.On("PolicyZtnaAlwaysOnUpdateSection", mock.Anything, "123", mock.Anything, mock.Anything).
						Return(decodeResponse[cato.PolicyZtnaAlwaysOnUpdateSection](t, body), err).Once()
				} else {
					update := client.On("PolicyZtnaAlwaysOnUpdateSection", mock.Anything, "123", mock.Anything, mock.Anything).
						Return(decodeResponse[cato.PolicyZtnaAlwaysOnUpdateSection](t, mutationJSON("updateSection", "section-123", "Renamed")), nil).Once()
					client.On("PolicyZtnaAlwaysOnMoveSection", mock.Anything, "123", mock.Anything, mock.Anything).
						Return(decodeResponse[cato.PolicyZtnaAlwaysOnMoveSection](t, body), err).Once().NotBefore(update)
				}
				r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
				state := sectionState(t, sectionModel("section-123", "Example", "LAST_IN_POLICY", ""))
				plan := sectionPlan(t, sectionModel("section-123", "Renamed", "BEFORE_SECTION", "other"))
				response := resource.UpdateResponse{State: state}
				r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
				if stage == "updateSection" {
					require.True(t, state.Raw.Equal(response.State.Raw))
					return
				}
				var actual Section
				require.False(t, response.State.Get(context.Background(), &actual).HasError())
				require.Equal(t, types.StringValue("Renamed"), actual.Section.Attributes()["name"])
				require.Equal(t, types.StringValue("LAST_IN_POLICY"), actual.At.Attributes()["position"])
			})
		}
	}
}

func TestSectionUpdateStaleMutationResponse(t *testing.T) {
	t.Parallel()
	for _, persisted := range []bool{true, false} {
		t.Run(fmt.Sprint(persisted), func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnSectionClient(t)
			update := client.On("PolicyZtnaAlwaysOnUpdateSection", mock.Anything, "123", mock.Anything, mock.Anything).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnUpdateSection](t, mutationJSON("updateSection", "section-123", "Example")), nil).Once()
			name := "Example"
			if persisted {
				name = "Renamed"
			}
			body := fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{"policy":{"sections":[{"section":{"id":"section-123","name":%q}}]}}}}`, name)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnPolicy](t, body), nil).Once().NotBefore(update)
			r := &alwaysOnSectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := sectionState(t, sectionModel("section-123", "Example", "LAST_IN_POLICY", ""))
			plan := sectionPlan(t, sectionModel("section-123", "Renamed", "LAST_IN_POLICY", ""))
			response := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &response)
			require.Equal(t, !persisted, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			if persisted {
				var actual Section
				require.False(t, response.State.Get(context.Background(), &actual).HasError())
				require.Equal(t, types.StringValue("Renamed"), actual.Section.Attributes()["name"])
			} else {
				require.True(t, state.Raw.Equal(response.State.Raw))
			}
		})
	}
}
