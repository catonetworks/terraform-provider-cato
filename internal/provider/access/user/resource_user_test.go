package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type userBackend struct {
	object    map[string]any
	calls     []string
	variables []map[string]any
	fail      string
	paged     bool
	malformed bool
}

func newUserBackend() *userBackend {
	return &userBackend{object: map[string]any{
		"id": "1001", "firstName": "Test", "lastName": "User", "email": "test@example.invalid",
		"department": "Engineering", "jobTitle": "Developer", "phone": "+15555550123",
		"displayName": "Test User", "userStatus": "ACTIVE", "userImportType": "MANUAL",
	}}
}

func testResource(t *testing.T, b *userBackend) *userResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		defer func() { require.NoError(t, request.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Query         string         `json:"query"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		operation := strings.TrimPrefix(body.OperationName, "user")
		operation = strings.ToLower(operation[:1]) + operation[1:]
		b.calls = append(b.calls, operation)
		b.variables = append(b.variables, body.Variables)
		require.Equal(t, "account-123", body.Variables["accountId"])
		w.Header().Set("Content-Type", "application/json")
		if operation == b.fail {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		payload := map[string]any{}
		switch operation {
		case "createUser":
			for _, field := range []string{"department", "jobTitle", "phone"} {
				b.object[field] = nil
			}
			input := body.Variables["createUserInput"].(map[string]any)
			for key, value := range input {
				if key == "phoneNumber" {
					key = "phone"
				}
				b.object[key] = value
			}
			b.object["displayName"] = b.object["firstName"].(string) + " " + b.object["lastName"].(string)
			payload["user"] = b.object
		case "updateUser":
			require.Equal(t, cato.UserUpdateUserDocument, body.Query)
			input := body.Variables["updateUserInput"].(map[string]any)
			require.Equal(t, "1001", input["id"])
			for key, value := range input {
				if key == "phoneNumber" {
					key = "phone"
				}
				if key != "id" {
					b.object[key] = value
				}
			}
			b.object["displayName"] = b.object["firstName"].(string) + " " + b.object["lastName"].(string)
			payload["user"] = b.object
		case "enableUser", "disableUser":
			input := body.Variables[operation+"Input"].(map[string]any)
			require.Equal(t, []any{float64(1001)}, input["userId"])
			b.object["userStatus"] = "ACTIVE"
			if operation == "disableUser" {
				b.object["userStatus"] = "DISABLED"
			}
			payload["users"] = []any{b.object}
		case "deleteUser":
			input := body.Variables["deleteUserInput"].(map[string]any)
			require.Equal(t, []any{float64(1001)}, input["userId"])
			b.object = nil
			payload["status"] = "SUCCESS"
		case "userList":
			payload = userListPayload(t, b, body.Variables["userListInput"].(map[string]any))
		default:
			t.Errorf("unexpected operation %s", operation)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user": map[string]any{operation: payload}}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "different-default-account", nil, nil)
	require.NoError(t, err)
	return &userResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}

func testSchema(t *testing.T) schema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	require.False(t, response.Diagnostics.HasError())
	return response.Schema
}

func testModel() resourceModel {
	return resourceModel{ID: types.StringValue("1001"), Email: types.StringValue("test@example.invalid"),
		FirstName: types.StringValue("Test"), LastName: types.StringValue("User"), Department: types.StringValue("Engineering"),
		JobTitle: types.StringValue("Developer"), PhoneNumber: types.StringValue("+15555550123"), Enabled: types.BoolValue(false),
		DisplayName: types.StringValue("Test User"), UserStatus: types.StringValue("DISABLED"), UserImportType: types.StringValue("MANUAL")}
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

func TestUserLifecycleAndClearing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := newUserBackend()
	r := testResource(t, b)
	model := testModel()
	createModel := model
	createModel.ID = types.StringUnknown()
	createModel.DisplayName = types.StringUnknown()
	createModel.UserStatus = types.StringUnknown()
	createModel.UserImportType = types.StringUnknown()
	created := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlan(t, createModel)}, &created)
	require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
	require.Equal(t, []string{"createUser", "disableUser"}, b.calls)
	var actual resourceModel
	require.False(t, created.State.Get(ctx, &actual).HasError())
	require.Equal(t, "1001", actual.ID.ValueString())
	require.False(t, actual.Enabled.ValueBool())
	read := resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	model.FirstName = types.StringValue("Updated")
	model.Enabled = types.BoolValue(true)
	model.Department = types.StringValue("")
	model.PhoneNumber = types.StringValue("")
	model.JobTitle = types.StringValue("")
	updated := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: testPlan(t, model)}, &updated)
	require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
	lastUpdate := b.variables[len(b.variables)-2]["updateUserInput"].(map[string]any)
	for _, name := range []string{"department", "jobTitle", "phoneNumber"} {
		value, exists := lastUpdate[name]
		require.True(t, exists, "missing explicit null for %s", name)
		require.Nil(t, value)
	}
	require.False(t, updated.State.Get(ctx, &actual).HasError())
	require.True(t, actual.Enabled.ValueBool())
	require.Equal(t, "", actual.Department.ValueString())
	require.Equal(t, "", actual.PhoneNumber.ValueString())
	deleted := resource.DeleteResponse{State: updated.State}
	r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError(), "%v", deleted.Diagnostics)
	require.True(t, deleted.State.Raw.IsNull())
	missing := resource.ReadResponse{State: updated.State}
	r.Read(ctx, resource.ReadRequest{State: updated.State}, &missing)
	require.False(t, missing.Diagnostics.HasError())
	require.True(t, missing.State.Raw.IsNull())
}

func TestUserCreateFailureRetainsIdentity(t *testing.T) {
	t.Parallel()
	b := newUserBackend()
	b.fail = "disableUser"
	r := testResource(t, b)
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, testModel())}, &response)
	require.True(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
}

func TestUserFailurePaths(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"createUser", "updateUser", "userList", "deleteUser", "enableUser"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			b := newUserBackend()
			b.object["userStatus"] = "DISABLED"
			b.fail = operation
			r := testResource(t, b)
			model := testModel()
			state := testState(t, model)
			switch operation {
			case "createUser":
				response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
				r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
				require.True(t, response.Diagnostics.HasError())
			case "userList":
				response := resource.ReadResponse{State: state}
				r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
				require.False(t, response.State.Raw.IsNull())
			case "deleteUser":
				response := resource.DeleteResponse{State: state}
				r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
				require.False(t, response.State.Raw.IsNull())
			default:
				model.Enabled = types.BoolValue(true)
				response := resource.UpdateResponse{State: state}
				r.Update(context.Background(), resource.UpdateRequest{State: state, Plan: testPlan(t, model)}, &response)
				require.True(t, response.Diagnostics.HasError())
				require.True(t, response.State.Raw.Equal(state.Raw))
			}
		})
	}
}

func TestUserDirectoryOwnership(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"read", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			b := newUserBackend()
			b.object["userImportType"] = "LDAP"
			r := testResource(t, b)
			state := testState(t, testModel())
			switch operation {
			case "read":
				response := resource.ReadResponse{State: state}
				r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
			case "update":
				response := resource.UpdateResponse{State: state}
				r.Update(context.Background(), resource.UpdateRequest{State: state, Plan: testPlan(t, testModel())}, &response)
				require.True(t, response.Diagnostics.HasError())
			case "delete":
				response := resource.DeleteResponse{State: state}
				r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
				require.True(t, response.Diagnostics.HasError())
			}
			require.Equal(t, []string{"userList"}, b.calls)
		})
	}
}

func TestUserPaginationAndLockedState(t *testing.T) {
	t.Parallel()
	b := newUserBackend()
	b.paged = true
	b.object["userStatus"] = "LOCKED"
	r := testResource(t, b)
	state := testState(t, testModel())
	response := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	var actual resourceModel
	require.False(t, response.State.Get(context.Background(), &actual).HasError())
	require.True(t, actual.Enabled.ValueBool())
	require.Equal(t, "LOCKED", actual.UserStatus.ValueString())
	require.Len(t, b.calls, 2)
}

func TestUserMalformedReadDoesNotRemoveState(t *testing.T) {
	t.Parallel()
	b := newUserBackend()
	b.malformed = true
	r := testResource(t, b)
	state := testState(t, testModel())
	response := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.True(t, response.State.Raw.Equal(state.Raw))
}

func TestUserValidationBeforeAPI(t *testing.T) {
	t.Parallel()
	models := []resourceModel{testModel(), testModel(), testModel(), testModel(), testModel()}
	models[0].Email = types.StringUnknown()
	models[1].FirstName = types.StringNull()
	models[2].Enabled = types.BoolUnknown()
	models[3].Email = types.StringValue("invalid")
	models[4].PhoneNumber = types.StringValue("invalid")
	for _, model := range models {
		b := newUserBackend()
		r := testResource(t, b)
		response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
		r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
		require.True(t, response.Diagnostics.HasError())
		require.Empty(t, b.calls)
	}
}

func TestUserImportAndConfiguration(t *testing.T) {
	t.Parallel()
	r := NewResource().(*userResource)
	ctx := context.Background()
	var metadata resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_user", metadata.TypeName)
	var configured resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &configured)
	require.True(t, configured.Diagnostics.HasError())
	imported := resource.ImportStateResponse{State: testState(t, testModel())}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "1001"}, &imported)
	require.False(t, imported.Diagnostics.HasError())
	var id types.String
	require.False(t, imported.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
}

func TestUserInjectedClient(t *testing.T) {
	t.Parallel()
	api := mocks.NewUserClient(t)
	var result cato.UserUserList
	body, err := json.Marshal(map[string]any{"user": map[string]any{"userList": map[string]any{"items": []any{newUserBackend().object}, "pageInfo": map[string]any{"total": 1}}}})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &result))
	api.EXPECT().UserUserList(mock.Anything, "account-123", mock.MatchedBy(func(input cato_models.UserListInput) bool {
		return input.Filter != nil && len(input.Filter.ID) == 1 && input.Filter.ID[0].Eq != nil && *input.Filter.ID[0].Eq == "1001" && input.Paging != nil && input.Paging.From == 0 && input.Paging.Limit == userPageSize
	})).Return(&result, nil).Once()
	r := &userResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	state := testState(t, testModel())
	response := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
}

func TestLockedUserDoesNotUnlockOnUpdate(t *testing.T) {
	t.Parallel()
	b := newUserBackend()
	b.object["userStatus"] = "LOCKED"
	r := testResource(t, b)
	model := testModel()
	model.Enabled = types.BoolValue(true)
	state := testState(t, model)
	response := resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{State: state, Plan: testPlan(t, model)}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	require.Equal(t, []string{"userList", "updateUser"}, b.calls)
}

func TestEmptyUserOptionalFieldsAreOmittedOnCreate(t *testing.T) {
	t.Parallel()
	b := newUserBackend()
	r := testResource(t, b)
	model := testModel()
	model.Department = types.StringValue("")
	model.JobTitle = types.StringValue("")
	model.PhoneNumber = types.StringValue("")
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	input := b.variables[0]["createUserInput"].(map[string]any)
	for _, field := range []string{"department", "jobTitle", "phoneNumber"} {
		require.NotContains(t, input, field)
	}
}

func TestPartialCreateErrorRetainsIdentity(t *testing.T) {
	t.Parallel()
	api := mocks.NewUserClient(t)
	var result cato.UserCreateUser
	body, err := json.Marshal(map[string]any{"user": map[string]any{"createUser": map[string]any{"user": newUserBackend().object}}})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &result))
	api.EXPECT().UserCreateUser(mock.Anything, "account-123", mock.Anything).Return(&result, errors.New("partial response")).Once()
	r := &userResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	model := testModel()
	model.ID = types.StringUnknown()
	model.UserStatus = types.StringUnknown()
	model.UserImportType = types.StringUnknown()
	model.DisplayName = types.StringUnknown()
	response := resource.CreateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: testPlan(t, model)}, &response)
	require.True(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
	require.True(t, response.State.Raw.IsFullyKnown())
}

func userListPayload(t *testing.T, b *userBackend, input map[string]any) map[string]any {
	t.Helper()
	require.Equal(t, map[string]any{"id": []any{map[string]any{"eq": "1001"}}}, input["filter"])
	items := []any{}
	total := 0
	if b.object != nil {
		items = []any{b.object}
		total = 1
	}
	if b.paged {
		total = 2
		if input["paging"].(map[string]any)["from"] == float64(0) {
			items = []any{map[string]any{"id": "1000"}}
		}
	}
	return map[string]any{"items": items, "pageInfo": map[string]any{"total": total}}
}
