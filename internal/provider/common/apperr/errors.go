package apperr

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type APIErrors interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

// ErrConfig generic configuration error
var ErrConfig = errors.New("configuration error")

func CheckAPIErrors[T APIErrors](err error, errs []T, summary string, diags *diag.Diagnostics) bool {
	if err != nil {
		diags.AddError(summary, err.Error())
		return true
	}
	if len(errs) > 0 {
		for _, e := range errs {
			if msg := e.GetErrorMessage(); msg != nil {
				diags.AddError(summary, *msg)
			} else if code := e.GetErrorCode(); code != nil {
				diags.AddError(summary, *code)
			} else {
				diags.AddError(summary, "API mutation failed without error details")
			}
		}
		return true
	}
	return false
}

func CheckErr(diags *diag.Diagnostics, in diag.Diagnostics) bool {
	diags.Append(in...)
	return diags.HasError()
}
