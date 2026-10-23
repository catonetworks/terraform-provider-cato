package role

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

// ResourceClient is the account-scoped role CRUD contract.
type ResourceClient interface {
	iam.RoleLookup
	RbacRoleManagementCreateRole(context.Context,
		string,
		m.RoleManagementCreateRoleInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementCreateRole,
		error)
	RbacRoleManagementUpdateRole(context.Context,
		string,
		m.RoleManagementUpdateRoleInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementUpdateRole,
		error)
	RbacRoleManagementDeleteRole(context.Context,
		string,
		m.RoleManagementDeleteRoleInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementDeleteRole,
		error)
}

type sdkClient struct{ *sdk.Client }

var _ ResourceClient = (*sdkClient)(nil)

func (r *roleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{r.client.V2()}
}

const (
	roleConsistencyTimeout  = 30 * time.Second
	roleConsistencyInterval = 250 * time.Millisecond
)

// waitForRole observes replica convergence without retrying a mutation.
func waitForRole(ctx context.Context, client ResourceClient, account string, expected *iam.RoleRecord) error {
	ctx, cancel := context.WithTimeout(ctx, roleConsistencyTimeout)
	defer cancel()
	timer := time.NewTicker(roleConsistencyInterval)
	defer timer.Stop()
	for {
		actual, err := iam.LookupRole(ctx, client, account, expected.ID)
		if err != nil {
			return err
		}
		if sameMutableRole(actual, expected) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for role read replica: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// deleteRoleWithRetry tolerates temporary IAM/CC2 outages without retrying create
// operations. Before repeating deletion, confirm the role still exists.
func deleteRoleWithRetry(ctx context.Context, client ResourceClient, account, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, roleConsistencyTimeout)
	defer cancel()
	for {
		response, err := client.RbacRoleManagementDeleteRole(ctx, account, m.RoleManagementDeleteRoleInput{ID: id})
		if err == nil {
			return response.GetRbac().GetRoleManagement().GetDeleteRole().GetID() == id, nil
		}
		if !roleDeleteUnavailable(err) {
			return false, err
		}
		timer := time.NewTimer(roleConsistencyInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false, fmt.Errorf("waiting to delete role: %w", errors.Join(ctx.Err(), err))
		case <-timer.C:
		}
		role, readErr := iam.LookupRole(ctx, client, account, id)
		if readErr != nil {
			return false, errors.Join(err, readErr)
		}
		if role == nil {
			return true, nil
		}
	}
}

func roleDeleteUnavailable(err error) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item == nil || item.Path.String() != "rbac.roleManagement.deleteRole" || item.Extensions["code"] != "DOWNSTREAM_SERVICE_ERROR" {
			return false
		}
		detail, _ := item.Extensions["full"].(string)
		if item.Extensions["errorType"] != "UNAVAILABLE" && !strings.Contains(detail, "Role management is temporarily unavailable") {
			return false
		}
	}
	return true
}
