package dnssuffixset

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped DNS suffix-set CRUD contract.
type ResourceClient interface {
	NetworkConfigDNSCreateSuffixSet(context.Context, string, m.NetworkConfigDNSCreateSuffixSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSCreateSuffixSet, error)
	NetworkConfigDNSUpdateSuffixSet(context.Context, string, m.NetworkConfigDNSUpdateSuffixSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSUpdateSuffixSet, error)
	NetworkConfigDNSDeleteSuffixSet(context.Context, string, m.NetworkConfigDNSDeleteSuffixSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSDeleteSuffixSet, error)
	NetworkConfigDNSSuffixSet(context.Context, string, m.NetworkConfigDNSSuffixSetRefInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSSuffixSet, error)
}
type sdkClient struct{ client *catoclient.Service }

const createSuffixSetDocument = `mutation networkConfigDnsCreateSuffixSet(
$accountId:ID!, $networkConfigDnsCreateSuffixSetInput:NetworkConfigDnsCreateSuffixSetInput!
) { networkConfig(accountId:$accountId) { dns {
createSuffixSet(input:$networkConfigDnsCreateSuffixSetInput) { dnsSuffixSet { id name suffix } }
} } }`

func (c *sdkClient) NetworkConfigDNSCreateSuffixSet(ctx context.Context, account string,
	input m.NetworkConfigDNSCreateSuffixSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSCreateSuffixSet, error) {
	result := new(cato.NetworkConfigDNSCreateSuffixSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsCreateSuffixSet", createSuffixSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsCreateSuffixSetInput": input}, interceptors...)
	return result, err
}

const updateSuffixSetDocument = `mutation networkConfigDnsUpdateSuffixSet(
$accountId:ID!, $networkConfigDnsUpdateSuffixSetInput:NetworkConfigDnsUpdateSuffixSetInput!
) { networkConfig(accountId:$accountId) { dns {
updateSuffixSet(input:$networkConfigDnsUpdateSuffixSetInput) { dnsSuffixSet { id name suffix } }
} } }`

func (c *sdkClient) NetworkConfigDNSUpdateSuffixSet(ctx context.Context, account string,
	input m.NetworkConfigDNSUpdateSuffixSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSUpdateSuffixSet, error) {
	items := make([]map[string]interface{}, 0, len(input.DNSSuffixSet))
	for _, item := range input.DNSSuffixSet {
		items = append(items, map[string]interface{}{"id": item.ID, "name": item.Name, "suffix": item.Suffix})
	}
	result := new(cato.NetworkConfigDNSUpdateSuffixSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsUpdateSuffixSet", updateSuffixSetDocument, result,
		map[string]interface{}{"accountId": account,
			"networkConfigDnsUpdateSuffixSetInput": map[string]interface{}{"dnsSuffixSet": items}}, interceptors...)
	return result, err
}

const deleteSuffixSetDocument = `mutation networkConfigDnsDeleteSuffixSet(
$accountId:ID!, $networkConfigDnsDeleteSuffixSetInput:NetworkConfigDnsDeleteSuffixSetInput!
) { networkConfig(accountId:$accountId) { dns {
deleteSuffixSet(input:$networkConfigDnsDeleteSuffixSetInput) { dnsSuffixSet { id name suffix } }
} } }`

func (c *sdkClient) NetworkConfigDNSDeleteSuffixSet(ctx context.Context, account string,
	input m.NetworkConfigDNSDeleteSuffixSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSDeleteSuffixSet, error) {
	result := new(cato.NetworkConfigDNSDeleteSuffixSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsDeleteSuffixSet", deleteSuffixSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsDeleteSuffixSetInput": input}, interceptors...)
	return result, err
}

const suffixSetDocument = `query networkConfigDnsSuffixSet(
$accountId:ID!, $networkConfigDnsSuffixSetRefInput:NetworkConfigDnsSuffixSetRefInput!
) { networkConfig(accountId:$accountId) { dns {
suffixSet(input:$networkConfigDnsSuffixSetRefInput) { id name suffix }
} } }`

func (c *sdkClient) NetworkConfigDNSSuffixSet(ctx context.Context, account string,
	input m.NetworkConfigDNSSuffixSetRefInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSSuffixSet, error) {
	result := new(cato.NetworkConfigDNSSuffixSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsSuffixSet", suffixSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsSuffixSetRefInput": input}, interceptors...)
	return result, err
}

type readRecord = cato.NetworkConfigDnsSuffixSet_NetworkConfig_DNS_SuffixSet

func (r *dnsSetResource) find(ctx context.Context, id string) (*readRecord, error) {
	result, err := r.getClient().NetworkConfigDNSSuffixSet(ctx, r.client.AccountId,
		m.NetworkConfigDNSSuffixSetRefInput{By: m.ObjectRefByID, Input: id})
	if err != nil {
		return nil, err
	}
	if result == nil || result.GetNetworkConfig() == nil {
		return nil, fmt.Errorf("API did not return a networkConfig response")
	}
	item := result.GetNetworkConfig().GetDNS().GetSuffixSet()
	if item != nil && item.GetID() != id {
		return nil, fmt.Errorf("API returned a different DNS set identity")
	}
	return item, nil
}
