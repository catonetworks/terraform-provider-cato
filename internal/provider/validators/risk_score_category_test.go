package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestRiscScoreCategory(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, RiscScoreCategory{}, enumStrings(cato_models.AllRiskScoreCategory), false)
}
