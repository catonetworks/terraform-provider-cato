package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestDateTimeValidator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, DateTimeValidator{}, []string{"2026-10-21T12:30:00Z"}, false)
	for _, s := range []string{"2026-10-21", "2026-10-21T12:30:00+02:00", "2026-10-21T12:30:00.123Z", "2026-10-21T12:30:00Z trailing"} {
		t.Run(s, func(t *testing.T) {
			resp := &validator.StringResponse{}
			DateTimeValidator{}.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(s)}, resp)
			require.True(t, resp.Diagnostics.HasError())
		})
	}
}
