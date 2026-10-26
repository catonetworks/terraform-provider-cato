package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestDaysValidator(t *testing.T) {
	t.Parallel()
	checkSetValidator(t, DaysValidator{}, enumStrings(cato_models.AllDayOfWeek))
}
