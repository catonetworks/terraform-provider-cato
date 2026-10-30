package cloudinterconnectphysicalconnection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ResourceClient is the account-scoped manual physical-connection CRUD contract.
type ResourceClient interface {
	SiteAddCloudInterconnectPhysicalConnection(ctx context.Context, input m.AddCloudInterconnectPhysicalConnectionInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteAddCloudInterconnectPhysicalConnection, error)
	SiteUpdateCloudInterconnectPhysicalConnection(ctx context.Context, input m.UpdateCloudInterconnectPhysicalConnectionInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateCloudInterconnectPhysicalConnection, error)
	SiteRemoveCloudInterconnectPhysicalConnection(ctx context.Context, input m.RemoveCloudInterconnectPhysicalConnectionInput, account string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteRemoveCloudInterconnectPhysicalConnection, error)
	SiteCloudInterconnectPhysicalConnection(ctx context.Context, account string, input m.CloudInterconnectPhysicalConnectionInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteCloudInterconnectPhysicalConnection, error)
	SiteCloudInterconnectPhysicalConnectionID(ctx context.Context, account string, input m.CloudInterconnectPhysicalConnectionIDInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato.SiteCloudInterconnectPhysicalConnectionID, error)
}
type sdkClient struct{ client *catoclient.Service }

func (c *sdkClient) SiteAddCloudInterconnectPhysicalConnection(
	ctx context.Context, input m.AddCloudInterconnectPhysicalConnectionInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteAddCloudInterconnectPhysicalConnection, error) {
	return c.client.V2().SiteAddCloudInterconnectPhysicalConnection(ctx, input, account, interceptors...)
}
func (c *sdkClient) SiteUpdateCloudInterconnectPhysicalConnection(
	ctx context.Context, input m.UpdateCloudInterconnectPhysicalConnectionInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteUpdateCloudInterconnectPhysicalConnection, error) {
	return c.client.V2().SiteUpdateCloudInterconnectPhysicalConnection(ctx, input, account, interceptors...)
}
func (c *sdkClient) SiteRemoveCloudInterconnectPhysicalConnection(
	ctx context.Context, input m.RemoveCloudInterconnectPhysicalConnectionInput, account string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteRemoveCloudInterconnectPhysicalConnection, error) {
	return c.client.V2().SiteRemoveCloudInterconnectPhysicalConnection(ctx, input, account, interceptors...)
}

// NetworkBandwidth is numeric on the wire, but SDK v1.1.3 maps its response to string.
// Preserve the SDK document/transport and normalize only those two fields before typed decoding.
func (c *sdkClient) SiteCloudInterconnectPhysicalConnection(
	ctx context.Context, account string,
	input m.CloudInterconnectPhysicalConnectionInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteCloudInterconnectPhysicalConnection, error) {
	var wire struct {
		Site struct {
			Connection json.RawMessage `json:"cloudInterconnectPhysicalConnection" graphql:"cloudInterconnectPhysicalConnection"`
		} `json:"site" graphql:"site"`
	}
	result := new(cato.SiteCloudInterconnectPhysicalConnection)
	err := c.client.V2().Client.Post(ctx, "siteCloudInterconnectPhysicalConnection",
		cato.SiteCloudInterconnectPhysicalConnectionDocument, &wire, map[string]interface{}{
			"accountId": account, "cloudInterconnectPhysicalConnectionInput": input,
		}, interceptors...)
	if err != nil {
		return result, err
	}
	if len(wire.Site.Connection) == 0 || bytes.Equal(bytes.TrimSpace(wire.Site.Connection), []byte("null")) {
		return result, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(wire.Site.Connection, &fields); err != nil {
		return result, fmt.Errorf("decode physical connection: %w", err)
	}
	for _, key := range []string{"upstreamBwLimit", "downstreamBwLimit"} {
		normalized, err := bandwidthJSON(fields[key])
		if err != nil {
			return result, fmt.Errorf("invalid %s response: %w", key, err)
		}
		fields[key] = normalized
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return result, err
	}
	item := new(connectionRecord)
	if err := json.Unmarshal(encoded, item); err != nil {
		return result, fmt.Errorf("decode physical connection fields: %w", err)
	}
	result.Site.CloudInterconnectPhysicalConnection = item
	return result, nil
}
func bandwidthJSON(raw json.RawMessage) (json.RawMessage, error) {
	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
	} else {
		var number json.Number
		if err := json.Unmarshal(raw, &number); err != nil {
			return nil, err
		}
		value = number.String()
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number < 0 {
		return nil, fmt.Errorf("expected a non-negative integer")
	}
	return json.Marshal(strconv.FormatInt(number, 10))
}
func (c *sdkClient) SiteCloudInterconnectPhysicalConnectionID(
	ctx context.Context, account string, input m.CloudInterconnectPhysicalConnectionIDInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.SiteCloudInterconnectPhysicalConnectionID, error) {
	return c.client.V2().SiteCloudInterconnectPhysicalConnectionID(ctx, account, input, interceptors...)
}

type connectionRecord = cato.SiteCloudInterconnectPhysicalConnection_Site_CloudInterconnectPhysicalConnection

func (r *connectionResource) find(ctx context.Context, id string) (*connectionRecord, error) {
	result, err := r.getClient().SiteCloudInterconnectPhysicalConnection(ctx, r.client.AccountId,
		m.CloudInterconnectPhysicalConnectionInput{ID: id})
	if err != nil {
		if missingConnection(err) {
			return nil, nil
		}
		return nil, err
	}
	item := result.GetSite().GetCloudInterconnectPhysicalConnection()
	if item == nil || item.GetID() != id {
		return nil, fmt.Errorf("API did not return the requested connection identity")
	}
	if role := item.GetHaRole(); role == nil || !role.IsValid() {
		return nil, fmt.Errorf("API returned an unsupported HA role")
	}
	if method := item.GetEncapsulationMethod(); method == nil || !method.IsValid() {
		return nil, fmt.Errorf("API returned an unsupported tagging method")
	}
	if item.GetSite().GetID() == "" || item.GetPopLocation().GetID() == "" {
		return nil, fmt.Errorf("API returned incomplete site or PoP identity")
	}
	return item, nil
}
func (r *connectionResource) slotOccupied(ctx context.Context, input m.AddCloudInterconnectPhysicalConnectionInput) (bool, error) {
	result, err := r.getClient().SiteCloudInterconnectPhysicalConnectionID(ctx, r.client.AccountId,
		m.CloudInterconnectPhysicalConnectionIDInput{Site: input.Site, HaRole: input.HaRole})
	if err != nil {
		if missingConnection(err) {
			return false, nil
		}
		return false, err
	}
	if result.GetSite().GetCloudInterconnectPhysicalConnectionID().GetID() == "" {
		return false, fmt.Errorf("slot lookup returned no identity or not-found error")
	}
	return true, nil
}
func missingConnection(err error) bool {
	return apperror.MatchesGraphQLMessages(err, "connection not found", "GQL-User-Error: connection not found")
}

const consistencyTimeout = 30 * time.Second
const consistencyInterval = 500 * time.Millisecond

func (r *connectionResource) waitForConnection(
	ctx context.Context, id string, input m.AddCloudInterconnectPhysicalConnectionInput,
) (*connectionRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, consistencyTimeout)
	defer cancel()
	ticker := time.NewTicker(consistencyInterval)
	defer ticker.Stop()
	for {
		item, err := r.find(ctx, id)
		if err != nil {
			return nil, err
		}
		if item != nil && matchesConnection(item, input) {
			return item, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for connection read convergence: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
