package static_host

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	domain "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

func Creation(err error) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if err == nil {
		return diagnostics
	}
	var invalid *domain.ValidationError
	if errors.As(err, &invalid) {
		diagnostics.AddAttributeError(path.Root(invalid.Field), "Invalid static-host configuration", invalid.Field+" must contain a valid value.")
		return diagnostics
	}
	detail := "The API request failed. Creation may have completed; check the site before retrying."
	switch {
	case errors.Is(err, context.Canceled):
		detail = "Creation was canceled. Check the site before retrying."
	case errors.Is(err, context.DeadlineExceeded):
		detail = "Creation exceeded its deadline. Check the site before retrying."
	case errors.Is(err, domain.ErrInvalidResponse):
		detail = "Cato did not return a valid host ID. Check the site before retrying."
	}
	diagnostics.AddError("Unable to create static host", detail)
	return diagnostics
}
