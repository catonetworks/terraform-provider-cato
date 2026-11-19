package networkconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestSiteUpdatePreservesPartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { require.NoError(t, r.Body.Close()) }()
		var request struct {
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "account-123", request.Variables["accountId"])
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"networkConfig":{"dns":{"updateSiteSettings":{"siteSettings":[{"siteId":"site-123","primaryServer":null,"secondaryServer":null,"suffix":[]}]}}}},"errors":[{"message":"partial response"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	sdk, err := cato.New(server.URL, "test-key", "wrong-default", server.Client(), nil)
	require.NoError(t, err)
	require.False(t, sdk.Client.ParseDataWhenErrors)
	client := NewDNSClient(catoclient.NewProviderSDKClient(sdk))
	result, err := client.UpdateDNSSiteSettings(context.Background(), "account-123", models.SiteRefInput{By: models.ObjectRefByID, Input: "site-123"}, DNSSettingsInput{Suffix: []string{}})
	require.False(t, sdk.Client.ParseDataWhenErrors)
	require.Error(t, err)
	require.NotNil(t, result)
	require.Equal(t, "site-123", result.GetNetworkConfig().GetDNS().GetUpdateSiteSettings().GetSiteSettings()[0].GetSiteID())
}
