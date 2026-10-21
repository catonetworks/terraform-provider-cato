package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestRiskScoreOperator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, RiskScoreOperator{}, enumStrings(cato_models.AllRiskScoreOperator), false)
}
