package catoclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
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
	result, err := NewProviderSDKClient(sdk).UpdateDNSSettings(context.Background(), "account-123", DNSSettingsInput{Suffix: []string{}})
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
	result, err := NewProviderSDKClient(sdk).UpdateDNSSiteSettings(context.Background(), "account-123", "site-123", DNSSettingsInput{Suffix: []string{}})
	require.NoError(t, err)
	require.Len(t, result.NetworkConfig.DNS.UpdateSiteSettings.SiteSettings, 1)
}

func TestProviderSDKClientAllowsUninitializedClient(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() { NewProviderSDKClient(nil) })
	require.NotPanics(t, func() { NewProviderSDKClient(&cato.Client{}) })
}
