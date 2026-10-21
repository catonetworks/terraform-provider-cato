package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestIPProtocolValidator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, IPProtocolValidator{}, enumStrings(cato_models.AllIPProtocol), false)
}
