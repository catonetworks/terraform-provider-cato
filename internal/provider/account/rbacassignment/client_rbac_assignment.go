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
	// There is no public assignment-ID filter. Restart offset pagination when
	// concurrent mutations change its total, rather than interpreting skipped rows
	// as a missing assignment. Bound retries so a busy account fails safely.
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		value, changed, err := scanAssignments(ctx, client, account, id)
		if err != nil || !changed {
			return value, err
		}
	}
	return nil, fmt.Errorf("assignment list changed during pagination after %d attempts; retry the operation", maxAttempts)
}

func scanAssignments(ctx context.Context, client ResourceClient, account, id string) (*assignmentRecord, bool, error) {
	seen := map[string]bool{}
	total := int64(-1)
	for pageNumber := int64(0); ; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		// The IAM service dereferences filter even when no criteria are supplied.
		response, err := client.RbacRbacAssignmentList(ctx, account, m.RBACAssignmentListInput{
			Filter: &m.RBACAssignmentFilterInput{},
			Paging: &m.PagingInput{From: pageNumber, Limit: iam.PageSize},
			Sort:   &m.RBACAssignmentSortInput{CreationDate: &m.SortOrderInput{Direction: m.SortOrderAsc}},
		})
		if err != nil {
			return nil, false, err
		}
		page := response.GetRbac().GetRbacAssignmentList()
		if page == nil {
			return nil, false, fmt.Errorf("API returned no assignment page")
		}
		if err := iam.ValidatePage(page.Paging.Typename, page.Items != nil, page.Paging.Total, -1); err != nil {
			return nil, false, err
		}
		if total >= 0 && total != page.Paging.Total {
			return nil, true, nil
		}
		total = page.Paging.Total
		found, err := findInPage(page.Items, seen, account, id)
		if err != nil {
			return nil, false, err
		}
		if int64(len(seen)) > total {
			return nil, false, fmt.Errorf("assignment page exceeds reported total")
		}
		if found != nil {
			return found, false, nil
		}
		if int64(len(seen)) == total {
			return nil, false, nil
		}
		if len(page.Items) == 0 {
			return nil, false, fmt.Errorf("assignment pagination made no progress")
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
