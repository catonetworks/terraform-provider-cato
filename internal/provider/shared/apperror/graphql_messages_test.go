package apperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/require"
)

func TestExactGraphQLMessages(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{`{"graphqlErrors":[{"message":"owned missing"}]}`, true},
		{`{"graphqlErrors":[{"message":"owned missing"},{"message":"forbidden"}]}`, false},
		{`{"graphqlErrors":[null]}`, false}, {`{"graphqlErrors":[]}`, false},
		{`{"networkErrors":{"code":503,"message":"unavailable"},"graphqlErrors":[{"message":"owned missing"}]}`, false},
	} {
		var response clientv2.ErrorResponse
		require.NoError(t, json.Unmarshal([]byte(tc.body), &response))
		require.Equal(t, tc.want, MatchesGraphQLMessages(fmt.Errorf("wrapped: %w", &response), "owned missing"))
	}
	require.False(t, MatchesGraphQLMessages(errors.New("owned missing"), "owned missing"))
	require.False(t, MatchesGraphQLMessages(nil, "owned missing"))
}
