package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
)

func TestPrivateAccessRuleDeleteAPIError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 1)
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"errors":[{"message":"delete request denied"}]}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "returned status: FAILURE", resp.Diagnostics[0].Detail())
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
}

func TestPrivateAccessRuleDeleteMutationErrors(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 1)
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound","errorMessage":"rule does not exist"},{"errorCode":"PolicyLocked","errorMessage":"policy is locked"}]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Len(t, resp.Diagnostics, 2)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[0].Summary())
	require.Equal(t, "ERROR: rule does not exist [RuleNotFound]", resp.Diagnostics[0].Detail())
	require.Equal(t, "Catov2 API PolicyPrivateAccessDeleteRule failed for 'test rule'", resp.Diagnostics[1].Summary())
	require.Equal(t, "ERROR: policy is locked [PolicyLocked]", resp.Diagnostics[1].Detail())
	require.Len(t, operations, 1)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
}

func TestPrivateAccessRuleDeletePublishAPIError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"errors":[{"message":"publish request denied"}]}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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

func TestPrivateAccessRuleDeletePublishMutationError(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyValidationError","errorMessage":"policy validation failed"}]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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

func TestPrivateAccessRuleDeletePublishesRevision(t *testing.T) {
	ctx := context.Background()
	operations := make(chan string, 2)
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
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "policyPrivateAccessDeleteRule":
			if body.Variables.Input.ID != "rule-123" {
				t.Errorf("unexpected rule ID: %q", body.Variables.Input.ID)
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`))
		case "policyPrivateAccessPublishRevision":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}}`))
		default:
			t.Errorf("unexpected GraphQL operation: %s", body.OperationName)
			http.Error(w, "unexpected operation", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.Timeout = 5 * time.Second
	client, err := cato.New(server.URL, "", "account-123", httpClient, nil)
	require.NoError(t, err)
	r := &privAccessRuleResource{client: &catoClientData{AccountId: "account-123", catov2: client}}
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

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

	require.Empty(t, resp.Diagnostics)
	require.Len(t, operations, 2)
	require.Equal(t, "policyPrivateAccessDeleteRule", <-operations)
	require.Equal(t, "policyPrivateAccessPublishRevision", <-operations)
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
