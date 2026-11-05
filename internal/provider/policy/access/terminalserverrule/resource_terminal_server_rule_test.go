package terminalserverrule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

type testBackend struct {
	publicObject        map[string]any
	snapshot            bool
	revisionExists      bool
	losePublishResponse bool
	object              map[string]any
	rules               []any
	rulesOnRevision     []any
	calls               []string
	variables           []map[string]any
	fail                string
	malformed           bool
}

func testResource(t *testing.T, backend *testBackend) *policyResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var request struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
		operation := strings.TrimPrefix(request.OperationName, "policyTerminalServer")
		operation = strings.ToLower(operation[:1]) + operation[1:]
		backend.calls = append(backend.calls, operation)
		backend.variables = append(backend.variables, request.Variables)
		require.Equal(t, "123", request.Variables["accountId"])
		if operation != "policy" && operation != "createPolicyRevision" {
			require.Equal(t, map[string]any{"revision": map[string]any{"id": "revision-1"}}, request.Variables["terminalServerPolicyMutationInput"])
		}
		w.Header().Set("Content-Type", "application/json")
		if backend.fail == operation {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if backend.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		writeBackendResponse(t, w, backend, operation, request.Variables)
	}))
	t.Cleanup(server.Close)
	client, err := cato.New(server.URL, "test-token", "sdk-default-account", nil, nil)
	require.NoError(t, err)
	return &policyResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
}

func testSchema(t *testing.T) schema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	return response.Schema
}

func testBackendData(t *testing.T) *testBackend {
	t.Helper()
	var object map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{"allowedHostIP": {"id": "item-1", "name": "Example"}, "description": "Managed by Terraform", "enabled": true, "excludeTraffic": [{"id": "item-1", "name": "Example"}], "id": "item-1", "index": 1, "name": "Example", "section": {"id": "section-1", "name": "Default"}}`), &object))
	return &testBackend{object: object}
}

func testModel(t *testing.T) resourceModel {
	t.Helper()
	backend := testBackendData(t)
	// Decode through the real generated SDK projection used by Read.
	body, err := json.Marshal(map[string]any{"policy": map[string]any{"terminalServer": map[string]any{"policy": map[string]any{"rules": []any{map[string]any{"rule": backend.object}}}}}})
	require.NoError(t, err)
	var result cato.PolicyTerminalServerPolicy
	require.NoError(t, json.Unmarshal(body, &result))
	var response resource.ReadResponse
	value := parseRead(context.Background(), result.GetPolicy().GetTerminalServer().GetPolicy().GetRules()[0].GetRule(), &response.Diagnostics)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	return resourceModel{Rule: value, At: types.ObjectNull(atAttrTypes)}
}

func testPlan(t *testing.T, model resourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: testSchema(t)}
	require.False(t, plan.Set(context.Background(), model).HasError())
	return plan
}

func testState(t *testing.T, model resourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: testSchema(t)}
	require.False(t, state.Set(context.Background(), model).HasError())
	return state
}

func TestLifecycle(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	backend := testBackendData(t)
	r := testResource(t, backend)
	model := testModel(t)
	create := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, model)}, &create)
	require.False(t, create.Diagnostics.HasError(), "%v", create.Diagnostics)
	require.Equal(t, []string{"createPolicyRevision", "addRule", "publishPolicyRevision"}, backend.calls)
	var id types.String
	require.False(t, create.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id).HasError())
	require.Equal(t, "item-1", id.ValueString())
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	require.False(t, read.State.Get(ctx, &model).HasError())
	fields := model.Rule.Attributes()
	fields["name"] = types.StringValue("Updated")
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	position := map[string]attr.Value{"position": types.StringValue("FIRST_IN_POLICY"), "ref": types.StringValue("")}
	model.At = types.ObjectValueMust(atAttrTypes, position)
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: testPlan(t, model)}, &update)
	require.False(t, update.Diagnostics.HasError(), "%v", update.Diagnostics)
	require.Equal(t, []string{"createPolicyRevision", "updateRule", "moveRule", "publishPolicyRevision"}, backend.calls[len(backend.calls)-4:])
	var name types.String
	require.False(t, update.State.GetAttribute(ctx, path.Root("rule").AtName("name"), &name).HasError())
	require.Equal(t, "Updated", name.ValueString())
	deletion := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &deletion)
	require.False(t, deletion.Diagnostics.HasError(), "%v", deletion.Diagnostics)
	require.Equal(t, "publishPolicyRevision", backend.calls[len(backend.calls)-1])
	missing := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &missing)
	require.False(t, missing.Diagnostics.HasError(), "%v", missing.Diagnostics)
	require.True(t, missing.State.Raw.IsNull())
}

func TestCreateFailures(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"createPolicyRevision", "addRule", "publishPolicyRevision"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			backend := testBackendData(t)
			backend.fail = operation
			r := testResource(t, backend)
			response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, testModel(t))}, &response)
			require.True(t, response.Diagnostics.HasError())
			if operation == "publishPolicyRevision" {
				var id types.String
				require.False(t, response.State.GetAttribute(context.Background(), path.Root("rule").AtName("id"), &id).HasError())
				require.Equal(t, "item-1", id.ValueString())
			}
		})
	}
}

func TestReadMalformed(t *testing.T) {
	t.Parallel()
	backend := testBackendData(t)
	backend.malformed = true
	r := testResource(t, backend)
	state := testState(t, testModel(t))
	response := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.False(t, response.State.Raw.IsNull())
}

func TestImportAndConfiguration(t *testing.T) {
	t.Parallel()
	r := NewResource().(*policyResource)
	var configured resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &configured)
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "wrong"}, &configured)
	require.True(t, configured.Diagnostics.HasError())
	var metadata resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_terminal_server_rule", metadata.TypeName)
	response := resource.ImportStateResponse{State: testState(t, testModel(t))}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "imported"}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var id types.String
	require.False(t, response.State.GetAttribute(context.Background(), path.Root("rule").AtName("id"), &id).HasError())
	require.Equal(t, "imported", id.ValueString())
}

func TestCreateUnknownObject(t *testing.T) {
	t.Parallel()
	model := testModel(t)
	model.Rule = types.ObjectUnknown(ruleAttrTypes)
	r := &policyResource{}
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
	require.True(t, response.Diagnostics.HasError())
}

type testPrivateState struct{ value []byte }

func (s *testPrivateState) GetKey(_ context.Context, _ string) ([]byte, diag.Diagnostics) {
	return s.value, nil
}
func (s *testPrivateState) SetKey(_ context.Context, _ string, value []byte) diag.Diagnostics {
	s.value = value
	return nil
}

func TestPendingRevisionRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &testPrivateState{}
	var diagnostics diag.Diagnostics
	policyshared.RememberRevision(ctx, store, "revision-1", &diagnostics)
	require.Equal(t, "revision-1", policyshared.PendingRevision(ctx, store, &diagnostics))
	require.False(t, diagnostics.HasError())
	policyshared.RememberRevision(ctx, store, "", &diagnostics)
	require.Empty(t, policyshared.PendingRevision(ctx, store, &diagnostics))
	store.value = []byte("invalid")
	require.Empty(t, policyshared.PendingRevision(ctx, store, &diagnostics))
	require.True(t, diagnostics.HasError())
}

func TestPositionRequiresReference(t *testing.T) {
	t.Parallel()
	model := testModel(t)
	model.At = types.ObjectValueMust(atAttrTypes, map[string]attr.Value{
		"position": types.StringValue("AFTER_RULE"), "ref": types.StringNull(),
	})
	r := &policyResource{}
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.Contains(t, fmt.Sprint(response.Diagnostics), "Missing position reference")
}

func TestInjectedClient(t *testing.T) {
	t.Parallel()
	client := mocks.NewTerminalServerRuleClient(t)
	client.On("PolicyTerminalServerCreatePolicyRevision", mock.Anything, "123", (*cato_models.TerminalServerPolicyMutationInput)(nil), mock.Anything).
		Return(nil, errors.New("denied")).Once()
	r := &policyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, testModel(t))}, &response)
	require.True(t, response.Diagnostics.HasError())
}

func TestUpdateAndDeleteFailures(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"updateRule", "moveRule", "removeRule", "publishPolicyRevision"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			backend := testBackendData(t)
			backend.fail = operation
			r := testResource(t, backend)
			model := testModel(t)
			state := testState(t, model)
			if operation == "removeRule" {
				response := resource.DeleteResponse{State: state}
				r.Delete(ctx, resource.DeleteRequest{State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
				require.False(t, response.State.Raw.IsNull())
				return
			}
			model.At = types.ObjectValueMust(atAttrTypes, map[string]attr.Value{
				"position": types.StringValue("AFTER_RULE"), "ref": types.StringValue("other"),
			})
			response := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{State: state, Plan: testPlan(t, model)}, &response)
			require.True(t, response.Diagnostics.HasError())
			require.False(t, response.State.Raw.IsNull())
		})
	}
}

// Terraform initializes its internal private-state type before calling resources.
// Reflection reproduces that initialization without importing a framework-internal package.
func initializePrivate(response any) {
	field := reflect.ValueOf(response).Elem().FieldByName("Private")
	field.Set(reflect.New(field.Type().Elem()))
}

func TestPublicationRecovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	backend := testBackendData(t)
	backend.fail = "publishPolicyRevision"
	r := testResource(t, backend)
	create := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	initializePrivate(&create)
	r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, testModel(t))}, &create)
	require.True(t, create.Diagnostics.HasError())
	var diagnostics diag.Diagnostics
	require.Equal(t, "revision-1", policyshared.PendingRevision(ctx, create.Private, &diagnostics))
	backend.fail = ""
	read := resource.ReadResponse{State: create.State, Private: create.Private}
	r.Read(ctx, resource.ReadRequest{State: create.State, Private: create.Private}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	require.Equal(t, "revision-1", policyshared.PendingRevision(ctx, read.Private, &diagnostics))
	require.Equal(t, []string{"createPolicyRevision", "addRule", "publishPolicyRevision", "policy", "policy"}, backend.calls)
}

func TestDeletePublicationRecovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	backend := testBackendData(t)
	backend.fail = "publishPolicyRevision"
	r := testResource(t, backend)
	state := testState(t, testModel(t))
	deletion := resource.DeleteResponse{State: state}
	initializePrivate(&deletion)
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deletion)
	require.True(t, deletion.Diagnostics.HasError())
	require.Nil(t, backend.object)
	backend.fail = ""
	retry := resource.DeleteResponse{State: state, Private: deletion.Private}
	r.Delete(ctx, resource.DeleteRequest{State: state, Private: deletion.Private}, &retry)
	require.False(t, retry.Diagnostics.HasError(), "%v", retry.Diagnostics)
	require.Equal(t, 2, operationCount(backend.calls, "createPolicyRevision"))
	require.Equal(t, 2, operationCount(backend.calls, "removeRule"))
	var diagnostics diag.Diagnostics
	require.Empty(t, policyshared.PendingRevision(ctx, retry.Private, &diagnostics))
}

func TestCreateFailureDiscardsOnlyOwnedRevision(t *testing.T) {
	t.Parallel()
	backend := testBackendData(t)
	backend.fail = "addRule"
	r := testResource(t, backend)
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, testModel(t))}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.Equal(t, []string{"createPolicyRevision", "addRule", "discardPolicyRevision"}, backend.calls)
	variables := backend.variables[len(backend.variables)-1]
	require.Equal(t, map[string]any{"id": "revision-1"}, variables["policyDiscardRevisionInput"])
}

func clearCriteria(value types.Object) types.Object {
	fields := value.Attributes()
	for key, item := range fields {
		switch typed := item.(type) {
		case types.Set:
			fields[key] = types.SetValueMust(typed.ElementType(context.Background()), nil)
		case types.Object:
			if !typed.IsNull() && !typed.IsUnknown() {
				fields[key] = clearCriteria(typed)
			}
		case types.Bool:
			fields[key] = types.BoolValue(false)
		case types.String:
			if key == "description" {
				fields[key] = types.StringValue("")
			}
		}
	}
	return types.ObjectValueMust(value.AttributeTypes(context.Background()), fields)
}

func TestClearCriteriaRequestShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	backend := testBackendData(t)
	r := testResource(t, backend)
	model := testModel(t)
	state := testState(t, model)
	model.Rule = clearCriteria(model.Rule)
	response := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: testPlan(t, model)}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var input map[string]any
	for _, variables := range backend.variables {
		if value, ok := variables["terminalServerUpdateRuleInput"].(map[string]any); ok {
			input = value
		}
	}
	require.NotNil(t, input)
	object := input["rule"].(map[string]any)
	require.Equal(t, []any{}, object["excludeTraffic"], "excludeTraffic")
	require.Equal(t, false, object["enabled"])
	require.Equal(t, "", object["description"])
}

func writeBackendResponse(t *testing.T, w http.ResponseWriter, backend *testBackend, operation string, variables map[string]any) {
	t.Helper()
	policy := map[string]any{"rules": []any{}, "sections": []any{}}
	if backend.rules != nil {
		policy["rules"] = backend.rules
	}
	if backend.object != nil {
		policy["rules"] = []any{map[string]any{"rule": backend.object}}
	}
	payload := map[string]any{"status": "SUCCESS", "errors": []any{}}
	switch operation {
	case "policy":
		writeReadBackend(t, w, backend, variables, policy)
		return
	case "createPolicyRevision":
		if !backend.snapshot {
			backend.publicObject = cloneObject(backend.object)
			backend.snapshot = true
		}
		backend.object = cloneObject(backend.publicObject)
		backend.revisionExists = true
		if backend.rulesOnRevision != nil {
			backend.rules = backend.rulesOnRevision
			policy["rules"] = backend.rules
		}
		policy["revision"] = map[string]any{"id": "revision-1"}
		payload["policy"] = policy
	case "addRule", "updateRule", "moveRule":
		updateBackendName(backend, variables)
		payload["rule"] = map[string]any{"rule": backend.object}
	case "removeRule":
		backend.object = nil
	case "discardPolicyRevision":
		if !backend.revisionExists {
			payload["status"] = "FAILURE"
			payload["errors"] = []any{map[string]any{"errorCode": "PolicyRevisionNotFound"}}
			break
		}
		backend.object = cloneObject(backend.publicObject)
		backend.revisionExists = false
	case "publishPolicyRevision":
		backend.publicObject = cloneObject(backend.object)
		backend.revisionExists = false
		if backend.losePublishResponse {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "connection lost after commit"}}}))
			return
		}
		payload["policy"] = policy
	}
	body := map[string]any{"data": map[string]any{"policy": map[string]any{"terminalServer": map[string]any{operation: payload}}}}
	require.NoError(t, json.NewEncoder(w).Encode(body))
}

func operationCount(calls []string, operation string) int {
	count := 0
	for _, call := range calls {
		if call == operation {
			count++
		}
	}
	return count
}

func cloneObject(object map[string]any) map[string]any {
	if object == nil {
		return nil
	}
	cloned := make(map[string]any, len(object))
	for key, value := range object {
		cloned[key] = value
	}
	return cloned
}

func TestPublicPolicyRequest(t *testing.T) {
	backend := testBackendData(t)
	r := testResource(t, backend)
	_, err := r.readPolicy(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, map[string]any{"revision": map[string]any{"type": "PUBLIC"}}, backend.variables[0]["terminalServerPolicyInput"])
	backend.revisionExists = true
	_, err = r.readPolicy(context.Background(), "revision-1")
	require.NoError(t, err)
	require.Equal(t, map[string]any{"revision": map[string]any{"id": "revision-1", "type": "PRIVATE"}}, backend.variables[1]["terminalServerPolicyInput"])
}

func TestRevisionNamesAreUnique(t *testing.T) {
	backend := testBackendData(t)
	r := testResource(t, backend)
	var diagnostics diag.Diagnostics
	r.beginRevision(context.Background(), "", &diagnostics)
	r.beginRevision(context.Background(), "", &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.NotEqual(t, backend.variables[0]["policyCreateRevisionInput"].(map[string]any)["name"], backend.variables[1]["policyCreateRevisionInput"].(map[string]any)["name"])
}

func TestLostPublishResponse(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			ctx := context.Background()
			backend := testBackendData(t)
			backend.losePublishResponse = true
			r := testResource(t, backend)
			model := testModel(t)
			state := testState(t, model)
			switch operation {
			case "create":
				response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
				initializePrivate(&response)
				r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, model)}, &response)
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			case "update":
				fields := model.Rule.Attributes()
				fields["name"] = types.StringValue("Updated")
				model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
				response := resource.UpdateResponse{State: state}
				initializePrivate(&response)
				r.Update(ctx, resource.UpdateRequest{Plan: testPlan(t, model), State: state}, &response)
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			case "delete":
				response := resource.DeleteResponse{State: state}
				initializePrivate(&response)
				r.Delete(ctx, resource.DeleteRequest{State: state}, &response)
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			}
			require.False(t, backend.revisionExists)
		})
	}
}

func TestFailedMoveDoesNotChangeStateOrPublishOnRead(t *testing.T) {
	ctx := context.Background()
	backend := testBackendData(t)
	backend.fail = "moveRule"
	r := testResource(t, backend)
	model := testModel(t)
	state := testState(t, model)
	fields := model.Rule.Attributes()
	fields["name"] = types.StringValue("Unpublished")
	model.Rule = types.ObjectValueMust(ruleAttrTypes, fields)
	model.At = types.ObjectValueMust(atAttrTypes, map[string]attr.Value{"position": types.StringValue("FIRST_IN_POLICY"), "ref": types.StringNull()})
	response := resource.UpdateResponse{State: state}
	initializePrivate(&response)
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: testPlan(t, model)}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.True(t, response.State.Raw.Equal(state.Raw))
	require.Equal(t, 0, operationCount(backend.calls, "publishPolicyRevision"))
	require.False(t, backend.revisionExists)
	read := resource.ReadResponse{State: state, Private: response.Private}
	r.Read(ctx, resource.ReadRequest{State: state, Private: response.Private}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	require.Equal(t, 0, operationCount(backend.calls, "publishPolicyRevision"))
}

func writeReadBackend(t *testing.T, w http.ResponseWriter, backend *testBackend, variables map[string]any, policy map[string]any) {
	t.Helper()
	input, _ := variables["terminalServerPolicyInput"].(map[string]any)
	revision, _ := input["revision"].(map[string]any)
	if revision["type"] == "PUBLIC" {
		object := backend.object
		if backend.snapshot {
			object = backend.publicObject
		}
		policy["rules"] = []any{}
		if object != nil {
			policy["rules"] = []any{map[string]any{"rule": object}}
		}
	} else if !backend.revisionExists {
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "PolicyRevisionNotFound", "extensions": map[string]any{"errorCode": "PolicyRevisionNotFound"}}}}))
		return
	}
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"policy": map[string]any{"terminalServer": map[string]any{"policy": policy}}}}))
}

func TestDestroyWithStaleRevision(t *testing.T) {
	ctx := context.Background()
	backend := testBackendData(t)
	r := testResource(t, backend)
	state := testState(t, testModel(t))
	response := resource.DeleteResponse{State: state}
	initializePrivate(&response)
	var diagnostics diag.Diagnostics
	policyshared.RememberRevision(ctx, response.Private, "revision-1", &diagnostics)
	r.Delete(ctx, resource.DeleteRequest{State: state, Private: response.Private}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	require.Equal(t, 1, operationCount(backend.calls, "removeRule"))
	require.Nil(t, backend.publicObject)
}

func TestReadDoesNotPublishPendingDeletion(t *testing.T) {
	ctx := context.Background()
	backend := testBackendData(t)
	backend.fail = "publishPolicyRevision"
	r := testResource(t, backend)
	state := testState(t, testModel(t))
	deletion := resource.DeleteResponse{State: state}
	initializePrivate(&deletion)
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deletion)
	require.True(t, deletion.Diagnostics.HasError())
	backend.fail = ""
	read := resource.ReadResponse{State: state, Private: deletion.Private}
	r.Read(ctx, resource.ReadRequest{State: state, Private: deletion.Private}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	require.False(t, read.State.Raw.IsNull())
	require.Equal(t, 1, operationCount(backend.calls, "publishPolicyRevision"))
	require.NotNil(t, backend.publicObject)
}

func updateBackendName(backend *testBackend, variables map[string]any) {
	for _, input := range variables {
		values, ok := input.(map[string]any)
		if !ok {
			continue
		}
		if object, ok := values["rule"].(map[string]any); ok {
			if name, ok := object["name"]; ok {
				backend.object["name"] = name
			}
		}
	}
}
