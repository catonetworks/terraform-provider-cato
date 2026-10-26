package iam

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
)

// RoleLookup is the single-role query contract.
type RoleLookup interface {
	RbacRoleManagementReadRole(context.Context, string, string, ...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementReadRole, error)
}

// RoleReader contains the read operations used by IAM data sources.
type RoleReader interface {
	RoleLookup
	RbacRoleManagementListRoles(context.Context,
		string,
		m.RoleManagementRoleListInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementListRoles,
		error)
	RbacRoleManagementPermissionCatalog(context.Context,
		string,
		...clientv2.RequestInterceptor) (*sdk.RbacRoleManagementPermissionCatalog,
		error)
}

// LookupRole distinguishes a missing role from an incomplete namespace response.
func LookupRole(ctx context.Context, client RoleLookup, account, id string) (*RoleRecord, error) {
	if client == nil {
		return nil, fmt.Errorf("configure the provider before reading roles")
	}
	response, err := client.RbacRoleManagementReadRole(ctx, account, id)
	if err != nil {
		return nil, err
	}
	if response == nil || response.Rbac.Typename == nil || response.Rbac.RoleManagement.Typename == nil {
		return nil, fmt.Errorf("API returned an incomplete role namespace")
	}
	value := ReadRoleRecord(response.Rbac.RoleManagement.Role)
	if response.Rbac.RoleManagement.Role != nil && value == nil {
		return nil, fmt.Errorf("API returned a malformed role")
	}
	if value != nil && value.ID != id {
		return nil, fmt.Errorf("API returned a different role identifier")
	}
	return value, nil
}

// ListRoles follows the role API's offset pagination and rejects non-progressing results.
func ListRoles(ctx context.Context, client RoleReader, account string, input m.RoleManagementRoleListInput) ([]*RoleRecord, error) {
	result := make([]*RoleRecord, 0)
	seen := map[string]bool{}
	total := int64(-1)
	for offset := int64(0); ; {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		input.Paging = &m.PagingInput{From: offset, Limit: PageSize}
		response, err := client.RbacRoleManagementListRoles(ctx, account, input)
		if err != nil {
			return nil, err
		}
		page := response.GetRbac().GetRoleManagement().GetRoleList()
		if page == nil {
			return nil, fmt.Errorf("API returned no role page")
		}
		if err := ValidatePage(page.Paging.Typename, page.Items != nil, page.Paging.Total, total); err != nil {
			return nil, err
		}
		total = page.Paging.Total
		for _, item := range page.Items {
			value := ListRoleRecord(item)
			if value == nil || value.ID == "" || seen[value.ID] {
				return nil, fmt.Errorf("API returned an invalid or repeated role")
			}
			seen[value.ID] = true
			result = append(result, value)
		}
		count := int64(len(result))
		if count > total {
			return nil, fmt.Errorf("role page exceeds reported total")
		}
		if count == total {
			return result, nil
		}
		if len(page.Items) == 0 {
			return nil, fmt.Errorf("role pagination made no progress")
		}
		offset = count
	}
}
