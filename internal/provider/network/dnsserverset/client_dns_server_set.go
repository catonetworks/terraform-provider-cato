package dnsserverset

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped DNS server-set CRUD contract.
type ResourceClient interface {
	NetworkConfigDNSCreateServerSet(context.Context, string, m.NetworkConfigDNSCreateServerSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSCreateServerSet, error)
	NetworkConfigDNSUpdateServerSet(context.Context, string, m.NetworkConfigDNSUpdateServerSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSUpdateServerSet, error)
	NetworkConfigDNSDeleteServerSet(context.Context, string, m.NetworkConfigDNSDeleteServerSetInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSDeleteServerSet, error)
	NetworkConfigDNSServerSet(context.Context, string, m.NetworkConfigDNSServerSetRefInput,
		...clientv2.RequestInterceptor,
	) (*cato.NetworkConfigDNSServerSet, error)
}
type sdkClient struct{ client *catoclient.Service }

const createServerSetDocument = `mutation networkConfigDnsCreateServerSet(
$accountId:ID!, $networkConfigDnsCreateServerSetInput:NetworkConfigDnsCreateServerSetInput!
) { networkConfig(accountId:$accountId) { dns {
createServerSet(input:$networkConfigDnsCreateServerSetInput) { dnsServerSet { id name server } }
} } }`

func (c *sdkClient) NetworkConfigDNSCreateServerSet(ctx context.Context, account string,
	input m.NetworkConfigDNSCreateServerSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSCreateServerSet, error) {
	result := new(cato.NetworkConfigDNSCreateServerSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsCreateServerSet", createServerSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsCreateServerSetInput": input}, interceptors...)
	return result, err
}

const updateServerSetDocument = `mutation networkConfigDnsUpdateServerSet(
$accountId:ID!, $networkConfigDnsUpdateServerSetInput:NetworkConfigDnsUpdateServerSetInput!
) { networkConfig(accountId:$accountId) { dns {
updateServerSet(input:$networkConfigDnsUpdateServerSetInput) { dnsServerSet { id name server } }
} } }`

func (c *sdkClient) NetworkConfigDNSUpdateServerSet(ctx context.Context, account string,
	input m.NetworkConfigDNSUpdateServerSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSUpdateServerSet, error) {
	items := make([]map[string]interface{}, 0, len(input.DNSServerSet))
	for _, item := range input.DNSServerSet {
		items = append(items, map[string]interface{}{"id": item.ID, "name": item.Name, "server": item.Server})
	}
	result := new(cato.NetworkConfigDNSUpdateServerSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsUpdateServerSet", updateServerSetDocument, result,
		map[string]interface{}{"accountId": account,
			"networkConfigDnsUpdateServerSetInput": map[string]interface{}{"dnsServerSet": items}}, interceptors...)
	return result, err
}

const deleteServerSetDocument = `mutation networkConfigDnsDeleteServerSet(
$accountId:ID!, $networkConfigDnsDeleteServerSetInput:NetworkConfigDnsDeleteServerSetInput!
) { networkConfig(accountId:$accountId) { dns {
deleteServerSet(input:$networkConfigDnsDeleteServerSetInput) { dnsServerSet { id name server } }
} } }`

func (c *sdkClient) NetworkConfigDNSDeleteServerSet(ctx context.Context, account string,
	input m.NetworkConfigDNSDeleteServerSetInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSDeleteServerSet, error) {
	result := new(cato.NetworkConfigDNSDeleteServerSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsDeleteServerSet", deleteServerSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsDeleteServerSetInput": input}, interceptors...)
	return result, err
}

const serverSetDocument = `query networkConfigDnsServerSet(
$accountId:ID!, $networkConfigDnsServerSetRefInput:NetworkConfigDnsServerSetRefInput!
) { networkConfig(accountId:$accountId) { dns {
serverSet(input:$networkConfigDnsServerSetRefInput) { id name server }
} } }`

func (c *sdkClient) NetworkConfigDNSServerSet(ctx context.Context, account string,
	input m.NetworkConfigDNSServerSetRefInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.NetworkConfigDNSServerSet, error) {
	result := new(cato.NetworkConfigDNSServerSet)
	err := c.client.V2().Client.Post(ctx, "networkConfigDnsServerSet", serverSetDocument, result,
		map[string]interface{}{"accountId": account, "networkConfigDnsServerSetRefInput": input}, interceptors...)
	return result, err
}

type readRecord = cato.NetworkConfigDnsServerSet_NetworkConfig_DNS_ServerSet

func (r *dnsSetResource) find(ctx context.Context, id string) (*readRecord, error) {
	result, err := r.getClient().NetworkConfigDNSServerSet(ctx, r.client.AccountId,
		m.NetworkConfigDNSServerSetRefInput{By: m.ObjectRefByID, Input: id})
	if err != nil {
		return nil, err
	}
	if result == nil || result.GetNetworkConfig() == nil {
		return nil, fmt.Errorf("API did not return a networkConfig response")
	}
	item := result.GetNetworkConfig().GetDNS().GetServerSet()
	if item != nil && item.GetID() != id {
		return nil, fmt.Errorf("API returned a different DNS set identity")
	}
	return item, nil
}
