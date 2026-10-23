package rule

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	domain "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

// Deletion translates workflow failures only at the Terraform boundary. Raw
// SDK causes may contain credentials or policy data and are never rendered.
func Deletion(err error) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if err == nil {
		return diagnostics
	}
	var invalid *domain.ValidationError
	if errors.As(err, &invalid) && invalid.Field == "id" {
		diagnostics.AddAttributeError(path.Root("id"), "Missing private-access rule ID", "Deletion requires a known, nonempty rule ID in state.")
		return diagnostics
	}
	detail := failureDetail(err)
	var failure *domain.DeleteError
	if errors.As(err, &failure) {
		detail = "Deletion failed during the " + string(failure.Stage) + " step. " + detail
		if failure.RemovalCompleted {
			detail += " Removal completed in the editable revision, but published deletion has not been confirmed."
		}
	}
	diagnostics.AddError("Unable to delete private-access rule", detail)
	return diagnostics
}

func failureDetail(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "The operation was canceled."
	case errors.Is(err, context.DeadlineExceeded):
		return "The operation exceeded its deadline."
	case errors.Is(err, domain.ErrStillPublished):
		return "The rule is still present in the published policy."
	case errors.Is(err, domain.ErrInvalidResponse):
		return "Cato returned an incomplete or invalid response."
	case errors.Is(err, domain.ErrRevisionConflict):
		return "The policy revision remained busy after bounded retries."
	case errors.Is(err, domain.ErrRejected):
		return "Cato rejected the operation. Check API permissions and policy revision conflicts."
	case errors.Is(err, domain.ErrUnavailable):
		return "The API request failed and the deletion outcome could not be confirmed."
	default:
		return "The deletion could not be completed. The resource remains managed by Terraform."
	}
}
