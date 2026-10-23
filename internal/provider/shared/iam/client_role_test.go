package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

func TestRolePagination(t *testing.T) {
	for _, scenario := range []string{"offsets", "repeated", "empty", "changing total", "missing paging"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Variables struct {
						Input m.RoleManagementRoleListInput `json:"roleManagementRoleListInput"`
					} `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				require.Equal(t, int64(calls), request.Variables.Input.Paging.From)
				calls++
				id, total := "first", "2"
				if calls > 1 {
					id = "second"
				}
				if scenario == "repeated" {
					id = "first"
				}
				if scenario == "changing total" && calls > 1 {
					total = "3"
				}
				item := `{"id":"` + id + `","name":"test","permission":[],"applicableAccountType":[]}`
				if scenario == "empty" {
					item = ""
				}
				paging := `"__typename":"PageInfo","total":` + total
				if scenario == "missing paging" {
					paging = ""
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"rbac":{"roleManagement":{"roleList":{"items":[` + item + `],"paging":{` + paging + `}}}}}}`))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
			require.NoError(t, err)
			values, err := ListRoles(context.Background(), client, "123", m.RoleManagementRoleListInput{Sort: &m.RoleManagementRoleSortInput{}})
			if scenario == "offsets" {
				require.NoError(t, err)
				require.Len(t, values, 2)
				require.Equal(t, 2, calls)
			} else {
				require.Error(t, err)
			}
		})
	}
}
func TestRoleLookupMissingVersusMalformed(t *testing.T) {
	for _, test := range []struct {
		body    string
		invalid bool
	}{
		{`{"data":{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":null}}}}`, false},
		{`{"data":{"rbac":{}}}`, true},
		{`{"data":{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":{"id":"wrong"}}}}}`, true},
		{`{"errors":[{"message":"denied"}]}`, true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(test.body))
		}))
		client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
		require.NoError(t, err)
		value, err := LookupRole(context.Background(), client, "123", "r")
		if test.invalid {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Nil(t, value)
		}
		server.Close()
	}
}
