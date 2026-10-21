package validators

import (
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestSocketInterfaceIndexValidator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, SocketInterfaceIndexValidator{}, enumStrings(cato_models.AllSocketInterfaceIDEnum), true)
}
