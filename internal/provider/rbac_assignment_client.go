package provider

import (
	"context"
	"fmt"

	rbacapi "github.com/catonetworks/terraform-provider-cato/internal/provider/rbacassignmentapi"
)

// RBACAssignmentClient is the account-scoped assignment API contract.
type RBACAssignmentClient interface {
	AddAssignment(context.Context, string, rbacapi.RBACAssignmentInput) (*rbacapi.RBACAssignmentAPI, error)
	UpdateAssignment(context.Context, string, rbacapi.RBACAssignmentInput) (*rbacapi.RBACAssignmentAPI, error)
	ListAssignments(context.Context, string, int) (*rbacapi.RBACAssignmentPage, error)
	RemoveAssignment(context.Context, string, string) error
}

const assignmentPageSize = 100

const assignmentSelection = `id account { id } user { id } group { id } assignmentType
 roles { id } allowedEntities { id type action } allowAllSites allowAllUsers dataObfuscationStatus description`
const addAssignmentDocument = `mutation addRbacAssignments($accountId: ID!, $input: AddRbacAssignmentsInput!) {
 rbac(accountId: $accountId) { addRbacAssignments(input: $input) { items { ` + assignmentSelection + ` } } }
}`
const updateAssignmentDocument = `mutation updateRbacAssignment($accountId: ID!, $input: UpdateRbacAssignmentInput!) {
 rbac(accountId: $accountId) { updateRbacAssignment(input: $input) { assignment { ` + assignmentSelection + ` } } }
}`
const listAssignmentsDocument = `query rbacAssignmentList($accountId: ID!, $input: RBACAssignmentListInput!) {
 rbac(accountId: $accountId) { rbacAssignmentList(input: $input) { items { ` + assignmentSelection + ` } paging { total } } }
}`
const removeAssignmentDocument = `mutation removeRbacAssignment($accountId: ID!, $input: RemoveRbacAssignmentInput!) {
 rbac(accountId: $accountId) { removeRbacAssignment(input: $input) { rbacAssignmentId } }
}`

// These provider-owned operations use the SDK's authenticated transport until the
// generated SDK exposes the assignment operations. Recheck them on SDK upgrades.
func (c *providerSDKClient) AddAssignment(
	ctx context.Context, account string, input rbacapi.RBACAssignmentInput,
) (*rbacapi.RBACAssignmentAPI, error) {
	var result struct {
		Rbac *struct {
			Payload *struct {
				Items []*rbacapi.RBACAssignmentAPI `json:"items" graphql:"items"`
			} `json:"addRbacAssignments" graphql:"addRbacAssignments"`
		} `json:"rbac" graphql:"rbac"`
	}
	// Creation sends both arrays explicitly; one is empty and the other contains
	// exactly one principal, so a Terraform resource owns one assignment.
	variables := assignmentScopeVariables(input)
	variables["users"] = input.Users
	variables["groups"] = input.Groups
	err := c.Client.Client.Post(ctx, "addRbacAssignments", addAssignmentDocument, &result,
		map[string]any{"accountId": account, "input": variables},
	)
	if err != nil {
		return nil, err
	}
	if result.Rbac == nil || result.Rbac.Payload == nil || len(result.Rbac.Payload.Items) != 1 || result.Rbac.Payload.Items[0] == nil {
		return nil, fmt.Errorf("API did not return exactly one created assignment")
	}
	return result.Rbac.Payload.Items[0], nil
}
func (c *providerSDKClient) UpdateAssignment(
	ctx context.Context, account string, input rbacapi.RBACAssignmentInput,
) (*rbacapi.RBACAssignmentAPI, error) {
	var result struct {
		Rbac *struct {
			Payload *struct {
				Assignment *rbacapi.RBACAssignmentAPI `json:"assignment" graphql:"assignment"`
			} `json:"updateRbacAssignment" graphql:"updateRbacAssignment"`
		} `json:"rbac" graphql:"rbac"`
	}
	variables := assignmentScopeVariables(input)
	variables["rbacAssignmentId"] = input.AssignmentID
	err := c.Client.Client.Post(ctx, "updateRbacAssignment", updateAssignmentDocument, &result,
		map[string]any{"accountId": account, "input": variables},
	)
	if err != nil {
		return nil, err
	}
	if result.Rbac == nil || result.Rbac.Payload == nil || result.Rbac.Payload.Assignment == nil {
		return nil, fmt.Errorf("API returned no updated assignment")
	}
	return result.Rbac.Payload.Assignment, nil
}
func (c *providerSDKClient) ListAssignments(ctx context.Context, account string, page int) (*rbacapi.RBACAssignmentPage, error) {
	var result struct {
		Rbac *struct {
			Payload *rbacapi.RBACAssignmentPage `json:"rbacAssignmentList" graphql:"rbacAssignmentList"`
		} `json:"rbac" graphql:"rbac"`
	}
	input := map[string]any{
		"paging": map[string]int{"from": page, "limit": assignmentPageSize},
		"sort":   map[string]any{"creationDate": map[string]string{"direction": "ASC"}},
	}
	err := c.Client.Client.Post(ctx, "rbacAssignmentList", listAssignmentsDocument, &result,
		map[string]any{"accountId": account, "input": input},
	)
	if err != nil {
		return nil, err
	}
	if result.Rbac == nil || result.Rbac.Payload == nil {
		return nil, fmt.Errorf("API returned no assignment list")
	}
	return result.Rbac.Payload, nil
}
func (c *providerSDKClient) RemoveAssignment(ctx context.Context, account, id string) error {
	var result struct {
		Rbac *struct {
			Payload *struct {
				ID string `json:"rbacAssignmentId" graphql:"rbacAssignmentId"`
			} `json:"removeRbacAssignment" graphql:"removeRbacAssignment"`
		} `json:"rbac" graphql:"rbac"`
	}
	err := c.Client.Client.Post(ctx, "removeRbacAssignment", removeAssignmentDocument, &result,
		map[string]any{"accountId": account, "input": map[string]string{"rbacAssignmentId": id}},
	)
	if err != nil {
		return err
	}
	if result.Rbac == nil || result.Rbac.Payload == nil || result.Rbac.Payload.ID != id {
		return fmt.Errorf("API did not confirm deletion of the requested assignment")
	}
	return nil
}

func lookupAssignment(ctx context.Context, client RBACAssignmentClient, account, id string) (*rbacapi.RBACAssignmentAPI, error) {
	if client == nil {
		return nil, fmt.Errorf("configure the Cato provider before managing assignments")
	}
	if id == "" {
		return nil, fmt.Errorf("assignment_id must not be empty")
	}
	seen := map[string]bool{}
	total := -1
	for page := 0; ; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := client.ListAssignments(ctx, account, page)
		if err != nil {
			return nil, err
		}
		if err := validateAssignmentPage(result, total); err != nil {
			return nil, err
		}
		total = result.Paging.Total
		found, err := findAssignmentInPage(result.Items, seen, account, id)
		if err != nil {
			return nil, err
		}
		if len(seen) > total {
			return nil, fmt.Errorf("API returned more assignments than its reported total")
		}
		if found != nil {
			return found, nil
		}
		if len(seen) == total {
			return nil, nil
		}
		if len(result.Items) == 0 {
			return nil, fmt.Errorf("assignment pagination made no progress")
		}
	}
}

func findAssignmentInPage(items []*rbacapi.RBACAssignmentAPI, seen map[string]bool, account, id string,
) (*rbacapi.RBACAssignmentAPI, error) {
	var found *rbacapi.RBACAssignmentAPI
	for _, item := range items {
		if item == nil || item.ID == "" || item.Account == nil || item.Account.ID != account {
			return nil, fmt.Errorf("API returned an invalid or cross-account assignment")
		}
		if seen[item.ID] {
			return nil, fmt.Errorf("API repeated an assignment during pagination")
		}
		seen[item.ID] = true
		if item.ID == id {
			found = item
		}
	}
	return found, nil
}

func validateAssignmentPage(result *rbacapi.RBACAssignmentPage, total int) error {
	if result == nil || result.Paging == nil || result.Items == nil || result.Paging.Total < 0 {
		return fmt.Errorf("API returned an incomplete assignment page")
	}
	if total >= 0 && total != result.Paging.Total {
		return fmt.Errorf("assignment total changed during pagination; retry after concurrent changes finish")
	}
	return nil
}

// Explicit maps keep create-only fields out of update input and preserve omitted
// optional preferences with the SDK GraphQL variable marshaler.
func assignmentScopeVariables(input rbacapi.RBACAssignmentInput) map[string]any {
	variables := map[string]any{
		"roles": input.Roles, "allowedEntities": input.AllowedEntities,
		"allowAllSites": input.AllowAllSites, "allowAllUsers": input.AllowAllUsers,
		"description": input.Description,
	}
	if input.AllowViewingUnobfuscatedData != nil {
		variables["allowViewingUnobfuscatedData"] = *input.AllowViewingUnobfuscatedData
	}
	return variables
}
