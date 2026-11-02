package customapplication

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type backend struct {
	object        map[string]any
	fail, missing string
	wrongID       bool
	calls         []string
	inputs        []map[string]any
}

func remote() map[string]any {
	return map[string]any{"id": "app-1", "name": "test", "description": "test", "category": []any{}, "criteria": []any{map[string]any{"protocol": "TCP", "port": []any{float64(443)}, "portRange": []any{}, "destination": nil}}}
}
func setup(t *testing.T, b *backend) *appResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		op, v := body.OperationName, body.Variables
		b.calls = append(b.calls, op)
		require.Equal(t, "account-123", v["accountId"])
		w.Header().Set("Content-Type", "application/json")
		write := func(value any) { require.NoError(t, json.NewEncoder(w).Encode(value)) }
		fail := func(message string) { write(map[string]any{"errors": []any{map[string]any{"message": message}}}) }
		if b.fail == op {
			fail("denied")
			return
		}
		payload := map[string]any{}
		switch op {
		case "customAppDataAddCustomApplication":
			in := v["addCustomApplicationInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			b.object = remote()
			applyInput(t, b, in)
			payload["addCustomApplication"] = map[string]any{"customApplication": b.object}
		case "customAppDataUpdateCustomApplication":
			in := v["updateCustomApplicationInput"].(map[string]any)
			require.Equal(t, "app-1", in["id"])
			b.inputs = append(b.inputs, in)
			applyInput(t, b, in)
			payload["updateCustomApplication"] = map[string]any{"customApplication": b.object}
		case "customAppDataCustomApplication":
			require.Equal(t, map[string]any{"by": "ID", "input": "app-1"}, v["customApplicationRefInput"])
			if b.object == nil {
				message := b.missing
				if message == "" {
					message = "Custom application not found by ID: app-1"
				}
				fail(message)
				return
			}
			payload["customApplication"] = b.object
		case "customAppDataDeleteCustomApplication":
			require.Equal(t, map[string]any{"customApplication": map[string]any{"by": "ID", "input": "app-1"}}, v["deleteCustomApplicationInput"])
			payload["deleteCustomApplication"] = map[string]any{"customApplication": b.object}
			b.object = nil
		default:
			t.Fatalf("unexpected operation %s", op)
		}
		if b.wrongID && b.object != nil {
			b.object["id"] = "foreign"
		}
		write(map[string]any{"data": map[string]any{"customAppData": payload}})
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &appResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func applyInput(t *testing.T, b *backend, in map[string]any) {
	t.Helper()
	for _, key := range []string{"name", "description", "category"} {
		if value, ok := in[key]; ok {
			if key == "category" {
				refs := []any{}
				for _, ref := range value.([]any) {
					refs = append(refs, map[string]any{"id": ref.(map[string]any)["input"], "name": "test category"})
				}
				value = refs
			}
			b.object[key] = value
		}
	}
	if value, ok := in["criteria"]; ok {
		raw, err := json.Marshal(value)
		require.NoError(t, err)
		var rows []map[string]any
		require.NoError(t, json.Unmarshal(raw, &rows))
		for _, row := range rows {
			ports := row["port"].([]any)
			for i, value := range ports {
				ports[i] = wirePort(t, value)
			}
			for _, value := range row["portRange"].([]any) {
				pair := value.(map[string]any)
				pair["from"] = wirePort(t, pair["from"])
				pair["to"] = wirePort(t, pair["to"])
			}
		}
		b.object["criteria"] = rows
	}
}
func wirePort(t *testing.T, v any) float64 {
	t.Helper()
	if value, ok := v.(float64); ok {
		return value
	}
	n, err := strconv.Atoi(v.(string))
	require.NoError(t, err)
	return float64(n)
}
func schemaFor(t *testing.T) schema.Schema {
	t.Helper()
	var r resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &r)
	return r.Schema
}
func planFor(t *testing.T, value resourceModel) tfsdk.Plan {
	t.Helper()
	p := tfsdk.Plan{Schema: schemaFor(t)}
	require.False(t, p.Set(context.Background(), value).HasError())
	return p
}
func stateFor(t *testing.T, value resourceModel) tfsdk.State {
	t.Helper()
	s := tfsdk.State{Schema: schemaFor(t)}
	require.False(t, s.Set(context.Background(), value).HasError())
	return s
}
func stringSet(items ...string) types.Set {
	values := []attr.Value{}
	for _, value := range items {
		values = append(values, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, values)
}
func criterion() types.Object {
	return types.ObjectValueMust(criteriaAttrTypes, map[string]attr.Value{"protocol": types.StringValue("TCP"), "port": stringSet("443"), "port_range": types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{}), "destination": types.ObjectNull(destinationAttrTypes)})
}
func model() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Description: types.StringValue("test"), Category: types.SetValueMust(types.ObjectType{AttrTypes: idname.RefAttrTypes}, []attr.Value{}), Criteria: types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{criterion()})}
}
func TestCRUDAndClearing(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var state resourceModel
	require.False(t, create.State.Get(ctx, &state).HasError())
	require.Equal(t, "app-1", state.ID.ValueString())
	state.Name = types.StringValue("updated")
	state.Description = types.StringNull()
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{State: create.State, Plan: planFor(t, state)}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, []any{}, b.inputs[1]["category"])
	require.Equal(t, "", b.inputs[1]["description"])
	read := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.Equal(t, update.State.Raw, read.State.Raw)
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}

//nolint:gocyclo // Table covers independent invalid inputs with one no-API assertion.
func TestValidationBeforeAPI(t *testing.T) {
	for _, kind := range []string{"null_name", "unknown_name", "unknown_description", "null_criteria", "unknown_criteria", "empty_criteria", "null_element", "unknown_element", "bad_protocol", "unknown_ports", "null_ports", "unknown_port", "bad_port", "bad_range", "unknown_destination", "unknown_category"} {
		t.Run(kind, func(t *testing.T) {
			p := model()
			c := criterion().Attributes()
			switch kind {
			case "null_name":
				p.Name = types.StringNull()
			case "unknown_name":
				p.Name = types.StringUnknown()
			case "unknown_description":
				p.Description = types.StringUnknown()
			case "null_criteria":
				p.Criteria = types.SetNull(types.ObjectType{AttrTypes: criteriaAttrTypes})
			case "unknown_criteria":
				p.Criteria = types.SetUnknown(types.ObjectType{AttrTypes: criteriaAttrTypes})
			case "empty_criteria":
				p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{})
			case "null_element":
				p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{types.ObjectNull(criteriaAttrTypes)})
			case "unknown_element":
				p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{types.ObjectUnknown(criteriaAttrTypes)})
			case "unknown_category":
				p.Category = types.SetUnknown(types.ObjectType{AttrTypes: idname.RefAttrTypes})
			default:
				switch kind {
				case "bad_protocol":
					c["protocol"] = types.StringValue("INVALID")
				case "unknown_ports":
					c["port"] = types.SetUnknown(types.StringType)
				case "null_ports":
					c["port"] = types.SetNull(types.StringType)
				case "unknown_port":
					c["port"] = types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()})
				case "bad_port":
					c["port"] = stringSet("65536")
				case "bad_range":
					c["port_range"] = types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{types.ObjectValueMust(rangeAttrTypes, map[string]attr.Value{"from": types.StringValue("500"), "to": types.StringValue("400")})})
				default:
					c["destination"] = types.ObjectUnknown(destinationAttrTypes)
				}
				p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{types.ObjectValueMust(criteriaAttrTypes, c)})
			}
			b := &backend{}
			r := setup(t, b)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestMissingAndFailures(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "foreign", "unrelated_not_found"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{object: remote()}
			switch kind {
			case "missing":
				b.object = nil
			case "denied":
				b.fail = "customAppDataCustomApplication"
			case "foreign":
				b.wrongID = true
			default:
				b.object = nil
				b.missing = "permission lookup not found"
			}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("app-1")
			s := stateFor(t, p)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
			del := resource.DeleteResponse{State: s}
			r.Delete(context.Background(), resource.DeleteRequest{State: s}, &del)
			require.Equal(t, kind != "missing", del.Diagnostics.HasError())
			require.NotContains(t, b.calls, "customAppDataDeleteCustomApplication")
		})
	}
}
func TestMutationFailures(t *testing.T) {
	for _, op := range []string{"customAppDataAddCustomApplication", "customAppDataUpdateCustomApplication", "customAppDataDeleteCustomApplication"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{object: remote(), fail: op}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("app-1")
			s := stateFor(t, p)
			switch op {
			case "customAppDataAddCustomApplication":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "customAppDataUpdateCustomApplication":
				resp := resource.UpdateResponse{State: s}
				r.Update(context.Background(), resource.UpdateRequest{State: s, Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			default:
				resp := resource.DeleteResponse{State: s}
				r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			}
		})
	}
}
func TestImport(t *testing.T) {
	ctx := context.Background()
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	NewResource().(*appResource).ImportState(ctx, resource.ImportStateRequest{ID: "app-1"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "app-1", id.ValueString())
}

func TestNestedCriteriaRoundTrip(t *testing.T) {
	p := model()
	fields := criterion().Attributes()
	fields["protocol"] = types.StringValue("UDP")
	fields["port"] = stringSet("53", "65535")
	fields["port_range"] = types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{parseRange("1000", "2000")})
	fields["destination"] = types.ObjectValueMust(destinationAttrTypes, map[string]attr.Value{
		"domain": stringSet("example.com"), "fqdn": stringSet("api.example.com"),
		"destination_ip": types.ObjectValueMust(ipAttrTypes, map[string]attr.Value{"ip": stringSet("192.0.2.1"), "subnet": stringSet("198.51.100.0/24"), "ip_range": types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{parseRange("203.0.113.1", "203.0.113.10")})}),
	})
	p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{criterion(), types.ObjectValueMust(criteriaAttrTypes, fields)})
	p.Category = types.SetValueMust(types.ObjectType{AttrTypes: idname.RefAttrTypes}, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("14"), "name": types.StringUnknown()})})
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.True(t, p.Criteria.Equal(state.Criteria))
	require.Len(t, state.Category.Elements(), 1)
	rows := b.object["criteria"].([]map[string]any)
	rows[0], rows[1] = rows[1], rows[0]
	read := resource.ReadResponse{State: resp.State}
	r.Read(ctx, resource.ReadRequest{State: resp.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.True(t, resp.State.Raw.Equal(read.State.Raw))
	state.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{criterion()})
	state.Category = types.SetNull(types.ObjectType{AttrTypes: idname.RefAttrTypes})
	update := resource.UpdateResponse{State: resp.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, state), State: resp.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, []any{}, b.inputs[1]["category"])
	require.Len(t, b.inputs[1]["criteria"], 1)
}
func TestPartialCreateRetainsIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewCustomApplicationClient(t)
	result := new(cato.CustomAppDataAddCustomApplication)
	require.NoError(t, json.Unmarshal([]byte(`{"customAppData":{"addCustomApplication":{"customApplication":{"id":"app-1"}}}}`), result))
	api.EXPECT().CustomAppDataAddCustomApplication(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("publication failed")).Once()
	r := &appResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "app-1", state.ID.ValueString())
}

func TestEmptyDestinationIPRejected(t *testing.T) {
	p := model()
	fields := criterion().Attributes()
	fields["destination"] = types.ObjectValueMust(destinationAttrTypes, map[string]attr.Value{"domain": stringSet("example.com"), "fqdn": stringSet(), "destination_ip": types.ObjectValueMust(ipAttrTypes, map[string]attr.Value{"ip": stringSet(), "subnet": stringSet(), "ip_range": types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{})})})
	p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{types.ObjectValueMust(criteriaAttrTypes, fields)})
	b := &backend{}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Empty(t, b.calls)
}
func TestPortZeroRoundTrip(t *testing.T) {
	p := model()
	fields := criterion().Attributes()
	fields["port"] = stringSet("0", "65535")
	fields["port_range"] = types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes}, []attr.Value{parseRange("0", "65535")})
	p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, []attr.Value{types.ObjectValueMust(criteriaAttrTypes, fields)})
	b := &backend{}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	var state resourceModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	require.True(t, p.Criteria.Equal(state.Criteria))
	read := resource.ReadResponse{State: resp.State}
	r.Read(context.Background(), resource.ReadRequest{State: resp.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.True(t, resp.State.Raw.Equal(read.State.Raw))
}
