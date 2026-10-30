package subscriptiongroup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type backend struct {
	object    map[string]any
	fail      string
	malformed bool
	calls     []string
	inputs    []map[string]any
}

func setup(t *testing.T, b *backend) *subscriptionGroupResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Query         string         `json:"query"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, "account-123", body.Variables["accountId"])
		require.NotContains(t, body.Query, "audit")
		op := strings.TrimPrefix(body.OperationName, "notification")
		op = strings.ToLower(op[:1]) + op[1:]
		b.calls = append(b.calls, op)
		w.Header().Set("Content-Type", "application/json")
		if b.fail == op {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		var payload any
		switch op {
		case "createSubscriptionGroup", "updateSubscriptionGroup":
			in := body.Variables[op+"Input"].(map[string]any)
			b.inputs = append(b.inputs, in)
			if op == "createSubscriptionGroup" {
				b.object = map[string]any{"id": "1001", "mailingList": []any{}, "integration": map[string]any{"webhookList": []any{}}}
			} else {
				require.Equal(t, updateDocument, body.Query)
				require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["subscriptionGroup"])
				require.Contains(t, in, "integration")
				require.Contains(t, in, "mailingList")
			}
			applyMembers(b, op, in)
			payload = map[string]any{"subscriptionGroup": b.object}
		case "subscriptionGroup", "deleteSubscriptionGroup":
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, body.Variables["subscriptionGroupRefInput"])
			payload = b.object
			if op == "deleteSubscriptionGroup" {
				payload = map[string]any{"subscriptionGroup": b.object}
				b.object = nil
			}
		default:
			t.Fatalf("unexpected operation: %s", op)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"notification": map[string]any{op: payload}}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &subscriptionGroupResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func schemaFor(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema
}
func planFor(t *testing.T, m resourceModel) tfsdk.Plan {
	t.Helper()
	p := tfsdk.Plan{Schema: schemaFor(t)}
	require.False(t, p.Set(context.Background(), m).HasError())
	return p
}
func stateFor(t *testing.T, m resourceModel) tfsdk.State {
	t.Helper()
	s := tfsdk.State{Schema: schemaFor(t)}
	require.False(t, s.Set(context.Background(), m).HasError())
	return s
}
func model() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Integration: types.SetValueMust(idname.RefObjectType, nil), MailingList: types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("42"), "name": types.StringUnknown()})})}
}
func TestCRUD(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var current resourceModel
	require.False(t, create.State.Get(ctx, &current).HasError())
	require.Equal(t, "1001", current.ID.ValueString())
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	p.ID = current.ID
	p.Name = types.StringValue("updated")
	p.Integration = types.SetNull(idname.RefObjectType)
	p.MailingList = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("42"), "name": types.StringUnknown()})})
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: read.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, []any{}, b.inputs[1]["integration"])
	require.Equal(t, []any{map[string]any{"by": "ID", "input": "42"}}, b.inputs[1]["mailingList"])
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Empty(t, current.Integration.Elements())
	require.False(t, current.Integration.IsNull())
	require.False(t, current.MailingList.IsNull())
	del := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestReadMissingAndFailures(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "different"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			if kind == "denied" {
				b.fail = "subscriptionGroup"
			}
			if kind == "malformed" {
				b.malformed = true
			}
			if kind == "different" {
				b.object = map[string]any{"id": "different", "name": "test"}
			}
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestMutationFailures(t *testing.T) {
	for _, op := range []string{"createSubscriptionGroup", "updateSubscriptionGroup", "deleteSubscriptionGroup"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op, object: map[string]any{"id": "1001", "name": "test"}}
			r := setup(t, b)
			m := model()
			m.ID = types.StringValue("1001")
			s := stateFor(t, m)
			ctx := context.Background()
			switch op {
			case "createSubscriptionGroup":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "updateSubscriptionGroup":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, m), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "deleteSubscriptionGroup":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}
func TestValidationAndImport(t *testing.T) {
	ctx := context.Background()
	for _, field := range []string{"name", "integration", "mailing_list"} {
		t.Run(field, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			m := model()
			switch field {
			case "name":
				m.Name = types.StringUnknown()
			case "integration":
				m.Integration = types.SetUnknown(idname.RefObjectType)
			case "mailing_list":
				m.MailingList = types.SetUnknown(idname.RefObjectType)
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
	resp := resource.ImportStateResponse{State: stateFor(t, resourceModel{ID: types.StringNull(), Name: types.StringNull(), Integration: types.SetNull(idname.RefObjectType), MailingList: types.SetNull(idname.RefObjectType)})}
	NewResource().(*subscriptionGroupResource).ImportState(ctx, resource.ImportStateRequest{ID: "1001"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
}

func TestClearUsesFreshMembership(t *testing.T) {
	ctx := context.Background()
	b := &backend{object: map[string]any{"id": "1001", "name": "test", "mailingList": []any{map[string]any{"id": "42", "name": "Fixture"}}, "integration": map[string]any{"webhookList": []any{}}}}
	r := setup(t, b)
	prior := model()
	prior.ID = types.StringValue("1001")
	prior.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	desired := prior
	response := resource.UpdateResponse{State: stateFor(t, prior)}
	r.Update(ctx, resource.UpdateRequest{State: stateFor(t, prior), Plan: planFor(t, desired)}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	require.Equal(t, []string{"subscriptionGroup", "updateSubscriptionGroup"}, b.calls)
	require.Equal(t, []any{map[string]any{"by": "ID", "input": "42"}}, b.inputs[0]["mailingListToRemove"])
	var actual resourceModel
	require.False(t, response.State.Get(ctx, &actual).HasError())
	require.Empty(t, actual.MailingList.Elements())
}

func TestPartialCreateRetainsIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewSubscriptionGroupClient(t)
	result := new(cato.NotificationCreateSubscriptionGroup)
	require.NoError(t, json.Unmarshal([]byte(`{"notification":{"createSubscriptionGroup":{"subscriptionGroup":{"id":"1001"}}}}`), result))
	api.EXPECT().NotificationCreateSubscriptionGroup(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &subscriptionGroupResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	m := model()
	m.MailingList = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("42"), "name": types.StringUnknown()})})
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "1001", state.ID.ValueString())
	for _, member := range state.MailingList.Elements() {
		for _, value := range member.(types.Object).Attributes() {
			require.False(t, value.IsUnknown())
		}
	}
}
func TestMissingListClassification(t *testing.T) {
	missing := func(message string) error {
		list := gqlerror.List{&gqlerror.Error{Message: message}}
		return &clientv2.ErrorResponse{GqlErrors: &list}
	}
	exact := "GQL-User-Error: Account ID [account-123] subscription [1001] not found"
	require.True(t, missingList(missing(exact), "account-123", "1001"))
	require.False(t, missingList(missing(exact), "another", "1001"))
	require.False(t, missingList(missing("permission denied"), "account-123", "1001"))
	require.False(t, missingList(errors.New(exact), "account-123", "1001"))
}
func TestMetadataConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*subscriptionGroupResource)
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_subscription_group", meta.TypeName)
	var resp resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	resp = resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123"}}, &resp)
	require.False(t, resp.Diagnostics.HasError())
}

func TestClearIntegrationUsesFreshMembership(t *testing.T) {
	ctx := context.Background()
	b := &backend{object: map[string]any{"id": "1001", "name": "test", "mailingList": []any{}, "integration": map[string]any{"webhookList": []any{map[string]any{"id": "77", "name": "Webhook"}}}}}
	r := setup(t, b)
	prior := model()
	prior.ID = types.StringValue("1001")
	prior.MailingList = types.SetValueMust(idname.RefObjectType, nil)
	response := resource.UpdateResponse{State: stateFor(t, prior)}
	r.Update(ctx, resource.UpdateRequest{State: stateFor(t, prior), Plan: planFor(t, prior)}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	require.Equal(t, []any{map[string]any{"by": "ID", "input": "77"}}, b.inputs[0]["integrationToRemove"])
	var actual resourceModel
	require.False(t, response.State.Get(ctx, &actual).HasError())
	require.Empty(t, actual.Integration.Elements())
}
func TestEmptyCreateAndDrift(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	p.MailingList = types.SetNull(idname.RefObjectType)
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.True(t, create.Diagnostics.HasError())
	require.Empty(t, b.calls)
	p.ID = types.StringValue("1001")
	b.object = map[string]any{"id": "1001", "name": "external change", "mailingList": []any{}, "integration": map[string]any{"webhookList": []any{}}}
	read := resource.ReadResponse{State: stateFor(t, p)}
	r.Read(ctx, resource.ReadRequest{State: stateFor(t, p)}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	var actual resourceModel
	require.False(t, read.State.Get(ctx, &actual).HasError())
	require.Equal(t, "external change", actual.Name.ValueString())
}

func applyMembers(b *backend, op string, in map[string]any) {
	for _, key := range []string{"name", "integration", "mailingList"} {
		value, ok := in[key]
		if !ok {
			continue
		}
		if key == "name" {
			b.object[key] = value
			continue
		}
		if op == "updateSubscriptionGroup" && len(value.([]any)) == 0 {
			removed, _ := in[key+"ToRemove"].([]any)
			if len(removed) == 0 {
				continue
			}
		}
		refs := []any{}
		for _, v := range value.([]any) {
			ref := v.(map[string]any)
			id, name := "42", "Fixture"
			if ref["by"] == "ID" {
				id = ref["input"].(string)
			} else {
				name = ref["input"].(string)
			}
			refs = append(refs, map[string]any{"id": id, "name": name})
		}
		if key == "integration" {
			b.object[key] = map[string]any{"webhookList": refs}
		} else {
			b.object[key] = refs
		}
	}
}
