package static_host

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	domain "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources"
)

type Resource struct {
	deps *resources.Dependencies
}

var (
	_ resource.Resource              = (*Resource)(nil)
	_ resource.ResourceWithConfigure = (*Resource)(nil)
)

func NewResource() resource.Resource { return &Resource{} }

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_host"
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.deps = nil
	if req.ProviderData == nil {
		return
	}
	deps, ok := req.ProviderData.(*resources.Dependencies)
	if !ok || deps == nil || deps.CatoAPI == nil {
		resp.Diagnostics.AddError("Invalid resource configuration", "The provider did not supply the Cato API adapter.")
		return
	}
	r.deps = deps
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.deps == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before creating a static host.")
		return
	}
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for name, value := range map[string]types.String{
		"site_id": plan.SiteID, "name": plan.Name, "ip": plan.IP, "mac_address": plan.MacAddress,
	} {
		if value.IsUnknown() || (name != "mac_address" && value.IsNull()) {
			resp.Diagnostics.AddAttributeError(path.Root(name), "Invalid static-host configuration",
				name+" must be known and configured before creation.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	useCase, err := domain.NewCreateStaticHostUseCase(r.deps.CatoAPI)
	if err != nil {
		resp.Diagnostics.AddError("Unable to initialize static-host creation", "The creation use case could not be initialized.")
		return
	}
	id, err := useCase.Execute(ctx, domain.CreateStaticHostCommand{
		AccountID: r.deps.AccountID, SiteID: plan.SiteID.ValueString(),
		Host: entities.StaticHost{Name: plan.Name.ValueString(), IP: plan.IP.ValueString(), MacAddress: plan.MacAddress.ValueStringPointer()},
	})
	resp.Diagnostics.Append(Creation(err)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("Read is outside this POC",
		"Static-host creation is the only implemented operation. Use the production provider for full lifecycle management.")
}

func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is outside this POC", "Static-host creation is the only implemented operation.")
}

func (r *Resource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is outside this POC",
		"Static-host creation is the only implemented operation. This POC does not delete the remote host.")
}
