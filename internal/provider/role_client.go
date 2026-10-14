package provider

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RoleManagementClient is the minimal SDK contract used by IAM role resources and data sources.
type RoleManagementClient interface {
	RbacRoleManagementCreateRole(context.Context, string, cato_models.RoleManagementCreateRoleInput,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementCreateRole, error)
	RbacRoleManagementUpdateRole(context.Context, string, cato_models.RoleManagementUpdateRoleInput,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementUpdateRole, error)
	RbacRoleManagementDeleteRole(context.Context, string, cato_models.RoleManagementDeleteRoleInput,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementDeleteRole, error)
	RbacRoleManagementRole(context.Context, string, string,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementRole, error)
	RbacRoleManagementRoleList(context.Context, string, cato_models.RoleManagementRoleListInput,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementRoleList, error)
	RbacRoleManagementPermissionCatalog(context.Context, string,
		...clientv2.RequestInterceptor,
	) (*cato.RbacRoleManagementPermissionCatalog, error)
}

type roleClientConfig struct {
	client     *catoClientData
	roleClient RoleManagementClient
}

func (c *roleClientConfig) configure(data any, diags *diag.Diagnostics) {
	if data == nil {
		return
	}
	client, ok := data.(*catoClientData)
	if !ok {
		diags.AddError("Unexpected provider configuration", fmt.Sprintf("Expected Cato provider configuration, received %T.", data))
		return
	}
	c.client = client
}
func (c *roleClientConfig) getRoleClient() RoleManagementClient {
	if c.roleClient != nil {
		return c.roleClient
	}
	if c.client == nil || c.client.catov2 == nil {
		return nil
	}
	return c.client.catov2
}

func lookupRole(ctx context.Context, client RoleManagementClient, accountID, roleID string) (*Role, diag.Diagnostics) {
	var diags diag.Diagnostics
	if client == nil {
		diags.AddError("Unconfigured role client", "Configure the Cato provider before reading roles.")
		return nil, diags
	}
	result, err := client.RbacRoleManagementRole(ctx, accountID, roleID)
	if err != nil {
		diags.AddError("Failed to read role", err.Error())
		return nil, diags
	}
	if result == nil || result.GetRbac() == nil || result.GetRbac().GetRoleManagement() == nil {
		diags.AddError("Invalid role response", "The API returned no role-management namespace.")
		return nil, diags
	}
	namespace := result.GetRbac().GetRoleManagement()
	if namespace.GetTypename() == nil || *namespace.GetTypename() != "RoleManagementQueries" {
		diags.AddError("Invalid role response", "The API returned no role-management namespace marker.")
		return nil, diags
	}
	value := namespace.GetRole()
	if value == nil {
		return nil, diags
	}
	model, d := roleFromAPI(ctx, accountID, value)
	diags.Append(d...)
	if model.RoleID.ValueString() != roleID {
		diags.AddError("Invalid role response", "The API returned a different role identifier than requested.")
	}
	return &model, diags
}
