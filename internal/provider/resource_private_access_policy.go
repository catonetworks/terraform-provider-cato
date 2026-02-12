package provider

import (
	"context"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privAccessPolicyResource{}
	_ resource.ResourceWithConfigure   = &privAccessPolicyResource{}
	_ resource.ResourceWithImportState = &privAccessPolicyResource{}
)

func NewPrivAccessPolicyResource() resource.Resource {
	return &privAccessPolicyResource{}
}

type privAccessPolicyResource struct {
	client      *catoClientData
	initialized bool
}

func (r *privAccessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_policy"
}

func (r *privAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_policy` resource contains the configuration parameters for private access policies in the Cato platform.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "fake ID",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Is the private access policy enabled?",
				Required:    true,
			},
			"audit": schema.SingleNestedAttribute{
				Description: "Audit log record",
				Optional:    true,
				Computed:    true,
			},
			"revision": schema.SingleNestedAttribute{
				Description: "Private access policy revision",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *privAccessPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
	tflog.Info(ctx, "XXX Configure", map[string]interface{}{"initialized": r.initialized})
	r.initialized = true
}

func (r *privAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "XXX Create", map[string]interface{}{"initialized": r.initialized})
	r.initialized = true

	// var plan PrivAccessPolicyModel
}

func (r *privAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "XXX Update", map[string]interface{}{"initialized": r.initialized})
	r.initialized = true

}

func (r *privAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivAccessPolicyModel
	tflog.Info(ctx, "XXX Read", map[string]interface{}{"initialized": r.initialized})
	r.initialized = true

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, hydrateErr := r.hydratePrivAccessPolicyState(ctx, state)
	if hydrateErr != nil {
		// Check if app-connector not found
		if hydrateErr.Error() == "app_connector not found" { // TODO: check the actual error
			tflog.Warn(ctx, "app_connector not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "XXX Delete", map[string]interface{}{"initialized": r.initialized})
	r.initialized = true

}

// hydratePrivAccessPolicyState fetches the current state of a privAccessPolicy from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessPolicyResource) hydratePrivAccessPolicyState(ctx context.Context, plan PrivAccessPolicyModel) (*PrivAccessPolicyModel, error) {
	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PolicyReadPrivateAccessPolicy", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}

	// Map API response to PrivAccessPolicyModel
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	state := &PrivAccessPolicyModel{
		Enabled: types.BoolValue(policy.Enabled),
	}

	return state, nil
}
