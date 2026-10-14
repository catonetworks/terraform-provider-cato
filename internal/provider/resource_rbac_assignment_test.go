package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	rbacapi "github.com/catonetworks/terraform-provider-cato/internal/provider/rbacassignmentapi"
)

func assignmentFixture() RBACAssignment {
	return RBACAssignment{ID: types.StringNull(), AssignmentID: types.StringNull(), AccountID: types.StringValue("123"), UserID: types.StringValue("7"), GroupID: types.StringNull(), RoleIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("42")}), AllowedEntities: types.SetValueMust(assignmentEntityObjectType, []attr.Value{}), AllowAllSites: types.StringValue("NONE"), AllowAllUsers: types.StringValue("NONE"), AllowViewingUnobfuscatedData: types.BoolValue(false), Description: types.StringValue(""), AssignmentType: types.StringNull(), DataObfuscationStatus: types.StringNull()}
}

const assignmentJSON = `{"id":"a1","account":{"id":"123"},"user":{"id":"7"},"group":null,"assignmentType":"ADMIN_USER","roles":[{"id":"42"}],"allowedEntities":[],"allowAllSites":"NONE","allowAllUsers":"NONE","dataObfuscationStatus":"OBFUSCATED","description":""}`

func assignmentTestResource(t *testing.T, handler http.HandlerFunc) (*rbacAssignmentResource, tfsdk.Plan, tfsdk.State) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := cato.New(server.URL, "test-token", "123", server.Client(), nil)
	require.NoError(t, err)
	r := &rbacAssignmentResource{client: &catoClientData{AccountId: "123", catov2: &providerSDKClient{Client: client}}}
	var sr resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &sr)
	p := tfsdk.Plan{Schema: sr.Schema}
	s := tfsdk.State{Schema: sr.Schema}
	require.False(t, p.Set(context.Background(), assignmentFixture()).HasError())
	require.False(t, s.Set(context.Background(), assignmentFixture()).HasError())
	return r, p, s
}

func TestRBACAssignmentLifecycle(t *testing.T) {
	ctx := context.Background()
	calls := 0
	r, plan, state := assignmentTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		calls++
		var body struct {
			OperationName string                     `json:"operationName"`
			Variables     map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.JSONEq(t, `"123"`, string(body.Variables["accountId"]))
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "addRbacAssignments":
			require.JSONEq(t, `{"users":[{"by":"ID","input":"7"}],"groups":[],"roles":[{"by":"ID","input":"42"}],"allowedEntities":[],"allowAllSites":"NONE","allowAllUsers":"NONE","allowViewingUnobfuscatedData":false,"description":""}`, string(body.Variables["input"]))
			_, _ = w.Write([]byte(`{"data":{"rbac":{"addRbacAssignments":{"items":[` + assignmentJSON + `]}}}}`))
		case "updateRbacAssignment":
			require.JSONEq(t, `{"rbacAssignmentId":"a1","roles":[{"by":"ID","input":"42"}],"allowedEntities":[],"allowAllSites":"NONE","allowAllUsers":"NONE","allowViewingUnobfuscatedData":false,"description":""}`, string(body.Variables["input"]))
			_, _ = w.Write([]byte(`{"data":{"rbac":{"updateRbacAssignment":{"assignment":` + assignmentJSON + `}}}}`))
		case "rbacAssignmentList":
			_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[` + assignmentJSON + `],"paging":{"total":1}}}}}`))
		case "removeRbacAssignment":
			require.JSONEq(t, `{"rbacAssignmentId":"a1"}`, string(body.Variables["input"]))
			_, _ = w.Write([]byte(`{"data":{"rbac":{"removeRbacAssignment":{"rbacAssignmentId":"a1"}}}}`))
		default:
			t.Errorf("unexpected operation %s", body.OperationName)
		}
	})
	cr := resource.CreateResponse{State: state}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &cr)
	require.False(t, cr.Diagnostics.HasError(), cr.Diagnostics)
	var got RBACAssignment
	require.False(t, cr.State.Get(ctx, &got).HasError())
	require.Equal(t, "123:a1", got.ID.ValueString())
	rr := resource.ReadResponse{State: cr.State}
	r.Read(ctx, resource.ReadRequest{State: cr.State}, &rr)
	require.False(t, rr.Diagnostics.HasError(), rr.Diagnostics)
	require.False(t, plan.Set(ctx, got).HasError())
	ur := resource.UpdateResponse{State: rr.State}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: rr.State}, &ur)
	require.False(t, ur.Diagnostics.HasError(), ur.Diagnostics)
	dr := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: ur.State}, &dr)
	require.False(t, dr.Diagnostics.HasError(), dr.Diagnostics)
	require.Equal(t, 5, calls)
}

func TestRBACAssignmentReadAndDeleteMissing(t *testing.T) {
	ctx := context.Background()
	r, _, state := assignmentTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[],"paging":{"total":0}}}}}`))
	})
	model := assignmentFixture()
	model.AssignmentID = types.StringValue("gone")
	require.False(t, state.Set(ctx, model).HasError())
	rr := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &rr)
	require.False(t, rr.Diagnostics.HasError(), rr.Diagnostics)
	require.True(t, rr.State.Raw.IsNull())
	dr := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &dr)
	require.False(t, dr.Diagnostics.HasError(), dr.Diagnostics)
}

func TestRBACAssignmentPagination(t *testing.T) {
	ctx := context.Background()
	calls := 0
	r, _, _ := assignmentTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables struct {
				Input struct {
					Paging struct {
						From int `json:"from"`
					} `json:"paging"`
				} `json:"input"`
			} `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, calls, body.Variables.Input.Paging.From)
		calls++
		item := assignmentJSON
		if calls == 2 {
			item = strings.ReplaceAll(item, `"a1"`, `"a2"`)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[` + item + `],"paging":{"total":2}}}}}`))
	})
	value, err := lookupAssignment(ctx, r.getAssignmentClient(), "123", "a2")
	require.NoError(t, err)
	require.Equal(t, "a2", value.ID)
	require.Equal(t, 2, calls)
}

func TestRBACAssignmentPaginationErrors(t *testing.T) {
	cases := map[string]string{
		"null list":      `null`,
		"missing paging": `{"items":[]}`,
		"negative total": `{"items":[],"paging":{"total":-1}}`,
		"no progress":    `{"items":[],"paging":{"total":2}}`,
		"repeat page":    `{"items":[` + assignmentJSON + `],"paging":{"total":2}}`,
		"cross account":  `{"items":[` + strings.ReplaceAll(assignmentJSON, `"123"`, `"456"`) + `],"paging":{"total":1}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			r, _, _ := assignmentTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":` + payload + `}}}`))
			})
			_, err := lookupAssignment(context.Background(), r.getAssignmentClient(), "123", "absent")
			require.Error(t, err)
		})
	}
}

func TestRBACAssignmentAPIErrors(t *testing.T) {
	for _, operation := range []string{"create", "read", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			r, plan, state := assignmentTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"backend rejected request"}]}`))
			})
			model := assignmentFixture()
			model.AssignmentID = types.StringValue("a1")
			model.AssignmentType = types.StringValue("ADMIN_USER")
			require.False(t, state.Set(context.Background(), model).HasError())
			switch operation {
			case "create":
				resp := resource.CreateResponse{State: state}
				r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "read":
				resp := resource.ReadResponse{State: state}
				r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			case "update":
				resp := resource.UpdateResponse{State: state}
				r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "delete":
				resp := resource.DeleteResponse{}
				r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			}
		})
	}
}

func TestRBACAssignmentInputValidation(t *testing.T) {
	cases := map[string]func(*RBACAssignment){
		"both principals":   func(v *RBACAssignment) { v.GroupID = types.StringValue("8") },
		"no principal":      func(v *RBACAssignment) { v.UserID = types.StringNull() },
		"unknown principal": func(v *RBACAssignment) { v.UserID = types.StringUnknown() },
		"empty roles":       func(v *RBACAssignment) { v.RoleIDs = types.SetValueMust(types.StringType, []attr.Value{}) },
		"unknown roles":     func(v *RBACAssignment) { v.RoleIDs = types.SetUnknown(types.StringType) },
		"empty role ID": func(v *RBACAssignment) {
			v.RoleIDs = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("")})
		},
		"unknown entities":    func(v *RBACAssignment) { v.AllowedEntities = types.SetUnknown(assignmentEntityObjectType) },
		"unknown description": func(v *RBACAssignment) { v.Description = types.StringUnknown() },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			v := assignmentFixture()
			change(&v)
			_, d := assignmentInput(context.Background(), v, true)
			require.True(t, d.HasError())
		})
	}
	v := assignmentFixture()
	v.UserID = types.StringNull()
	v.GroupID = types.StringValue("8")
	v.AllowViewingUnobfuscatedData = types.BoolNull()
	input, d := assignmentInput(context.Background(), v, true)
	require.False(t, d.HasError())
	require.Empty(t, input.Users)
	require.Equal(t, []rbacapi.RBACReference{{By: "ID", Input: "8"}}, input.Groups)
	require.Nil(t, input.AllowViewingUnobfuscatedData)
}

func TestRBACAssignmentHydration(t *testing.T) {
	var api rbacapi.RBACAssignmentAPI
	require.NoError(t, json.Unmarshal([]byte(assignmentJSON), &api))
	previous := assignmentFixture()
	for _, status := range []string{"DISABLED", "OBFUSCATED", "UNOBFUSCATED"} {
		t.Run(status, func(t *testing.T) {
			api.DataObfuscationStatus = status
			previous.AllowViewingUnobfuscatedData = types.BoolValue(true)
			model, d := assignmentFromAPI("123", &api, &previous)
			require.False(t, d.HasError())
			require.Equal(t, status != "OBFUSCATED", model.AllowViewingUnobfuscatedData.ValueBool())
		})
	}
	api.AssignmentType = "ADMIN"
	_, d := assignmentFromAPI("123", &api, &previous)
	require.True(t, d.HasError())
	api.AssignmentType = "ADMIN_USER"
	api.Account.ID = "456"
	_, d = assignmentFromAPI("123", &api, &previous)
	require.True(t, d.HasError())
}

func TestRBACAssignmentImport(t *testing.T) {
	r, _, state := assignmentTestResource(t, func(_ http.ResponseWriter, _ *http.Request) { t.Fatal("import must not make an API call") })
	for _, id := range []string{"123:a1", "a1", "123:", ":a1", "123:a1:extra"} {
		t.Run(id, func(t *testing.T) {
			resp := resource.ImportStateResponse{State: state}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
			if id != "123:a1" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError())
			var got RBACAssignment
			require.False(t, resp.State.Get(context.Background(), &got).HasError())
			require.Equal(t, "a1", got.AssignmentID.ValueString())
		})
	}
}

func TestRBACAssignmentGroupCreate(t *testing.T) {
	ctx := context.Background()
	r, plan, state := assignmentTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.JSONEq(t, `{"users":[],"groups":[{"by":"ID","input":"8"}],"roles":[{"by":"ID","input":"42"}],"allowedEntities":[{"type":"SITE","ref":{"by":"ID","input":"99"},"action":"EDIT"}],"allowAllSites":"NONE","allowAllUsers":"NONE","description":""}`, string(body.Variables["input"]))
		value := strings.ReplaceAll(assignmentJSON, `"user":{"id":"7"},"group":null,"assignmentType":"ADMIN_USER"`, `"user":null,"group":{"id":"8"},"assignmentType":"GROUP"`)
		value = strings.ReplaceAll(value, `"allowedEntities":[]`, `"allowedEntities":[{"id":"99","type":"SITE","action":"EDIT"}]`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"rbac":{"addRbacAssignments":{"items":[` + value + `]}}}}`))
	})
	model := assignmentFixture()
	model.UserID = types.StringNull()
	model.GroupID = types.StringValue("8")
	model.AllowViewingUnobfuscatedData = types.BoolNull()
	model.AllowedEntities = types.SetValueMust(assignmentEntityObjectType, []attr.Value{types.ObjectValueMust(assignmentEntityAttrTypes, map[string]attr.Value{"id": types.StringValue("99"), "type": types.StringValue("SITE"), "action": types.StringValue("EDIT")})})
	require.False(t, plan.Set(ctx, model).HasError())
	resp := resource.CreateResponse{State: state}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.False(t, resp.State.Get(ctx, &model).HasError())
	require.Equal(t, "GROUP", model.AssignmentType.ValueString())
	require.True(t, model.AllowViewingUnobfuscatedData.IsNull())
}

func TestRBACAssignmentLegacyDeleteRejected(t *testing.T) {
	ctx := context.Background()
	calls := 0
	r, _, state := assignmentTestResource(t, func(w http.ResponseWriter, req *http.Request) {
		calls++
		var body struct {
			OperationName string `json:"operationName"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, "rbacAssignmentList", body.OperationName)
		value := strings.ReplaceAll(assignmentJSON, `"assignmentType":"ADMIN_USER"`, `"assignmentType":"ADMIN"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[` + value + `],"paging":{"total":1}}}}}`))
	})
	model := assignmentFixture()
	model.AssignmentID = types.StringValue("a1")
	require.False(t, state.Set(ctx, model).HasError())
	resp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, 1, calls)
}

func TestRBACAssignmentInvalidMutationResponses(t *testing.T) {
	for _, payload := range []string{`null`, `{"items":[]}`, `{"items":[null]}`, `{"items":[` + assignmentJSON + `,` + assignmentJSON + `]}`, `{"items":[` + strings.ReplaceAll(assignmentJSON, `"123"`, `"456"`) + `]}`} {
		t.Run(payload, func(t *testing.T) {
			r, plan, state := assignmentTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"rbac":{"addRbacAssignments":` + payload + `}}}`))
			})
			resp := resource.CreateResponse{State: state}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
			require.True(t, resp.Diagnostics.HasError())
		})
	}
}

func TestRBACAssignmentImportRefresh(t *testing.T) {
	ctx := context.Background()
	r, _, state := assignmentTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[` + assignmentJSON + `],"paging":{"total":1}}}}}`))
	})
	model := assignmentFixture()
	model.UserID = types.StringNull()
	model.AssignmentID = types.StringValue("a1")
	model.AllowViewingUnobfuscatedData = types.BoolNull()
	require.False(t, state.Set(ctx, model).HasError())
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.False(t, resp.State.Get(ctx, &model).HasError())
	require.Equal(t, "7", model.UserID.ValueString())
	require.True(t, model.AllowViewingUnobfuscatedData.IsNull())
}

func TestRBACAssignmentUpdateReplacesRolesAndScope(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewRBACAssignmentClient(t)
	r := &rbacAssignmentResource{client: &catoClientData{AccountId: "123"}, assignmentClient: client}
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	original := assignmentFixture()
	original.ID = types.StringValue("123:a1")
	original.AssignmentID = types.StringValue("a1")
	original.AssignmentType = types.StringValue("ADMIN_USER")
	state := tfsdk.State{Schema: sr.Schema}
	require.False(t, state.Set(ctx, original).HasError())
	changed := original
	changed.Description = types.StringValue("updated")
	changed.AllowViewingUnobfuscatedData = types.BoolNull()
	changed.RoleIDs = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("43")})
	changed.AllowedEntities = types.SetValueMust(assignmentEntityObjectType, []attr.Value{types.ObjectValueMust(assignmentEntityAttrTypes, map[string]attr.Value{"id": types.StringValue("99"), "type": types.StringValue("SYSTEMGROUP"), "action": types.StringValue("VIEW")})})
	plan := tfsdk.Plan{Schema: sr.Schema}
	require.False(t, plan.Set(ctx, changed).HasError())
	var api rbacapi.RBACAssignmentAPI
	require.NoError(t, json.Unmarshal([]byte(assignmentJSON), &api))
	api.Roles = []*rbacapi.RBACObject{{ID: "43"}}
	api.AllowedEntities = []*rbacapi.RBACEntity{{ID: "99", Type: "SYSTEMGROUP", Action: "VIEW"}}
	description := "updated"
	api.Description = &description
	api.AllowAllSites = "VIEW"
	client.EXPECT().UpdateAssignment(mock.Anything, "123", mock.MatchedBy(func(input rbacapi.RBACAssignmentInput) bool {
		require.Equal(t, "a1", input.AssignmentID)
		require.Equal(t, []rbacapi.RBACReference{{By: "ID", Input: "43"}}, input.Roles)
		require.Equal(t, []rbacapi.RBACEntityInput{{Type: "SYSTEMGROUP", Ref: rbacapi.RBACReference{By: "ID", Input: "99"}, Action: "VIEW"}}, input.AllowedEntities)
		require.Equal(t, "NONE", input.AllowAllSites)
		require.Nil(t, input.AllowViewingUnobfuscatedData)
		require.Equal(t, "updated", input.Description)
		return true
	})).Return(&api, nil).Once()
	resp := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	var got RBACAssignment
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "VIEW", got.AllowAllSites.ValueString())
	require.True(t, got.RoleIDs.Equal(changed.RoleIDs))
	require.True(t, got.AllowedEntities.Equal(changed.AllowedEntities))
}

func TestRBACAssignmentConfigureAndMetadata(t *testing.T) {
	r := NewRBACAssignmentResource().(*rbacAssignmentResource)
	var metadata resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_rbac_assignment", metadata.TypeName)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "invalid"}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	resp = resource.ConfigureResponse{}
	client := &catoClientData{AccountId: "123"}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Same(t, client, r.client)
}
