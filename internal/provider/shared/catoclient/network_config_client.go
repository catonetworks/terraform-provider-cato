package catoclient

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

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

func (c *ProviderSDKClient) UpdateDNSSettings(ctx context.Context,
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

func (c *ProviderSDKClient) UpdateDNSSiteSettings(ctx context.Context,
	accountID,
	siteID string,
	input DNSSettingsInput) (*cato.NetworkConfigDNSUpdateSiteSettings,
	error) {
	var result cato.NetworkConfigDNSUpdateSiteSettings
	settings := dnsSiteSettingsInput{
		Site:          cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID},
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
