//go:build acctest

package dns_dhcp

import "testing"

func TestDNSDHCPEnvironmentGuard(t *testing.T) {
	for _, test := range []struct {
		name, endpoint, account, allowed string
		want                             bool
	}{
		{"legacy environment", "https://system.cc.dev-us1-m1.catonet.works/api/v1/graphql", "123", "", true},
		{"dedicated HTTPS account", "https://api.example.test/api/v1/graphql2", "123", "123", true},
		{"missing opt in", "https://api.example.test/api/v1/graphql2", "123", "", false},
		{"different account", "https://api.example.test/api/v1/graphql2", "123", "456", false},
		{"empty account", "https://api.example.test/api/v1/graphql2", "", "", false},
		{"insecure endpoint", "http://api.example.test/graphql", "123", "123", false},
		{"credential in URL", "https://user:password@example.test/graphql", "123", "123", false},
		{"missing host", "https:///graphql", "123", "123", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := dnsDHCPEnvironmentAllowed(test.endpoint, test.account, test.allowed); got != test.want {
				t.Fatalf("allowed=%t, want %t", got, test.want)
			}
		})
	}
}
