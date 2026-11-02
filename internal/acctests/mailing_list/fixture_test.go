//go:build acctest

package mailinglist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

func TestEligibleMailingListAdmin(t *testing.T) {
	login, service := m.AdminTypeLogin, m.AdminTypeServicePrinciple
	email, internal := "test@example.invalid", "service@internal.system"
	for _, tc := range []struct {
		name  string
		admin *cato.Admins_Admins_Items
		want  bool
	}{
		{"login", &cato.Admins_Admins_Items{ID: "42", AdminType: &login, Email: &email}, true},
		{"service principal", &cato.Admins_Admins_Items{ID: "43", AdminType: &service, Email: &internal}, false},
		{"missing type", &cato.Admins_Admins_Items{ID: "44", Email: &email}, false},
		{"missing id", &cato.Admins_Admins_Items{AdminType: &login, Email: &email}, false},
		{"nil", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.want, eligibleMailingListAdmin(tc.admin)) })
	}
}

func TestFindMailingListAdmin(t *testing.T) {
	for _, tc := range []struct {
		name       string
		secondPage string
		want       string
		wantError  bool
	}{
		{"paginate past service principal", `{"id":"42","adminType":"LOGIN","email":"test@example.invalid"}`, "42", false},
		{"no login fixture", `{"id":"43","adminType":"SERVICE_PRINCIPLE"}`, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				var body struct {
					Variables map[string]any `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "123", body.Variables["accountID"])
				require.Equal(t, float64(calls), body.Variables["from"])
				require.Equal(t, float64(100), body.Variables["limit"])
				item := `{"id":"41","adminType":"SERVICE_PRINCIPLE"}`
				if calls == 1 {
					item = tc.secondPage
				}
				calls++
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"data":{"admins":{"items":[%s],"total":2}}}`, item)
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
			require.NoError(t, err)
			id, err := findMailingListAdmin(context.Background(), client, "123")
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, id)
			require.Equal(t, 2, calls)
		})
	}
}
