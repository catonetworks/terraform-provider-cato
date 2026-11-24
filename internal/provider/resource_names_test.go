package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestBulkMoveResourceNames(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	server := providerserver.NewProtocol6(New("test")())()
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemas.Diagnostics)

	renames := map[string]string{
		"amfh": "anti_malware_file_hash",
		"apc":  "app_control",
		"atr":  "app_tenant_restriction",
		"cc":   "client_connectivity",
		"dia":  "dynamic_ip_allocation",
		"rpf":  "remote_port_forward",
		"sb":   "socket_bypass",
		"ssia": "sdp_static_ip_allocation",
		"st":   "split_tunnel",
		"swp":  "site_web_proxy",
		"ts":   "terminal_server",
	}

	for shortName, fullName := range renames {
		t.Run(fullName, func(t *testing.T) {
			t.Parallel()

			oldType := "cato_bulk_" + shortName + "_move_rule"
			newType := "cato_bulk_" + fullName + "_move_rule"
			newSchema := schemas.ResourceSchemas[newType]
			require.NotContains(t, schemas.ResourceSchemas, oldType)
			require.NotNil(t, newSchema)
			require.False(t, newSchema.Block.Deprecated)

			state := &tfprotov6.RawState{JSON: []byte(`{
				"section_data": {"Office": {"section_index": 1}},
				"rule_data": {"Allow": {"index_in_section": 2, "section_name": "Office"}}
			}`)}
			req := &tfprotov6.MoveResourceStateRequest{
				SourceProviderAddress: "registry.terraform.io/catonetworks/cato",
				SourceTypeName:        oldType,
				SourceSchemaVersion:   0, // State written by v2.0.0-prerelease.1.
				SourceState:           state,
				TargetTypeName:        newType,
				SourcePrivate:         []byte(`{"test":"dHJ1ZQ=="}`), // Framework private values contain base64-encoded JSON (true).
			}
			result, err := server.MoveResourceState(t.Context(), req)
			require.NoError(t, err)
			require.Empty(t, result.Diagnostics)
			require.NotNil(t, result.TargetState)
			moved, err := result.TargetState.Unmarshal(newSchema.ValueType())
			require.NoError(t, err)
			original, err := state.Unmarshal(newSchema.ValueType())
			require.NoError(t, err)
			require.True(t, original.Equal(moved))
			require.Equal(t, req.SourcePrivate, result.TargetPrivate)
		})
	}
}

func TestBulkMoveResourceRejectsUnrelatedState(t *testing.T) {
	t.Parallel()

	valid := tfprotov6.MoveResourceStateRequest{
		SourceProviderAddress: "registry.terraform.io/catonetworks/cato",
		SourceTypeName:        "cato_bulk_cc_move_rule",
		SourceState:           &tfprotov6.RawState{JSON: []byte(`{"section_data": null, "rule_data": {}}`)},
		TargetTypeName:        "cato_bulk_client_connectivity_move_rule",
	}
	tests := map[string]func(*tfprotov6.MoveResourceStateRequest){
		"other provider": func(req *tfprotov6.MoveResourceStateRequest) {
			req.SourceProviderAddress = "registry.terraform.io/other/cato"
		},
		"other policy": func(req *tfprotov6.MoveResourceStateRequest) {
			req.SourceTypeName = "cato_bulk_apc_move_rule"
		},
		"other schema version": func(req *tfprotov6.MoveResourceStateRequest) {
			req.SourceSchemaVersion = 1
		},
		"invalid state": func(req *tfprotov6.MoveResourceStateRequest) {
			req.SourceState = &tfprotov6.RawState{JSON: []byte(`{"rule_data": "invalid"}`)}
		},
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := valid
			change(&req)
			server := providerserver.NewProtocol6(New("test")())()
			result, err := server.MoveResourceState(t.Context(), &req)
			require.NoError(t, err)
			require.NotEmpty(t, result.Diagnostics)
			require.Equal(t, tfprotov6.DiagnosticSeverityError, result.Diagnostics[0].Severity)
			require.Nil(t, result.TargetState)
		})
	}
}

func TestDeprecatedPolicyResourcesAreRemoved(t *testing.T) {
	t.Parallel()

	server := providerserver.NewProtocol6(New("test")())()
	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemas.Diagnostics)

	for oldType, newType := range map[string]string{
		"cato_pa_section":              "cato_private_access_section",
		"cato_remote_port_fwd_rule":    "cato_remote_port_forward_rule",
		"cato_remote_port_fwd_section": "cato_remote_port_forward_section",
	} {
		newSchema := schemas.ResourceSchemas[newType]
		require.NotContains(t, schemas.ResourceSchemas, oldType)
		require.NotNil(t, newSchema, newType)
		require.False(t, newSchema.Block.Deprecated, newType)
	}
}
