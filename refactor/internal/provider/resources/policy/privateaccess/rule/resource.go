package rule

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	domain "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources"
)

type Resource struct {
	deps *resources.Dependencies
}

var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
)

func NewResource() resource.Resource { return &Resource{} }

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule"
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.deps = nil
	if req.ProviderData == nil {
		return
	}
	deps, ok := req.ProviderData.(*resources.Dependencies)
	if !ok || deps == nil || deps.CatoAPI == nil || deps.ResourceLock == nil {
		resp.Diagnostics.AddError("Invalid resource configuration", "The provider did not supply API operations and policy coordination.")
		return
	}
	r.deps = deps
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.TrimSpace(req.ID) == "" {
		resp.Diagnostics.AddError("Missing rule ID", "Import an existing private-access rule by its ID.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.deps == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before reading the rule.")
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	useCase, err := domain.NewReadPrivateAccessPolicyRuleUseCase(r.deps.CatoAPI)
	if err != nil {
		resp.Diagnostics.AddError("Unable to initialize rule read", "The policy reader could not be initialized.")
		return
	}
	exists, err := useCase.Execute(ctx, domain.ReadPrivateAccessPolicyRuleCommand{
		AccountID: r.deps.AccountID, RuleID: id.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read private-access rule", "The published policy could not be read; the existing state is retained.")
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
	}
}

func (r *Resource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Import an existing rule",
		"This deletion POC supports import, refresh, and destroy. Import a rule before using it.")
}

func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is outside this POC",
		"This provider implements the rule deletion use case. Rule updates are not supported.")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.deps == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before deleting the rule.")
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing private-access rule ID",
			"Deletion requires a known, nonempty rule ID in state.")
		return
	}
	command := domain.DeletePrivateAccessPolicyRuleCommand{
		AccountID: r.deps.AccountID, RuleID: id.ValueString(),
	}
	useCase, err := domain.NewDeletePrivateAccessPolicyRuleUseCase(
		r.deps.CatoAPI, r.deps.ResourceLock, domain.DefaultRetryPolicy(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to initialize rule deletion", "The deletion use case could not be initialized.")
		return
	}
	resp.Diagnostics.Append(Deletion(useCase.Execute(ctx, command))...)
}
