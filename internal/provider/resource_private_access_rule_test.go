package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
)

func TestPrivateAccessRuleDelete(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Empty(t, resp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
}

func TestPrivateAccessRuleDeleteGraphQLError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 1)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"errors":[{"message":"delete request denied"}]}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Contains(t, resp.Diagnostics[0].Detail(), "delete request denied")
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
}

func TestPrivateAccessRuleDeleteFailureWithoutErrorDetails(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 1)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "returned status: FAILURE", resp.Diagnostics[0].Detail())
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
}

func TestPrivateAccessRuleDeleteRuleNotFoundWithOtherError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 1)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"},{"errorCode":"PolicyLocked","errorMessage":"policy is locked"}]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "ERROR: policy is locked [PolicyLocked]", resp.Diagnostics[0].Detail())
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
}

func TestPrivateAccessRuleDeletePublishGraphQLError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"errors":[{"message":"publish request denied"}]}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Contains(t, resp.Diagnostics[0].Detail(), "publish request denied")
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
}

func TestPrivateAccessRuleDeletePublishFailureWithoutErrorDetails(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "returned status: FAILURE", resp.Diagnostics[0].Detail())
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
}

func TestPrivateAccessRuleDeletePublishRevisionNotFoundWithOtherError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyRevisionNotFound","errorMessage":"no draft revision exists"},{"errorCode":"PolicyValidationError","errorMessage":"policy validation failed"}]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "ERROR: policy validation failed [PolicyValidationError]", resp.Diagnostics[0].Detail())
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
}

func TestPrivateAccessRuleDeleteRetriesAfterDeleteGraphQLError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"errors":[{"message":"transient delete failure"}]}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			publishCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// The failed removal must retain state without attempting to publish.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "transient delete failure")
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.EqualValues(t, 1, deleteCalls.Load())
	require.Zero(t, publishCalls.Load())

	// Terraform retries Delete with the retained state after the failure clears.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 1, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterDeleteServerError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("delete service unavailable"))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			publishCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// The failed removal must retain state without attempting to publish.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "503")
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "delete service unavailable")
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.EqualValues(t, 1, deleteCalls.Load())
	require.Zero(t, publishCalls.Load())

	// Terraform retries Delete with the retained state after the failure clears.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 1, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterDeleteFailureWithoutErrorDetails(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[]}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			publishCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// The failed removal must retain state without attempting to publish.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Equal(t, "returned status: FAILURE", firstResp.Diagnostics[0].Detail())
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.EqualValues(t, 1, deleteCalls.Load())
	require.Zero(t, publishCalls.Load())

	// Terraform retries Delete with the retained state after the failure clears.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 1, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterDeleteMutationError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"PolicyLocked","errorMessage":"policy is locked"}]}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			publishCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// The failed removal must retain state without attempting to publish.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Equal(t, "ERROR: policy is locked [PolicyLocked]", firstResp.Diagnostics[0].Detail())
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.EqualValues(t, 1, deleteCalls.Load())
	require.Zero(t, publishCalls.Load())

	// Terraform retries Delete with the retained state after the failure clears.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 1, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterPublishGraphQLError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"}]}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			if publishCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"errors":[{"message":"transient publish failure"}]}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
			}
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// Terraform retains state after the first publish fails.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "transient publish failure")
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)

	// A second Delete must publish even though the rule is already absent from the draft.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 2, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterPublishServerError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"}]}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			if publishCalls.Add(1) == 1 {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("publish service unavailable"))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
			}
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// Terraform retains state after the first publish fails.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "503")
	require.Contains(t, firstResp.Diagnostics[0].Detail(), "publish service unavailable")
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)

	// A second Delete must publish even though the rule is already absent from the draft.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 2, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterPublishFailureWithoutErrorDetails(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"}]}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			if publishCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[]}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
			}
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// Terraform retains state after the first publish fails.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Equal(t, "returned status: FAILURE", firstResp.Diagnostics[0].Detail())
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)

	// A second Delete must publish even though the rule is already absent from the draft.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 2, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteRetriesAfterPublishMutationError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	var deleteCalls, publishCalls atomic.Int32
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			if deleteCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"}]}}}}}`))
			}
		case "policyPrivateAccessPublishRevision":
			if publishCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyLocked","errorMessage":"policy is locked"}]}}}}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
			}
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	// Terraform retains state after the first publish fails.
	firstResp := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &firstResp)

	require.Len(t, firstResp.Diagnostics, 1)
	require.True(t, firstResp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for 'test rule'", firstResp.Diagnostics[0].Summary())
	require.Equal(t, "ERROR: policy is locked [PolicyLocked]", firstResp.Diagnostics[0].Detail())
	require.Equal(t, state, firstResp.State)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)

	// A second Delete must publish even though the rule is already absent from the draft.
	retryResp := resource.DeleteResponse{State: firstResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: firstResp.State}, &retryResp)

	require.Empty(t, retryResp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
	require.EqualValues(t, 2, deleteCalls.Load())
	require.EqualValues(t, 2, publishCalls.Load())
}

func TestPrivateAccessRuleDeleteAlreadyPublished(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
	r := newPrivateAccessRuleTestResource(t, operations, func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessDeleteRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"}]}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyRevisionNotFound","errorMessage":"no draft revision exists"}]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", operation)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	})

	state := newPrivateAccessRuleTestState(ctx, t, r)

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Empty(t, resp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
}

func newPrivateAccessRuleTestResource(
	t *testing.T,
	operations chan<- string,
	respond func(http.ResponseWriter, string),
) *privAccessRuleResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			OperationName string `json:"operationName"`
			Variables     struct {
				AccountID string `json:"accountID"`
				Input     struct {
					ID string `json:"id"`
				} `json:"input"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode GraphQL request: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if body.Variables.AccountID != "account-123" {
			t.Errorf("unexpected account ID: %q", body.Variables.AccountID)
		}
		if body.OperationName == "policyPrivateAccessDeleteRule" && body.Variables.Input.ID != "rule-123" {
			t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
		}
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		respond(w, body.OperationName)
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
	return r
}

func newPrivateAccessRuleTestState(ctx context.Context, t *testing.T, r *privAccessRuleResource) tfsdk.State {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	refType := types.ObjectType{AttrTypes: parse.IDNameRefModelTypes}
	diags := state.Set(ctx, PrivateAccessRuleModel{
		ID:                types.StringValue("rule-123"),
		Name:              types.StringValue("test rule"),
		Action:            types.StringValue("ALLOW"),
		Enabled:           types.BoolValue(true),
		Description:       types.StringNull(),
		ActivePeriod:      types.ObjectNull(PolicyRuleActivePeriodTypes),
		Applications:      types.SetNull(refType),
		ConnectionOrigins: types.SetNull(types.StringType),
		Countries:         types.SetNull(refType),
		Devices:           types.SetNull(refType),
		Platforms:         types.SetNull(types.StringType),
		Schedule:          types.ObjectNull(PolicyScheduleTypes),
		Source:            types.ObjectNull(SourceTypes),
		Tracking:          types.ObjectNull(PolicyRuleTrackingTypes),
		UserAttributes:    types.ObjectNull(UserAttributesTypes),
	})
	require.False(t, diags.HasError(), "seed state diagnostics: %v", diags)
	return state
}

func TestMove(t *testing.T) {
	type tc struct {
		id          string
		newPos      int
		expectedIDs []string
	}
	tcs := []tc{
		{id: "r1", newPos: 0, expectedIDs: []string{"r1", "r2", "r3", "r4"}},
		{id: "r3", newPos: 1, expectedIDs: []string{"r1", "r3", "r2", "r4"}},
		{id: "r4", newPos: 3, expectedIDs: []string{"r1", "r2", "r3", "r4"}},
		{id: "r4", newPos: 1, expectedIDs: []string{"r1", "r4", "r2", "r3"}},
		{id: "r4", newPos: 0, expectedIDs: []string{"r4", "r1", "r2", "r3"}},
		{id: "r1", newPos: 1, expectedIDs: []string{"r1", "r1", "r3", "r4"}}, // moving down is not supported!
	}

	ctx := context.Background()
	r := privAccessRuleBulkResource{}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s-%d", tc.id, tc.newPos), func(t *testing.T) {
			currentRules := getTestRules()
			if err := r.moveToPosition(ctx, currentRules, tc.id, "some name", tc.newPos); err != nil {
				t.Fatalf("moveToPosition returned error: %v", err)
			}
			if len(currentRules) != len(tc.expectedIDs) {
				t.Errorf("Expected %d rules, got %d", len(tc.expectedIDs), len(currentRules))
				return
			}
			for j, rule := range currentRules {
				if rule.ID.ValueString() != tc.expectedIDs[j] {
					t.Errorf("Expected rule %d to be '%s', got '%s'", j, tc.expectedIDs[j], rule.ID.String())
				}
			}
		})
	}
}

func getTestRules() []*PrivateAccessBulkRule {
	return []*PrivateAccessBulkRule{
		{ID: types.StringValue("r1"), Name: types.StringValue("Rule1")},
		{ID: types.StringValue("r2"), Name: types.StringValue("Rule2")},
		{ID: types.StringValue("r3"), Name: types.StringValue("Rule3")},
		{ID: types.StringValue("r4"), Name: types.StringValue("Rule4")},
	}
}
