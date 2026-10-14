package catoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
)

// Generated SDK namespace structs lose the distinction between null and an absent
// leaf. Validate wire presence before unmarshalling only standalone read responses.
type networkConfigReadClient struct{ clientv2.HttpClient }

func (c *networkConfigReadClient) Do(req *http.Request) (*http.Response, error) {
	domain, leaf := networkConfigReadPath(req)
	response, err := c.HttpClient.Do(req)
	if err != nil || domain == "" || response.StatusCode != http.StatusOK {
		return response, err
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	if err := validateNetworkConfigRead(body, domain, leaf); err != nil {
		return nil, err
	}
	return response, nil
}

func networkConfigReadPath(req *http.Request) (domain, leaf string) {
	if req.GetBody == nil {
		return "", ""
	}
	body, err := req.GetBody()
	if err != nil {
		return "", ""
	}
	defer func() { _ = body.Close() }()
	var request struct {
		OperationName string `json:"operationName"`
	}
	if json.NewDecoder(body).Decode(&request) != nil {
		return "", ""
	}
	switch request.OperationName {
	case "networkConfigDnsSettings":
		return "dns", "settings"
	case "networkConfigDnsSiteSettings":
		return "dns", "siteSettings"
	case "networkConfigDnsForwardingRule":
		return "dns", "forwardingRule"
	case "networkConfigDhcpSettings":
		return "dhcp", "settings"
	case "networkConfigDhcpOption":
		return "dhcp", "option"
	case "networkConfigDhcpRelayGroup":
		return "dhcp", "relayGroup"
	default:
		return "", ""
	}
}

func validateNetworkConfigRead(body []byte, domain, leaf string) error {
	var response struct {
		Data   map[string]json.RawMessage `json:"data"`
		Errors []json.RawMessage          `json:"errors"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	// Preserve structured authorization, rollout and NOT_FOUND errors for SDK handling.
	if len(response.Errors) > 0 {
		return nil
	}
	var namespaces map[string]json.RawMessage
	if err := json.Unmarshal(response.Data["networkConfig"], &namespaces); err != nil {
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
