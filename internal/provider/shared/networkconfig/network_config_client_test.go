package networkconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestDNSUpdatePreservesNullAndEmpty(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "test-key", r.Header.Get("x-api-key"))
		require.Equal(t, "account-123", r.Header.Get("x-account-id"))
		require.JSONEq(t, `{"primaryServer":null,"secondaryServer":null,"suffix":[],"acceptDnsRequestsOnLanInterfaceIp":false}`,
			string(body.Variables["networkConfigDnsUpdateSettingsInput"]))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"networkConfig":{"dns":{"updateSettings":{"settings":{"primaryServer":null,"secondaryServer":null,"suffix":[],"acceptDnsRequestsOnLanInterfaceIp":false}}}}}}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
	require.NoError(t, err)
	result, err := NewDNSClient(catoclient.NewProviderSDKClient(sdk)).UpdateDNSSettings(context.Background(), "account-123", DNSSettingsInput{Suffix: []string{}})
	require.NoError(t, err)
	require.NotNil(t, result.NetworkConfig.DNS.UpdateSettings.Settings)
}

func TestSiteDNSUpdatePreservesNullAndEmpty(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.JSONEq(t, `{"siteSettings":[{"site":{"by":"ID","input":"site-123"},"primaryServer":null,"secondaryServer":null,"suffix":[]}]}`,
			string(body.Variables["networkConfigDnsUpdateSiteSettingsInput"]))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"networkConfig":{"dns":{"updateSiteSettings":{"siteSettings":[{"siteId":"site-123","primaryServer":null,"secondaryServer":null,"suffix":[]}]}}}}}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
	require.NoError(t, err)
	result, err := NewDNSClient(catoclient.NewProviderSDKClient(sdk)).UpdateDNSSiteSettings(context.Background(), "account-123", cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: "site-123"}, DNSSettingsInput{Suffix: []string{}})
	require.NoError(t, err)
	require.Len(t, result.NetworkConfig.DNS.UpdateSiteSettings.SiteSettings, 1)
}
