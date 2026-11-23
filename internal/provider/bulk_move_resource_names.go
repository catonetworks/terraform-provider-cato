package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/antimalware/amfhindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol/acindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/clientconnectivity/ccindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip/diaindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/remoteportfwd/rpfindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sdpstaticip/ssiaindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy/swpindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/socketbypass/sbindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel/stindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/tenant/atindex"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/terminalserver/tsindex"
)

// Full names wrap the bulk ordering implementations and retain state migration from abbreviated names.
func bulkMoveResourceRegistrations() []func() resource.Resource {
	policies := []struct {
		name        string
		newResource func() resource.ResourceWithConfigure
	}{
		{"anti_malware_file_hash", amfhindex.NewAmfhRulesIndexResource},
		{"app_control", acindex.NewApcRulesIndexResource},
		{"app_tenant_restriction", atindex.NewAtrRulesIndexResource},
		{"client_connectivity", ccindex.NewCcRulesIndexResource},
		{"dynamic_ip_allocation", diaindex.NewDiaRulesIndexResource},
		{"remote_port_forward", rpfindex.NewRpfRulesIndexResource},
		{"sdp_static_ip_allocation", ssiaindex.NewSsiaRulesIndexResource},
		{"site_web_proxy", swpindex.NewSwpRulesIndexResource},
		{"socket_bypass", sbindex.NewSbRulesIndexResource},
		{"split_tunnel", stindex.NewStRulesIndexResource},
		{"terminal_server", tsindex.NewTsRulesIndexResource},
	}
	registrations := make([]func() resource.Resource, 0, len(policies))
	for _, policy := range policies {
		suffix := "_bulk_" + policy.name + "_move_rule"
		registrations = append(registrations,
			func() resource.Resource {
				return &bulkMoveResource{ResourceWithConfigure: policy.newResource(), suffix: suffix}
			},
		)
	}
	return registrations
}

type bulkMoveResource struct {
	resource.ResourceWithConfigure
	suffix string
}

var _ resource.ResourceWithMoveState = (*bulkMoveResource)(nil)

func (r *bulkMoveResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.suffix
}

func (r *bulkMoveResource) MoveState(ctx context.Context) []resource.StateMover {
	var source resource.MetadataResponse
	r.ResourceWithConfigure.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &source)
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)

	return []resource.StateMover{{
		SourceSchema: &schema.Schema,
		StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			// Version zero is the schema used by the abbreviated names in the beta.
			if req.SourceProviderAddress != "registry.terraform.io/catonetworks/cato" ||
				req.SourceTypeName != source.TypeName || req.SourceSchemaVersion != 0 || schema.Schema.Version != 0 {
				return
			}
			if req.SourceState == nil || req.SourceState.Raw.IsNull() {
				resp.Diagnostics.AddError("Invalid bulk ordering state", "Terraform could not decode the state of "+source.TypeName+".")
				return
			}
			resp.TargetState.Raw = req.SourceState.Raw
			resp.TargetPrivate = req.SourcePrivate
		},
	}}
}
