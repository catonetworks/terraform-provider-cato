package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

const roleTestJSON = `{"id":"42","name":"Role","description":"","predefined":false,"isUsedOnExternalAccess":true,"accountType":"REGULAR","permission":[{"resource":"Sites","action":"VIEW"},{"resource":"Administrators","action":"EDIT"}]}`

func roleTestSDK(t *testing.T, handler func(string, map[string]json.RawMessage) string) RoleManagementClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query     string                     `json:"query"`
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(handler(request.Query, request.Variables)))
	}))
	t.Cleanup(server.Close)
	return cato.NewClient(server.Client(), server.URL, nil)
}
func roleEnvelope(leaf, payload string) string {
	marker := ""
	if leaf == "role" {
		marker = `"__typename":"RoleManagementQueries",`
	}
	return `{"data":{"rbac":{"roleManagement":{` + marker + `"` + leaf + `":` + payload + `}}}}`
}
func roleTestModel(t *testing.T) Role {
	t.Helper()
	value := emptyRole()
	value.ID = types.StringValue("123:42")
	value.AccountID = types.StringValue("123")
	value.RoleID = types.StringValue("42")
	value.Name = types.StringValue("Role")
	value.Description = types.StringValue("")
	value.Predefined = types.BoolValue(false)
	value.AccountType = types.StringValue("REGULAR")
	value.IsUsedOnExternalAccess = types.BoolValue(true)
	value.Permissions = types.SetValueMust(rolePermissionObjectType, []attr.Value{
		types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Administrators"), "action": types.StringValue("EDIT")}),
		types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Sites"), "action": types.StringValue("VIEW")}),
	})
	return value
}
func roleTestState(t *testing.T, r *roleResource, model Role) tfsdk.State {
	t.Helper()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	state := tfsdk.State{Schema: resp.Schema}
	require.False(t, state.Set(context.Background(), model).HasError())
	return state
}
func roleTestPlan(t *testing.T, r *roleResource, model Role) tfsdk.Plan {
	t.Helper()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	plan := tfsdk.Plan{Schema: resp.Schema}
	require.False(t, plan.Set(context.Background(), model).HasError())
	return plan
}

func TestRoleCreateMockPayload(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewRoleManagementClient(t)
	var result cato.RbacRoleManagementCreateRole
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"roleManagement":{"createRole":{"role":`+roleTestJSON+`}}}}`), &result))
	client.EXPECT().RbacRoleManagementCreateRole(mock.Anything, "456", mock.MatchedBy(func(input cato_models.RoleManagementCreateRoleInput) bool {
		require.Equal(t, "Role", input.Name)
		require.NotNil(t, input.Description)
		require.Empty(t, *input.Description)
		require.Len(t, input.Permission, 2)
		return true
	})).Return(&result, nil).Once()
	r := &roleResource{roleClientConfig: roleClientConfig{client: &catoClientData{AccountId: "123"}, roleClient: client}}
	plan := roleTestModel(t)
	plan.AccountID = types.StringValue("456")
	resp := resource.CreateResponse{State: roleTestState(t, r, emptyRole())}
	r.Create(ctx, resource.CreateRequest{Plan: roleTestPlan(t, r, plan)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	var got Role
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "456:42", got.ID.ValueString())
	require.True(t, got.IsUsedOnExternalAccess.ValueBool())
	require.True(t, got.Permissions.Equal(plan.Permissions))
}

func TestRoleLifecycle(t *testing.T) {
	ctx := context.Background()
	var calls []string
	client := roleTestSDK(t, func(query string, variables map[string]json.RawMessage) string {
		require.JSONEq(t, `"123"`, string(variables["accountId"]))
		switch {
		case strings.Contains(query, "rbacRoleManagementCreateRole"):
			calls = append(calls, "create")
			return roleEnvelope("createRole", `{"role":`+roleTestJSON+`}`)
		case strings.Contains(query, "rbacRoleManagementUpdateRole"):
			calls = append(calls, "update")
			var input cato_models.RoleManagementUpdateRoleInput
			require.NoError(t, json.Unmarshal(variables["input"], &input))
			require.Equal(t, "42", input.ID)
			require.Equal(t, "Updated", input.Name)
			require.NotNil(t, input.Description)
			require.Empty(t, *input.Description)
			require.Len(t, input.Permission, 1)
			require.Equal(t, "Sites", input.Permission[0].Resource)
			updated := strings.ReplaceAll(roleTestJSON, `"Role"`, `"Updated"`)
			updated = strings.ReplaceAll(updated, `,{"resource":"Administrators","action":"EDIT"}`, "")
			return roleEnvelope("updateRole", `{"role":`+updated+`}`)
		case strings.Contains(query, "rbacRoleManagementDeleteRole"):
			calls = append(calls, "delete")
			require.JSONEq(t, `{"id":"42"}`, string(variables["input"]))
			return roleEnvelope("deleteRole", `{"id":"42"}`)
		default:
			calls = append(calls, "read")
			return roleEnvelope("role", roleTestJSON)
		}
	})
	r := &roleResource{roleClientConfig: roleClientConfig{client: &catoClientData{AccountId: "123"}, roleClient: client}}
	model := roleTestModel(t)
	model.AccountID = types.StringNull()
	create := resource.CreateResponse{State: roleTestState(t, r, emptyRole())}
	r.Create(ctx, resource.CreateRequest{Plan: roleTestPlan(t, r, model)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	var got Role
	require.False(t, read.State.Get(ctx, &got).HasError())
	require.True(t, got.Permissions.Equal(model.Permissions))
	got.Name = types.StringValue("Updated")
	got.Permissions = types.SetValueMust(rolePermissionObjectType, []attr.Value{types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Sites"), "action": types.StringValue("VIEW")})})
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{Plan: roleTestPlan(t, r, got), State: read.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.False(t, update.State.Get(ctx, &got).HasError())
	require.Equal(t, "Updated", got.Name.ValueString())
	require.Len(t, got.Permissions.Elements(), 1)
	deletion := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &deletion)
	require.False(t, deletion.Diagnostics.HasError(), deletion.Diagnostics)
	require.Equal(t, []string{"create", "read", "update", "read", "delete"}, calls)
}

func TestRoleMissingAndPredefined(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name, payload      string
		wantError, removed bool
	}{{"missing", "null", false, true}, {"predefined", strings.ReplaceAll(roleTestJSON, `"predefined":false`, `"predefined":true`), true, false}} {
		t.Run(test.name, func(t *testing.T) {
			client := roleTestSDK(t, func(query string, _ map[string]json.RawMessage) string {
				require.NotContains(t, query, "rbacRoleManagementDeleteRole")
				return roleEnvelope("role", test.payload)
			})
			r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
			state := roleTestState(t, r, roleTestModel(t))
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			require.Equal(t, test.wantError, read.Diagnostics.HasError(), read.Diagnostics)
			require.Equal(t, test.removed, read.State.Raw.IsNull())
			deletion := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &deletion)
			require.Equal(t, test.wantError, deletion.Diagnostics.HasError(), deletion.Diagnostics)
		})
	}
}

func TestRoleErrorsAndNullPayloads(t *testing.T) {
	ctx := context.Background()
	for _, body := range []string{`{"errors":[{"message":"Permission denied"}],"data":null}`, `{"data":{"rbac":{"roleManagement":{}}}}`, `{"data":{"rbac":null}}`} {
		t.Run(body, func(t *testing.T) {
			client := roleTestSDK(t, func(_ string, _ map[string]json.RawMessage) string { return body })
			r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
			state := roleTestState(t, r, roleTestModel(t))
			plan := roleTestPlan(t, r, roleTestModel(t))
			create := resource.CreateResponse{State: state}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
			require.True(t, create.Diagnostics.HasError())
			update := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &update)
			require.True(t, update.Diagnostics.HasError())
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			require.True(t, read.Diagnostics.HasError())
			require.False(t, read.State.Raw.IsNull())
		})
	}
	client := mocks.NewRoleManagementClient(t)
	client.EXPECT().RbacRoleManagementRole(mock.Anything, "123", "42").Return(nil, errors.New("unavailable")).Once()
	r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
	state := roleTestState(t, r, roleTestModel(t))
	deletion := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deletion)
	require.True(t, deletion.Diagnostics.HasError())
	require.False(t, deletion.State.Raw.IsNull())
}

func rolesTestConfig(t *testing.T, d *rolesDataSource, filter types.Object) tfsdk.Config {
	t.Helper()
	resp := datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	config := tfsdk.Config{Schema: resp.Schema}
	state := tfsdk.State{Schema: resp.Schema}
	require.False(t, state.Set(context.Background(), RoleList{ID: types.StringNull(), AccountID: types.StringNull(), Filter: filter, SortDirection: types.StringNull(), Items: types.ListNull(roleObjectType), Total: types.Int64Null()}).HasError())
	config.Raw = state.Raw
	return config
}

func TestRolesPagination(t *testing.T) {
	ctx := context.Background()
	for _, scenario := range []string{"multiple", "empty", "no-progress", "repeated", "api-error", "null-payload"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			client := roleTestSDK(t, func(_ string, variables map[string]json.RawMessage) string {
				calls++
				var input cato_models.RoleManagementRoleListInput
				require.NoError(t, json.Unmarshal(variables["input"], &input))
				require.Equal(t, int64(rolePageSize), input.Paging.Limit)
				require.Equal(t, cato_models.SortOrderAsc, input.Sort.Name.Direction)
				if scenario == "api-error" {
					return `{"errors":[{"message":"Failed"}],"data":null}`
				}
				if scenario == "null-payload" {
					return roleEnvelope("roleList", "null")
				}
				if scenario == "empty" {
					return roleEnvelope("roleList", `{"items":[],"paging":{"total":0}}`)
				}
				if scenario == "no-progress" {
					return roleEnvelope("roleList", `{"items":[],"paging":{"total":1}}`)
				}
				if scenario == "repeated" {
					return roleEnvelope("roleList", `{"items":[`+roleTestJSON+`],"paging":{"total":2}}`)
				}
				items := make([]string, 0)
				count := rolePageSize
				if calls == 2 {
					count = 1
				}
				require.Equal(t, int64((calls-1)*rolePageSize), input.Paging.From)
				for i := 0; i < count; i++ {
					items = append(items, strings.ReplaceAll(roleTestJSON, `"42"`, fmt.Sprintf(`"%d"`, (calls-1)*rolePageSize+i+1)))
				}
				return roleEnvelope("roleList", `{"items":[`+strings.Join(items, ",")+`],"paging":{"total":101}}`)
			})
			d := &rolesDataSource{roleClientConfig: roleClientConfig{client: &catoClientData{AccountId: "123"}, roleClient: client}}
			config := rolesTestConfig(t, d, types.ObjectNull(roleFilterAttrTypes))
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: config.Schema}}
			d.Read(ctx, datasource.ReadRequest{Config: config}, &resp)
			if scenario != "multiple" && scenario != "empty" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var got RoleList
			require.False(t, resp.State.Get(ctx, &got).HasError())
			require.Equal(t, "ASC", got.SortDirection.ValueString())
			require.Equal(t, "123", got.AccountID.ValueString())
			if scenario == "empty" {
				require.Empty(t, got.Items.Elements())
				require.False(t, got.Items.IsNull())
				require.Zero(t, got.Total.ValueInt64())
				require.Equal(t, 1, calls)
			} else {
				require.Len(t, got.Items.Elements(), 101)
				require.Equal(t, int64(101), got.Total.ValueInt64())
				require.Equal(t, 2, calls)
			}
		})
	}
}

func TestRoleFilterForwarding(t *testing.T) {
	ctx := context.Background()
	operators := map[string]attr.Value{"eq": types.StringValue("42"), "neq": types.StringValue("43"), "in": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("42")}), "nin": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("43")})}
	filter := types.ObjectValueMust(roleFilterAttrTypes, map[string]attr.Value{"id": types.ObjectValueMust(roleStringFilterAttrTypes, operators), "name": types.ObjectValueMust(roleStringFilterAttrTypes, operators), "predefined": types.ObjectValueMust(roleBooleanFilterAttrTypes, map[string]attr.Value{"eq": types.BoolValue(false), "neq": types.BoolValue(true)})})
	input, diags := roleListInput(ctx, filter, types.StringValue("DESC"))
	require.False(t, diags.HasError())
	require.Equal(t, cato_models.SortOrderDesc, input.Sort.Name.Direction)
	require.Equal(t, "42", *input.Filter.ID.Eq)
	require.Equal(t, "43", *input.Filter.ID.Neq)
	require.Equal(t, []string{"42"}, input.Filter.ID.In)
	require.Equal(t, []string{"43"}, input.Filter.ID.Nin)
	require.Equal(t, "42", *input.Filter.Name.Eq)
	require.False(t, *input.Filter.Predefined.Eq)
	require.True(t, *input.Filter.Predefined.Neq)
	_, diags = roleListInput(ctx, types.ObjectUnknown(roleFilterAttrTypes), types.StringValue("ASC"))
	require.True(t, diags.HasError())
	_, diags = roleListInput(ctx, types.ObjectNull(roleFilterAttrTypes), types.StringValue("INVALID"))
	require.True(t, diags.HasError())
}

func TestRoleDataSources(t *testing.T) {
	ctx := context.Background()
	for _, scenario := range []string{"custom", "predefined", "missing", "error", "null-namespace"} {
		t.Run("role/"+scenario, func(t *testing.T) {
			client := roleTestSDK(t, func(_ string, v map[string]json.RawMessage) string {
				require.JSONEq(t, `"456"`, string(v["accountId"]))
				require.JSONEq(t, `"42"`, string(v["id"]))
				switch scenario {
				case "missing":
					return roleEnvelope("role", "null")
				case "error":
					return `{"errors":[{"message":"Failed"}],"data":null}`
				case "null-namespace":
					return `{"data":{"rbac":null}}`
				case "predefined":
					return roleEnvelope("role", strings.ReplaceAll(roleTestJSON, `"predefined":false`, `"predefined":true`))
				default:
					return roleEnvelope("role", roleTestJSON)
				}
			})
			d := &roleDataSource{roleClientConfig: roleClientConfig{client: &catoClientData{AccountId: "123"}, roleClient: client}}
			schemaResp := datasource.SchemaResponse{}
			d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
			configModel := emptyRole()
			configModel.AccountID = types.StringValue("456")
			configModel.RoleID = types.StringValue("42")
			state := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, state.Set(ctx, configModel).HasError())
			resp := datasource.ReadResponse{State: state}
			d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config(state)}, &resp)
			if scenario != "custom" && scenario != "predefined" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var got Role
			require.False(t, resp.State.Get(ctx, &got).HasError())
			require.Equal(t, "456:42", got.ID.ValueString())
			require.Equal(t, scenario == "predefined", got.Predefined.ValueBool())
		})
	}
	for _, scenario := range []string{"success", "empty", "null", "error"} {
		t.Run("catalog/"+scenario, func(t *testing.T) {
			client := roleTestSDK(t, func(_ string, v map[string]json.RawMessage) string {
				require.JSONEq(t, `"123"`, string(v["accountId"]))
				switch scenario {
				case "empty":
					return roleEnvelope("permissionCatalog", `{"resource":[]}`)
				case "null":
					return roleEnvelope("permissionCatalog", "null")
				case "error":
					return `{"errors":[{"message":"Failed"}],"data":null}`
				default:
					return roleEnvelope("permissionCatalog", `{"resource":[{"resource":"Sites","supportedAction":["VIEW","EDIT"]}]}`)
				}
			})
			d := &permissionCatalogDataSource{roleClientConfig: roleClientConfig{client: &catoClientData{AccountId: "123"}, roleClient: client}}
			schemaResp := datasource.SchemaResponse{}
			d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, state.Set(ctx, PermissionCatalog{ID: types.StringNull(), AccountID: types.StringNull(), Resources: types.SetNull(roleCatalogObjectType)}).HasError())
			resp := datasource.ReadResponse{State: state}
			d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config(state)}, &resp)
			if scenario == "null" || scenario == "error" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var got PermissionCatalog
			require.False(t, resp.State.Get(ctx, &got).HasError())
			require.False(t, got.Resources.IsNull())
			if scenario == "empty" {
				require.Empty(t, got.Resources.Elements())
			} else {
				require.Len(t, got.Resources.Elements(), 1)
				values := got.Resources.Elements()[0].(types.Object).Attributes()
				require.Equal(t, "Sites", values["resource"].(types.String).ValueString())
				require.Len(t, values["supported_actions"].(types.Set).Elements(), 2)
			}
		})
	}
}

func TestRoleValidationAndProtection(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewRoleManagementClient(t)
	r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
	for _, permissions := range []types.Set{
		types.SetUnknown(rolePermissionObjectType),
		types.SetNull(rolePermissionObjectType),
		types.SetValueMust(rolePermissionObjectType, []attr.Value{
			types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Sites"), "action": types.StringValue("VIEW")}),
			types.ObjectValueMust(rolePermissionAttrTypes, map[string]attr.Value{"resource": types.StringValue("Sites"), "action": types.StringValue("EDIT")}),
		}),
	} {
		plan := roleTestModel(t)
		plan.Permissions = permissions
		resp := resource.CreateResponse{State: roleTestState(t, r, emptyRole())}
		r.Create(ctx, resource.CreateRequest{Plan: roleTestPlan(t, r, plan)}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	}
	stateModel := roleTestModel(t)
	stateModel.Predefined = types.BoolValue(true)
	state := roleTestState(t, r, stateModel)
	update := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: roleTestPlan(t, r, stateModel), State: state}, &update)
	require.True(t, update.Diagnostics.HasError())
	deletion := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deletion)
	require.True(t, deletion.Diagnostics.HasError())
	input, diags := rolePermissionsInput(ctx, types.SetValueMust(rolePermissionObjectType, []attr.Value{}))
	require.False(t, diags.HasError())
	require.NotNil(t, input)
	require.Empty(t, input)
}

func TestRoleConfigurationAndMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewRoleResource().(*roleResource)
	metadata := resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_role", metadata.TypeName)
	configure := resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{}, &configure)
	require.False(t, configure.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &configure)
	require.True(t, configure.Diagnostics.HasError())
	configure = resource.ConfigureResponse{}
	data := &catoClientData{AccountId: "123"}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: data}, &configure)
	require.False(t, configure.Diagnostics.HasError())
	require.Same(t, data, r.client)
	for _, test := range []struct {
		name   string
		source datasource.DataSource
	}{{"cato_role", NewRoleDataSource()}, {"cato_roles", NewRolesDataSource()}, {"cato_permission_catalog", NewPermissionCatalogDataSource()}} {
		response := datasource.MetadataResponse{}
		test.source.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "cato"}, &response)
		require.Equal(t, test.name, response.TypeName)
		dsConfigure := datasource.ConfigureResponse{}
		test.source.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &dsConfigure)
		require.False(t, dsConfigure.Diagnostics.HasError())
	}
	account, err := effectiveRoleAccount(types.StringNull(), data)
	require.NoError(t, err)
	require.Equal(t, "123", account)
	account, err = effectiveRoleAccount(types.StringValue("456"), data)
	require.NoError(t, err)
	require.Equal(t, "456", account)
	_, err = effectiveRoleAccount(types.StringNull(), nil)
	require.Error(t, err)
}

func TestRoleDeleteMutationFailures(t *testing.T) {
	ctx := context.Background()
	for _, payload := range []string{`null`, `{"id":"different"}`, `{"id":"42"}`} {
		t.Run(payload, func(t *testing.T) {
			client := roleTestSDK(t, func(query string, _ map[string]json.RawMessage) string {
				if strings.Contains(query, "rbacRoleManagementDeleteRole") {
					return roleEnvelope("deleteRole", payload)
				}
				return roleEnvelope("role", roleTestJSON)
			})
			r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
			state := roleTestState(t, r, roleTestModel(t))
			resp := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
			require.Equal(t, payload != `{"id":"42"}`, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
	client := roleTestSDK(t, func(query string, _ map[string]json.RawMessage) string {
		if strings.Contains(query, "rbacRoleManagementDeleteRole") {
			return `{"errors":[{"message":"Role is in use"}],"data":null}`
		}
		return roleEnvelope("role", roleTestJSON)
	})
	r := &roleResource{roleClientConfig: roleClientConfig{roleClient: client}}
	state := roleTestState(t, r, roleTestModel(t))
	resp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull())
}

func TestRoleAccountChangeRequiresReplacement(t *testing.T) {
	ctx := context.Background()
	r := &roleResource{}
	schemaResp := resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	stateModel := roleTestModel(t)
	state := roleTestState(t, r, stateModel)
	for _, accountID := range []string{"123", "456"} {
		t.Run(accountID, func(t *testing.T) {
			planModel := stateModel
			planModel.AccountID = types.StringValue(accountID)
			plan := roleTestPlan(t, r, planModel)
			request := planmodifier.StringRequest{State: state, Plan: plan, StateValue: stateModel.AccountID, PlanValue: planModel.AccountID, ConfigValue: planModel.AccountID}
			response := planmodifier.StringResponse{PlanValue: planModel.AccountID}
			schemaResp.Schema.Attributes["account_id"].(schema.StringAttribute).PlanModifiers[0].PlanModifyString(ctx, request, &response)
			require.False(t, response.Diagnostics.HasError())
			require.Equal(t, accountID != "123", response.RequiresReplace)
		})
	}
}
