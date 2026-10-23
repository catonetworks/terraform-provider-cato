package rbacassignment

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	sdk "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

// ResourceClient contains only the assignment operations used by this resource.
type ResourceClient interface {
	RbacAddRbacAssignments(context.Context,
		string,
		m.AddRbacAssignmentsInput,
		...clientv2.RequestInterceptor) (*sdk.RbacAddRbacAssignments,
		error)
	RbacUpdateRbacAssignment(context.Context,
		string,
		m.UpdateRbacAssignmentInput,
		...clientv2.RequestInterceptor) (*sdk.RbacUpdateRbacAssignment,
		error)
	RbacRemoveRbacAssignment(context.Context,
		string,
		m.RemoveRbacAssignmentInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRemoveRbacAssignment,
		error)
	RbacRbacAssignmentList(context.Context,
		string,
		m.RBACAssignmentListInput,
		...clientv2.RequestInterceptor) (*sdk.RbacRbacAssignmentList,
		error)
}

type sdkClient struct{ *sdk.Client }

var _ ResourceClient = (*sdkClient)(nil)

func (r *assignmentResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{r.client.V2()}
}

func lookupAssignment(ctx context.Context, client ResourceClient, account, id string) (*assignmentRecord, error) {
	if client == nil {
		return nil, fmt.Errorf("configure the provider before reading assignments")
	}
	seen := map[string]bool{}
	total := int64(-1)
	for pageNumber := int64(0); ; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		response, err := client.RbacRbacAssignmentList(ctx, account, m.RBACAssignmentListInput{
			Paging: &m.PagingInput{From: pageNumber, Limit: iam.PageSize},
			Sort:   &m.RBACAssignmentSortInput{CreationDate: &m.SortOrderInput{Direction: m.SortOrderAsc}},
		})
		if err != nil {
			return nil, err
		}
		page := response.GetRbac().GetRbacAssignmentList()
		if page == nil {
			return nil, fmt.Errorf("API returned no assignment page")
		}
		if err := iam.ValidatePage(page.Paging.Typename, page.Items != nil, page.Paging.Total, total); err != nil {
			return nil, err
		}
		total = page.Paging.Total
		found, err := findInPage(page.Items, seen, account, id)
		if err != nil {
			return nil, err
		}
		if int64(len(seen)) > total {
			return nil, fmt.Errorf("assignment page exceeds reported total")
		}
		if found != nil {
			return found, nil
		}
		if int64(len(seen)) == total {
			return nil, nil
		}
		if len(page.Items) == 0 {
			return nil, fmt.Errorf("assignment pagination made no progress")
		}
	}
}

func findInPage(items []*sdk.RbacRbacAssignmentList_Rbac_RbacAssignmentList_Items,
	seen map[string]bool, account, id string) (*assignmentRecord, error) {
	var found *assignmentRecord
	for _, item := range items {
		value := listedRecord(item)
		if value == nil || value.ID == "" || value.Account != account || seen[value.ID] {
			return nil, fmt.Errorf("API returned an invalid, repeated, or cross-account assignment")
		}
		seen[value.ID] = true
		if value.ID == id {
			found = value
		}
	}

	return found, nil
}
