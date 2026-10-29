package vsocket

import (
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

func TestParentSnapshot(t *testing.T) {
	var snapshot cato.AccountSnapshot
	require.NoError(t, json.Unmarshal([]byte(`{"accountSnapshot":{"id":"account","sites":[{"id":"site","infoSiteSnapshot":{"connType":"SOCKET_GCP1500","sockets":[{"id":"primary","isPrimary":true}]}}]}}`), &snapshot))
	require.NoError(t, ValidateEmptySlot(&snapshot, "account", "site", "GCP", m.ProtoTypeSocketGCP1500))
	require.Error(t, ValidateEmptySlot(&snapshot, "other", "site", "GCP", m.ProtoTypeSocketGCP1500))
	require.Error(t, ValidateEmptySlot(&snapshot, "account", "other", "GCP", m.ProtoTypeSocketGCP1500))
	require.Error(t, ValidateEmptySlot(&snapshot, "account", "site", "AWS", m.ProtoTypeVsocketVgxAWS))
	require.Error(t, ValidateEmptySlot(nil, "account", "site", "GCP", m.ProtoTypeSocketGCP1500))
}
