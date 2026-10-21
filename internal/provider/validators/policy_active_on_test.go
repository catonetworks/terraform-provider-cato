package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestPolicyActiveOnValidator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, PolicyActiveOnValidator{}, enumStrings(cato_models.AllPolicyActiveOnEnum), false)
}
