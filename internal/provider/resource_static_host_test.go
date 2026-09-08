package provider

import (
	"context"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestStaticHostReadRefreshesState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewStaticHostClient(t)
	macAddress := "00:00:00:00:00:51"
	mockClient.EXPECT().
		SiteStaticHost(
			mock.Anything,
			"account-123",
			mock.MatchedBy(func(input cato_models.SiteRefInput) bool {
				return input.By == cato_models.ObjectRefByID && input.Input == "site-123"
			}),
			"host-123",
		).
		Return(staticHostResponse("host-123", "updated-host", "192.168.220.21", &macAddress), nil).
		Once()

	state := newStaticHostState(ctx, t, StaticHost{
		ID:         types.StringValue("host-123"),
		SiteID:     types.StringValue("site-123"),
		Name:       types.StringValue("old-host"),
		IP:         types.StringValue("192.168.220.20"),
		MacAddress: types.StringNull(),
	})
	r := &staticHostResource{
		client:           &catoClientData{AccountId: "account-123"},
		staticHostClient: mockClient,
	}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	assertStaticHostState(ctx, t, resp.State, StaticHost{
		ID:         types.StringValue("host-123"),
		SiteID:     types.StringValue("site-123"),
		Name:       types.StringValue("updated-host"),
		IP:         types.StringValue("192.168.220.21"),
		MacAddress: types.StringValue(macAddress),
	})
}

func TestStaticHostReadSetsNullMacAddress(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewStaticHostClient(t)
	mockClient.EXPECT().
		SiteStaticHost(mock.Anything, "account-123", mock.Anything, "host-123").
		Return(staticHostResponse("host-123", "host", "192.168.220.20", nil), nil).
		Once()

	state := newStaticHostState(ctx, t, StaticHost{
		ID:         types.StringValue("host-123"),
		SiteID:     types.StringValue("site-123"),
		Name:       types.StringValue("host"),
		IP:         types.StringValue("192.168.220.20"),
		MacAddress: types.StringValue("00:00:00:00:00:50"),
	})
	r := &staticHostResource{
		client:           &catoClientData{AccountId: "account-123"},
		staticHostClient: mockClient,
	}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var got StaticHost
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	if !got.MacAddress.IsNull() {
		t.Fatalf("expected null mac_address, got %q", got.MacAddress.ValueString())
	}
}

func TestStaticHostReadRemovesMissingResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewStaticHostClient(t)
	mockClient.EXPECT().
		SiteStaticHost(mock.Anything, "account-123", mock.Anything, "host-123").
		Return(&cato.SiteStaticHost{Site: cato.SiteStaticHost_Site{}}, nil).
		Once()

	r := &staticHostResource{
		client:           &catoClientData{AccountId: "account-123"},
		staticHostClient: mockClient,
	}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}

	r.Read(ctx, resource.ReadRequest{State: newStaticHostState(ctx, t, StaticHost{
		ID:     types.StringValue("host-123"),
		SiteID: types.StringValue("site-123"),
	})}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state to be removed when static host is missing")
	}
}

func TestStaticHostReadReportsAPIError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewStaticHostClient(t)
	mockClient.EXPECT().
		SiteStaticHost(mock.Anything, "account-123", mock.Anything, "host-123").
		Return(nil, errors.New("query failed")).
		Once()

	r := &staticHostResource{
		client:           &catoClientData{AccountId: "account-123"},
		staticHostClient: mockClient,
	}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}

	r.Read(ctx, resource.ReadRequest{State: newStaticHostState(ctx, t, StaticHost{
		ID:     types.StringValue("host-123"),
		SiteID: types.StringValue("site-123"),
	})}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for static host query error")
	}
}

func TestStaticHostImportState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &staticHostResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}
	if diags := resp.State.Set(ctx, StaticHost{
		ID:         types.StringNull(),
		SiteID:     types.StringNull(),
		Name:       types.StringNull(),
		IP:         types.StringNull(),
		MacAddress: types.StringNull(),
	}); diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "site-123/host-123"}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	assertStaticHostState(ctx, t, resp.State, StaticHost{
		ID:         types.StringValue("host-123"),
		SiteID:     types.StringValue("site-123"),
		Name:       types.StringNull(),
		IP:         types.StringNull(),
		MacAddress: types.StringNull(),
	})
}

func TestStaticHostImportStateRejectsInvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &staticHostResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getStaticHostSchema(ctx, t)}}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "host-123"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for malformed composite import ID")
	}
}

func getStaticHostSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &staticHostResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func newStaticHostState(ctx context.Context, t *testing.T, model StaticHost) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: getStaticHostSchema(ctx, t)}
	if diags := state.Set(ctx, model); diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	return state
}

func assertStaticHostState(ctx context.Context, t *testing.T, state tfsdk.State, want StaticHost) {
	t.Helper()

	var got StaticHost
	if diags := state.Get(ctx, &got); diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	if got.ID != want.ID {
		t.Errorf("id: got %q, want %q", got.ID.ValueString(), want.ID.ValueString())
	}
	if got.SiteID != want.SiteID {
		t.Errorf("site_id: got %q, want %q", got.SiteID.ValueString(), want.SiteID.ValueString())
	}
	if got.Name != want.Name {
		t.Errorf("name: got %q, want %q", got.Name.ValueString(), want.Name.ValueString())
	}
	if got.IP != want.IP {
		t.Errorf("ip: got %q, want %q", got.IP.ValueString(), want.IP.ValueString())
	}
	if got.MacAddress != want.MacAddress {
		t.Errorf("mac_address: got %q, want %q", got.MacAddress.ValueString(), want.MacAddress.ValueString())
	}
}

func staticHostResponse(hostID, name, ip string, macAddress *string) *cato.SiteStaticHost {
	return &cato.SiteStaticHost{
		Site: cato.SiteStaticHost_Site{
			StaticHost: &cato.SiteStaticHost_Site_StaticHost{
				Host: cato.SiteStaticHost_Site_StaticHost_Host{
					HostID:     hostID,
					Name:       name,
					IP:         ip,
					MacAddress: macAddress,
				},
			},
		},
	}
}
