package iam

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PageSize bounds IAM list requests.
const PageSize int64 = 100

// Account resolves an optional account override without silently accepting unknown values.
func Account(value types.String, fallback string) (string, error) {
	if value.IsUnknown() {
		return "", fmt.Errorf("account_id must be known")
	}
	if !value.IsNull() {
		fallback = value.ValueString()
	}
	if strings.TrimSpace(fallback) == "" {
		return "", fmt.Errorf("configure account_id in the resource or provider")
	}
	return fallback, nil
}

// ParseID validates an account-scoped Terraform import identifier.
func ParseID(value string) (accountID, objectID string, err error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" ||
		strings.TrimSpace(parts[0]) != parts[0] || strings.TrimSpace(parts[1]) != parts[1] {
		return "", "", fmt.Errorf("expected ACCOUNT_ID:OBJECT_ID with two nonempty identifiers")
	}
	return parts[0], parts[1], nil
}

// ValidatePage rejects missing metadata and inconsistent totals before interpreting absence.
func ValidatePage(marker *string, itemsPresent bool, total, previous int64) error {
	if marker == nil || *marker != "PageInfo" || !itemsPresent || total < 0 {
		return fmt.Errorf("API returned an incomplete page")
	}
	if previous >= 0 && previous != total {
		return fmt.Errorf("API total changed during pagination")
	}
	return nil
}
