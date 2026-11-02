package apperror

import "regexp"

var numericApplicationID = regexp.MustCompile(`^\d+$`)

// MissingCustomApplication matches the account-qualified lookup key used for numeric IDs.
// Use the same exact matching for resource refresh and acceptance destroy checks.
func MissingCustomApplication(err error, id, account string) bool {
	lookup := id
	if numericApplicationID.MatchString(id) {
		lookup = "CustomApp_" + account + "_" + id
	}
	expected := "Custom application not found by ID: " + lookup
	return MatchesGraphQLMessages(err, expected, "GQL-User-Error: "+expected)
}
