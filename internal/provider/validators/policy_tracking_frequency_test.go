package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestPolicyTrackingFrequency(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, PolicyTrackingFrequency{}, enumStrings(cato_models.AllPolicyRuleTrackingFrequencyEnum), false)
}
