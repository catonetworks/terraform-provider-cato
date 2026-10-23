package rbacassignment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
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

func assignmentPlan(t *testing.T, r *assignmentResource) tfsdk.Plan {
	t.Helper()
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	model := Model{ID: types.StringNull(), AccountID: types.StringNull(), AssignmentID: types.StringNull(), UserID: types.StringValue("u"), GroupID: types.StringNull(), RoleIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("r")}), AllowedEntities: types.SetValueMust(entityObjectType, []attr.Value{}), Description: types.StringValue(""), AllowViewingUnobfuscatedData: types.BoolValue(true), AllowAllSites: types.StringNull(), AllowAllUsers: types.StringNull(), AssignmentType: types.StringNull(), DataObfuscationStatus: types.StringNull()}
	require.False(t, plan.Set(context.Background(), model).HasError())
	return plan
}
func TestAssignmentCRUD(t *testing.T) {
	ctx := context.Background()
	missing := false
	assignment := `{"id":"a","account":{"id":"123"},"user":{"id":"u"},"group":null,"assignmentType":"ADMIN_USER","roles":[{"id":"r"}],"allowedEntities":[],"allowAllSites":"NONE","allowAllUsers":"NONE","dataObfuscationStatus":"DISABLED","description":""}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Account string                     `json:"accountID"`
				Input   map[string]json.RawMessage `json:"input"`
			} `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "123", body.Variables.Account)
		w.Header().Set("Content-Type", "application/json")
		payload := ""
		switch {
		case strings.Contains(body.Query, "rbacAddRbacAssignments"):
			require.Equal(t, "[]", string(body.Variables.Input["groups"]))
			payload = `"addRbacAssignments":{"items":[` + assignment + `]}`
		case strings.Contains(body.Query, "rbacUpdateRbacAssignment"):
			require.NotContains(t, body.Variables.Input, "users")
			require.NotContains(t, body.Variables.Input, "groups")
			require.Equal(t, `"a"`, string(body.Variables.Input["rbacAssignmentId"]))
			payload = `"updateRbacAssignment":{"assignment":` + assignment + `}`
		case strings.Contains(body.Query, "rbacRbacAssignmentList"):
			require.JSONEq(t, `{}`, string(body.Variables.Input["filter"]))
			items, total := assignment, "1"
			if missing {
				items, total = "", "0"
			}
			payload = `"rbacAssignmentList":{"items":[` + items + `],"paging":{"__typename":"PageInfo","total":` + total + `}}`
		default:
			payload = `"removeRbacAssignment":{"rbacAssignmentId":"a"}`
		}
		_, _ = w.Write([]byte(`{"data":{"rbac":{` + payload + `}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
	require.NoError(t, err)
	r := &assignmentResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
	plan := assignmentPlan(t, r)
	created := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
	require.False(t, created.Diagnostics.HasError(), created.Diagnostics)
	var state Model
	require.False(t, created.State.Get(ctx, &state).HasError())
	require.Equal(t, "123:a", state.ID.ValueString())
	require.True(t, state.AllowViewingUnobfuscatedData.ValueBool())
	read := resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	updated := resource.UpdateResponse{State: created.State}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: created.State}, &updated)
	require.False(t, updated.Diagnostics.HasError(), updated.Diagnostics)
	deleted := resource.DeleteResponse{State: updated.State}
	r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError(), deleted.Diagnostics)
	missing = true
	read = resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.True(t, read.State.Raw.IsNull())
	deleted = resource.DeleteResponse{State: created.State}
	r.Delete(ctx, resource.DeleteRequest{State: created.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError())
	imported := resource.ImportStateResponse{State: created.State}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "123:a"}, &imported)
	require.False(t, imported.Diagnostics.HasError(), imported.Diagnostics)
}
func TestAssignmentPagination(t *testing.T) {
	for _, scenario := range []string{"pages", "repeated", "empty", "changed total", "missing paging", "cross account", "graphql error"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Variables struct {
						Input struct {
							Paging struct {
								From int64 `json:"from"`
							} `json:"paging"`
						} `json:"input"`
					} `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				expectedPage := calls
				if scenario == "changed total" {
					expectedPage = calls % 2
				}
				require.Equal(t, int64(expectedPage), body.Variables.Input.Paging.From)
				calls++
				id, account, total := "other", "123", "2"
				if calls > 1 {
					id = "target"
				}
				if scenario == "repeated" {
					id = "other"
				}
				if scenario == "cross account" {
					account = "456"
				}
				if scenario == "changed total" {
					id = "other"
					if calls%2 == 0 {
						id, total = "second", "3"
					}
				}
				item := `{"id":"` + id + `","account":{"id":"` + account + `"}}`
				if scenario == "empty" {
					item = ""
				}
				paging := `"__typename":"PageInfo","total":` + total
				if scenario == "missing paging" {
					paging = ""
				}
				w.Header().Set("Content-Type", "application/json")
				if scenario == "graphql error" {
					_, _ = w.Write([]byte(`{"errors":[{"message":"denied"}]}`))
					return
				}
				_, _ = w.Write([]byte(`{"data":{"rbac":{"rbacAssignmentList":{"items":[` + item + `],"paging":{` + paging + `}}}}}`))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
			require.NoError(t, err)
			value, err := lookupAssignment(context.Background(), client, "123", "target")
			if scenario == "pages" {
				require.NoError(t, err)
				require.Equal(t, "target", value.ID)
				require.Equal(t, 2, calls)
			} else {
				require.Error(t, err)
				if scenario == "changed total" {
					require.Equal(t, 6, calls)
					r := &assignmentResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
					state := tfsdk.State(assignmentPlan(t, r))
					require.False(t, state.SetAttribute(context.Background(), path.Root("id"), "123:target").HasError())
					read := resource.ReadResponse{State: state}
					r.Read(context.Background(), resource.ReadRequest{State: state}, &read)
					require.True(t, read.Diagnostics.HasError())
					require.False(t, read.State.Raw.IsNull(), "concurrent changes must not remove an assignment from state")
					require.Equal(t, 12, calls)
				}
			}
		})
	}
}

func TestAssignmentPaginationRestartsAfterConcurrentDeletion(t *testing.T) {
	client := mocks.NewRBACAssignmentResourceClient(t)
	page := func(items string, total int) *sdk.RbacRbacAssignmentList {
		response := &sdk.RbacRbacAssignmentList{}
		require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"rbac":{"rbacAssignmentList":{"items":[%s],"paging":{"__typename":"PageInfo","total":%d}}}}`, items, total)), response))
		return response
	}
	items := make([]string, iam.PageSize)
	for i := range items {
		items[i] = fmt.Sprintf(`{"id":"other-%d","account":{"id":"123"}}`, i)
	}
	firstPage := strings.Join(items, ",")
	// Deleting an earlier row shifts the target from page one to page zero.
	items[len(items)-1] = `{"id":"target","account":{"id":"123"}}`
	for i, response := range []*sdk.RbacRbacAssignmentList{
		page(firstPage, 101),
		page(``, 100),
		page(strings.Join(items, ","), 100),
	} {
		from := int64(i % 2)
		client.EXPECT().RbacRbacAssignmentList(mock.Anything, "123", mock.MatchedBy(func(input m.RBACAssignmentListInput) bool {
			return input.Filter != nil && input.Paging != nil && input.Paging.From == from && input.Paging.Limit == iam.PageSize
		})).Return(response, nil).Once()
	}
	value, err := lookupAssignment(context.Background(), client, "123", "target")
	require.NoError(t, err)
	require.Equal(t, "target", value.ID)
}

func TestAssignmentValidationStopsBeforeAPI(t *testing.T) {
	client := mocks.NewRBACAssignmentResourceClient(t)
	r := &assignmentResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := assignmentPlan(t, r)
	var model Model
	require.False(t, plan.Get(context.Background(), &model).HasError())
	model.GroupID = types.StringValue("g")
	require.False(t, plan.Set(context.Background(), model).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
}
func TestAssignmentRejectsLegacyMutation(t *testing.T) {
	client := mocks.NewRBACAssignmentResourceClient(t)
	page := &sdk.RbacRbacAssignmentList{}
	require.NoError(t, json.Unmarshal([]byte(`{"rbac":{"rbacAssignmentList":{"items":[{"id":"a","account":{"id":"123"},"assignmentType":"ADMIN","roles":[{"id":"r"}],"allowedEntities":[]}],"paging":{"__typename":"PageInfo","total":1}}}}`), page))
	client.EXPECT().RbacRbacAssignmentList(mock.Anything, "123", mock.Anything).Return(page, nil).Twice()
	r := &assignmentResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := assignmentPlan(t, r)
	state := tfsdk.State(plan)
	require.False(t, state.SetAttribute(context.Background(), path.Root("id"), "123:a").HasError())
	updated := resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &updated)
	require.True(t, updated.Diagnostics.HasError())
	deleted := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &deleted)
	require.True(t, deleted.Diagnostics.HasError())
}
