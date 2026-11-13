package privateapp

import (
	"encoding/json"
	"errors"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/require"
)

func TestPrivateAppNotFoundRequiresLeafError(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		want          bool
	}{
		{"leaf", `{"graphqlErrors":[{"extensions":{"code":"NOT_FOUND"},"path":["privateApplication","privateApplication"]}]}`, true},
		{"account", `{"graphqlErrors":[{"extensions":{"code":"NOT_FOUND"},"path":["privateApplication"]}]}`, false},
		{"wrong code", `{"graphqlErrors":[{"extensions":{"code":"FORBIDDEN"},"path":["privateApplication","privateApplication"]}]}`, false},
		{"mixed", `{"graphqlErrors":[{"extensions":{"code":"NOT_FOUND"},"path":["privateApplication","privateApplication"]},{"message":"denied"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response clientv2.ErrorResponse
			require.NoError(t, json.Unmarshal([]byte(tc.payload), &response))
			require.Equal(t, tc.want, privateAppNotFound(&response))
		})
	}
	require.False(t, privateAppNotFound(errors.New("account NOT_FOUND")))
}
