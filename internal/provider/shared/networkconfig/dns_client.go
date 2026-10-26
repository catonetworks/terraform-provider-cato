package networkconfig

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// DNSClient adds DNS request and response validation to the provider SDK.
type DNSClient struct{ *catoclient.ProviderSDKClient }

// NewDNSClient returns the DNS-specific SDK adapter.
func NewDNSClient(client *catoclient.ProviderSDKClient) *DNSClient {
	return &DNSClient{ProviderSDKClient: client}
}

// DNSSettingsInput preserves explicit null servers and an empty suffix list.
// Generated SDK inputs use omitempty, which cannot express these clear operations.
type DNSSettingsInput struct {
	PrimaryServer                     *string  `json:"primaryServer"`
	SecondaryServer                   *string  `json:"secondaryServer"`
	Suffix                            []string `json:"suffix"`
	AcceptDNSRequestsOnLanInterfaceIP bool     `json:"acceptDnsRequestsOnLanInterfaceIp"`
}

type dnsSiteSettingsInput struct {
	Site            cato_models.SiteRefInput `json:"site"`
	PrimaryServer   *string                  `json:"primaryServer"`
	SecondaryServer *string                  `json:"secondaryServer"`
	Suffix          []string                 `json:"suffix"`
}

const dnsSettingsDocument = `mutation networkConfigDnsUpdateSettings (
 $accountId: ID!,
 $networkConfigDnsUpdateSettingsInput: NetworkConfigDnsUpdateSettingsInput!
) {
	networkConfig(accountId: $accountId) {
		dns {
			updateSettings(input: $networkConfigDnsUpdateSettingsInput) {
				settings {
					primaryServer
					secondaryServer
					suffix
					acceptDnsRequestsOnLanInterfaceIp
				}
			}
		}
	}
}
`
const dnsSiteSettingsDocument = `mutation networkConfigDnsUpdateSiteSettings (
 $accountId: ID!,
 $networkConfigDnsUpdateSiteSettingsInput: NetworkConfigDnsUpdateSiteSettingsInput!
) {
	networkConfig(accountId: $accountId) {
		dns {
			updateSiteSettings(input: $networkConfigDnsUpdateSiteSettingsInput) {
				siteSettings {
					siteId
					primaryServer
					secondaryServer
					suffix
				}
			}
		}
	}
}
`

func (c *DNSClient) UpdateDNSSettings(ctx context.Context,
	accountID string,
	input DNSSettingsInput) (*cato.NetworkConfigDNSUpdateSettings,
	error) {
	var result cato.NetworkConfigDNSUpdateSettings
	err := c.Client.Client.Post(ctx, "networkConfigDnsUpdateSettings", dnsSettingsDocument, &result,
		map[string]any{"accountId": accountID, "networkConfigDnsUpdateSettingsInput": input})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DNSClient) UpdateDNSSiteSettings(ctx context.Context,
	accountID string,
	site cato_models.SiteRefInput,
	input DNSSettingsInput) (*cato.NetworkConfigDNSUpdateSiteSettings,
	error) {
	var result cato.NetworkConfigDNSUpdateSiteSettings
	settings := dnsSiteSettingsInput{
		Site:          site,
		PrimaryServer: input.PrimaryServer, SecondaryServer: input.SecondaryServer, Suffix: input.Suffix,
	}
	err := c.Client.Client.Post(ctx, "networkConfigDnsUpdateSiteSettings", dnsSiteSettingsDocument, &result,
		map[string]any{"accountId": accountID,
			"networkConfigDnsUpdateSiteSettingsInput": map[string]any{"siteSettings": []dnsSiteSettingsInput{settings}}})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DNSClient) NetworkConfigDNSSettings(ctx context.Context, accountID string,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSSettings, error) {
	return c.ProviderSDKClient.NetworkConfigDNSSettings(ctx, accountID,
		append(interceptors, validateReadResponse("dns", "settings"))...)
}

func (c *DNSClient) NetworkConfigDNSSiteSettings(ctx context.Context, accountID string, siteRefInput cato_models.SiteRefInput,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSSiteSettings, error) {
	return c.ProviderSDKClient.NetworkConfigDNSSiteSettings(ctx, accountID, siteRefInput,
		append(interceptors, validateReadResponse("dns", "siteSettings"))...)
}

func (c *DNSClient) NetworkConfigDNSForwardingRule(ctx context.Context, accountID string,
	input cato_models.NetworkConfigDNSForwardingRuleRefInput,
	interceptors ...clientv2.RequestInterceptor) (*cato.NetworkConfigDNSForwardingRule, error) {
	return c.ProviderSDKClient.NetworkConfigDNSForwardingRule(ctx, accountID, input,
		append(interceptors, validateReadResponse("dns", "forwardingRule"))...)
}
