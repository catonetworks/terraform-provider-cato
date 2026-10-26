package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestPlatformValidator(t *testing.T) {
	t.Parallel()
	checkSetValidator(t, PlatformValidator{}, enumStrings(cato_models.AllOperatingSystem))
}
