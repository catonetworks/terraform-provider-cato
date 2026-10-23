package role

import (
	"context"
	"fmt"
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
