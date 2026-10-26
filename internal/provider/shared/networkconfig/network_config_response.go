package networkconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/gqlgo/gqlgenc/graphqljson"
)

// Generated namespace structs collapse null namespaces into absent leaves.
// Domain adapters validate the raw projection before decoding SDK result types.
func validateReadResponse(domain, leaf string) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, info *clientv2.GQLRequestInfo, result any,
		next clientv2.RequestInterceptorFunc) error {
		var projection struct {
			NetworkConfig json.RawMessage `json:"networkConfig"`
		}
		if err := next(ctx, req, info, &projection); err != nil {
			return err
		}
		if err := validateNetworkConfigProjection(projection.NetworkConfig, domain, leaf); err != nil {
			return err
		}
		data, err := json.Marshal(projection)
		if err != nil {
			return err
		}
		return graphqljson.UnmarshalData(data, result)
	}
}

func validateNetworkConfigProjection(projection json.RawMessage, domain, leaf string) error {
	var namespaces map[string]json.RawMessage
	if err := json.Unmarshal(projection, &namespaces); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(namespaces[domain], &fields); err != nil {
		return err
	}
	if _, present := fields[leaf]; !present {
		return fmt.Errorf("invalid network configuration response: missing %s.%s; state retained", domain, leaf)
	}
	return nil
}
