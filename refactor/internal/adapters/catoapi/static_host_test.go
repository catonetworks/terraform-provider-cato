package catoapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

func TestCreateStaticHostSDKMapping(t *testing.T) {
	t.Parallel()
	mac := "00:11:22:33:44:55"
	for _, configuredMAC := range []*string{nil, &mac} {
		client, requests := testCatoAPI(t, `{"data":{"site":{"addStaticHost":{"hostId":"host-123"}}}}`)
		id, err := NewAdapter(client).CreateStaticHost(context.Background(), "account-123", "site-123",
			entities.StaticHost{Name: "printer", IP: "192.0.2.10", MacAddress: configuredMAC})
		require.NoError(t, err)
		require.Equal(t, "host-123", id)
		request := <-requests
		require.Equal(t, "siteAddStaticHost", request.OperationName)
		require.JSONEq(t, `"account-123"`, string(request.Variables["accountId"]))
		require.JSONEq(t, `"site-123"`, string(request.Variables["siteId"]))
		want := `{"name":"printer","ip":"192.0.2.10"}`
		if configuredMAC != nil {
			want = `{"name":"printer","ip":"192.0.2.10","macAddress":"00:11:22:33:44:55"}`
		}
		require.JSONEq(t, want, string(request.Variables["addStaticHostInput"]))
	}
}

func TestCreateStaticHostRejectsUnusableResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		response string
		want     error
	}{
		{`{"data":{"site":null}}`, network.ErrInvalidResponse},
		{`{"data":{"site":{"addStaticHost":null}}}`, network.ErrInvalidResponse},
		{`{"data":{"site":{"addStaticHost":{"hostId":""}}}}`, network.ErrInvalidResponse},
		{`{"errors":[{"message":"sensitive backend details"}]}`, network.ErrUnavailable},
	} {
		client, _ := testCatoAPI(t, tc.response)
		id, err := NewAdapter(client).CreateStaticHost(context.Background(), "account-123", "site-123",
			entities.StaticHost{Name: "printer", IP: "192.0.2.10"})
		require.ErrorIs(t, err, tc.want)
		require.Empty(t, id)
		require.NotContains(t, err.Error(), "sensitive backend details")
	}
}
