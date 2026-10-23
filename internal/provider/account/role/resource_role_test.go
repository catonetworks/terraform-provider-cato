package role

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

func rolePlan(t *testing.T, r *roleResource) tfsdk.Plan {
	t.Helper()
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	value := Model{ID: types.StringNull(), AccountID: types.StringNull(), RoleID: types.StringNull(), Name: types.StringValue("viewer"), Description: types.StringValue(""), Permissions: types.SetValueMust(iam.PermissionObjectType, []attr.Value{}), Predefined: types.BoolNull(), ApplicableAccountType: types.SetNull(types.StringType), IsUsedOnExternalAccess: types.BoolNull()}
	require.False(t, plan.Set(context.Background(), value).HasError())
	return plan
}
func TestRoleCRUD(t *testing.T) {
	ctx := context.Background()
	phase := "create"
	role := `{"id":"r","name":"viewer","description":"","predefined":false,"applicableAccountType":["CUSTOMER"],"isUsedOnExternalAccess":false,"permission":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string                     `json:"query"`
			Variables map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Contains(t, []string{string(body.Variables["accountID"]), string(body.Variables["accountId"])}, `"123"`)
		w.Header().Set("Content-Type", "application/json")
		payload := ""
		switch {
		case strings.Contains(body.Query, "rbacRoleManagementReadRole"):
			if phase == "missing" {
				payload = `{"__typename":"RoleManagementQueries","role":null}`
			} else {
				payload = `{"__typename":"RoleManagementQueries","role":` + role + `}`
			}
		case strings.Contains(body.Query, "rbacRoleManagementCreateRole"):
			payload = `{"createRole":{"role":` + role + `}}`
		case strings.Contains(body.Query, "rbacRoleManagementUpdateRole"):
			var input map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(body.Variables["roleManagementUpdateRoleInput"], &input))
			require.Equal(t, `[]`, string(input["permission"]))
			require.Equal(t, `""`, string(input["description"]))
			payload = `{"updateRole":{"role":` + role + `}}`
		default:
			payload = `{"deleteRole":{"id":"r"}}`
		}
		namespace := `"roleManagement":` + payload
		if strings.Contains(body.Query, "rbacRoleManagementReadRole") {
			namespace = `"__typename":"RBACQueries",` + namespace
		}
		_, _ = w.Write([]byte(`{"data":{"rbac":{` + namespace + `}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
	require.NoError(t, err)
	r := &roleResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
	plan := rolePlan(t, r)
	created := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
	require.False(t, created.Diagnostics.HasError(), created.Diagnostics)
	var state Model
	require.False(t, created.State.Get(ctx, &state).HasError())
	require.Equal(t, "123:r", state.ID.ValueString())
	read := resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	updated := resource.UpdateResponse{State: created.State}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: created.State}, &updated)
	require.False(t, updated.Diagnostics.HasError(), updated.Diagnostics)
	deleted := resource.DeleteResponse{State: updated.State}
	r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError(), deleted.Diagnostics)
	phase = "missing"
	read = resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.True(t, read.State.Raw.IsNull())
	imported := resource.ImportStateResponse{State: created.State}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "123:r"}, &imported)
	require.False(t, imported.Diagnostics.HasError(), imported.Diagnostics)
}
func TestRoleCreateAPIError(t *testing.T) {
	client := mocks.NewRoleResourceClient(t)
	client.EXPECT().RbacRoleManagementCreateRole(mock.Anything, "123", mock.Anything).Return(nil, errors.New("denied")).Once()
	r := &roleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := rolePlan(t, r)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}
func TestRoleImmutableAndInvalidInput(t *testing.T) {
	require.Error(t, mutableRole(&iam.RoleRecord{Predefined: true}))
	require.Error(t, mutableRole(&iam.RoleRecord{External: true}))
	require.Error(t, mutableRole(nil))
	r := &roleResource{}
	plan := rolePlan(t, r)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	configure := resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "bad"}, &configure)
	require.True(t, configure.Diagnostics.HasError())
	imported := resource.ImportStateResponse{}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "bad"}, &imported)
	require.True(t, imported.Diagnostics.HasError())
}

func TestRoleValidationStopsBeforeAPI(t *testing.T) {
	client := mocks.NewRoleResourceClient(t)
	r := &roleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := rolePlan(t, r)
	var model Model
	require.False(t, plan.Get(context.Background(), &model).HasError())
	model.Permissions = types.SetUnknown(iam.PermissionObjectType)
	require.False(t, plan.Set(context.Background(), model).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
}
func TestRoleImmutableLifecycleGuards(t *testing.T) {
	for _, kind := range []string{"predefined", "external"} {
		t.Run(kind, func(t *testing.T) {
			client := mocks.NewRoleResourceClient(t)
			response := &sdk.RbacRoleManagementReadRole{}
			role := `{"id":"r","name":"viewer","permission":[],"applicableAccountType":[],"predefined":` + strconv.FormatBool(kind == "predefined") + `,"isUsedOnExternalAccess":` + strconv.FormatBool(kind == "external") + `}`
			require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":`+role+`}}}`), response))
			client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(response, nil).Twice()
			r := &roleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			plan := rolePlan(t, r)
			state := tfsdk.State(plan)
			require.False(t, state.SetAttribute(context.Background(), path.Root("id"), "123:r").HasError())
			updated := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &updated)
			require.True(t, updated.Diagnostics.HasError())
			deleted := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &deleted)
			require.True(t, deleted.Diagnostics.HasError())
		})
	}
}

func TestRoleCreateRetainsStateWhenReplicaFails(t *testing.T) {
	client := mocks.NewRoleResourceClient(t)
	created := &sdk.RbacRoleManagementCreateRole{}
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"roleManagement":{"createRole":{"role":{"id":"r","name":"viewer","permission":[],"applicableAccountType":[]}}}}}`), created))
	client.EXPECT().RbacRoleManagementCreateRole(mock.Anything, "123", mock.Anything).Return(created, nil).Once()
	client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(nil, errors.New("replica unavailable")).Once()
	r := &roleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := rolePlan(t, r)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Equal(t, 1, resp.Diagnostics.WarningsCount())
	require.Contains(t, resp.Diagnostics[0].Detail(), "replica unavailable")
	var state Model
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	require.Equal(t, "123:r", state.ID.ValueString())
}
func TestRoleReplicaConvergence(t *testing.T) {
	client := mocks.NewRoleResourceClient(t)
	missing := &sdk.RbacRoleManagementReadRole{}
	visible := &sdk.RbacRoleManagementReadRole{}
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":null}}}`), missing))
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":{"id":"r","name":"viewer","permission":[],"applicableAccountType":[]}}}}`), visible))
	client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(missing, nil).Once()
	client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(visible, nil).Once()
	err := waitForRole(context.Background(), client, "123", &iam.RoleRecord{ID: "r", Name: "viewer", Permissions: []iam.PermissionRecord{}})
	require.NoError(t, err)
}

func TestRoleUpdateRetainsStateWhenReplicaFails(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewRoleResourceClient(t)
	read := &sdk.RbacRoleManagementReadRole{}
	updated := &sdk.RbacRoleManagementUpdateRole{}
	role := `{"id":"r","name":"viewer","description":"updated","permission":[],"applicableAccountType":[]}`
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":`+role+`}}}`), read))
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"roleManagement":{"updateRole":{"role":`+role+`}}}}`), updated))
	client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(read, nil).Once()
	client.EXPECT().RbacRoleManagementUpdateRole(mock.Anything, "123", mock.Anything).Return(updated, nil).Once()
	client.EXPECT().RbacRoleManagementReadRole(mock.Anything, "123", "r").Return(nil, errors.New("replica unavailable")).Once()
	r := &roleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := rolePlan(t, r)
	require.False(t, plan.SetAttribute(ctx, path.Root("description"), "updated").HasError())
	state := tfsdk.State(plan)
	require.False(t, state.SetAttribute(ctx, path.Root("id"), "123:r").HasError())
	resp := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Equal(t, 1, resp.Diagnostics.WarningsCount())
	var value Model
	require.False(t, resp.State.Get(ctx, &value).HasError())
	require.Equal(t, "123:r", value.ID.ValueString())
	require.Equal(t, "updated", value.Description.ValueString())
}
