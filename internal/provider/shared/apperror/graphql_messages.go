package apperror

import (
	"errors"
	"slices"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
)

// MatchesGraphQLMessages requires every structured API error to match an exact
// allowed message. Transport failures, empty lists and mixed errors never prove absence.
func MatchesGraphQLMessages(err error, allowed ...string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item == nil || !slices.Contains(allowed, item.Message) {
			return false
		}
	}
	return true
}
