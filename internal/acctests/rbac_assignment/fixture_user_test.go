//go:build acctest

package rbacassignment

import (
	"context"
	"testing"
	"time"

	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/stretchr/testify/require"
)

func createAssignmentTestUser(t *testing.T, client *sdk.Client, account, name string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := client.UserCreateUser(ctx, account, m.CreateUserInput{
		FirstName: "Terraform", LastName: name, Email: name + "@example.invalid",
	})
	require.NoError(t, err)
	user := response.GetUser().GetCreateUser().GetUser()
	require.NotNil(t, user, "API did not return the created test user")
	require.NotEmpty(t, user.ID, "API did not return the created test user ID")
	id := user.ID
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		deleted, deleteErr := client.UserDeleteUser(cleanupCtx, account, m.DeleteUserInput{UserID: []scalars.Long{scalars.Long(id)}})
		if deleteErr != nil {
			t.Errorf("delete assignment test user %s: %v", id, deleteErr)
			return
		}
		status := deleted.GetUser().GetDeleteUser().GetStatus()
		if status == nil || *status != m.DeleteUserStatusSuccess {
			t.Errorf("API did not confirm deletion of assignment test user %s", id)
		}
	})
	return id
}
