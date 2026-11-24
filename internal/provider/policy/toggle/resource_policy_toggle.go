// Package toggle adds account-level enable/disable resources for the four
// policies that had none: Internet Firewall, WAN Firewall, TLS Inspection and
// LAN (socket LAN) Firewall. Every other policy family already has a
// cato_<family>_policy toggle; without these four, a tenant-to-tenant
// replication could create every rule and still leave these policies switched
// off on the target (a new tenant ships with them disabled).
//
// The pattern matches cato_wan_network_policy: updatePolicy(state), publish the
// draft, read back the live `enabled` flag. Delete is a no-op (the policy
// always exists).
package toggle

import (
	"context"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

type family struct {
	typeName string // resource type suffix, e.g. "if_policy"
	id       string // fixed import id
	label    string
	read     func(ctx context.Context, c *cato.Client, acct string) (bool, error)
	update   func(ctx context.Context, c *cato.Client, acct string, st cato_models.PolicyToggleState) error
	publish  func(ctx context.Context, c *cato.Client, acct string) error
}

type model struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type toggleResource struct {
	f      family
	client *catoclient.Service
}

var (
	_ resource.Resource                = &toggleResource{}
	_ resource.ResourceWithConfigure   = &toggleResource{}
	_ resource.ResourceWithImportState = &toggleResource{}
)

func (r *toggleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.f.typeName
}

func (r *toggleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Account-level %s policy enable/disable toggle. Delete is a no-op.", r.f.label),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Fixed identifier for import",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Description: fmt.Sprintf("Whether the %s policy is enabled", r.f.label),
				Required:    true,
			},
		},
	}
}

func (r *toggleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *toggleResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), r.f.id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

func (r *toggleResource) apply(ctx context.Context, plan model) (model, error) {
	st := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		st = cato_models.PolicyToggleStateEnabled
	}
	c := r.client.V2()
	if err := r.f.update(ctx, c, r.client.AccountId, st); err != nil {
		return plan, err
	}
	if err := r.f.publish(ctx, c, r.client.AccountId); err != nil {
		return plan, err
	}
	return r.readModel(ctx)
}

func (r *toggleResource) readModel(ctx context.Context) (model, error) {
	enabled, err := r.f.read(ctx, r.client.V2(), r.client.AccountId)
	if err != nil {
		return model{}, err
	}
	return model{ID: types.StringValue(r.f.id), Enabled: types.BoolValue(enabled)}, nil
}

func (r *toggleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	st, err := r.apply(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(r.f.label+" policy toggle", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *toggleResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	st, err := r.readModel(ctx)
	if err != nil {
		resp.Diagnostics.AddError(r.f.label+" policy read", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *toggleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	st, err := r.apply(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(r.f.label+" policy toggle", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *toggleResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Warn(ctx, "cato_"+r.f.typeName+" delete is a no-op; the policy always exists")
}

// publishOK treats "no draft to publish" as success, like the other toggles.
func publishOK[T sharedpolicy.MutationError](status *cato_models.PolicyMutationStatus, errs []T) error {
	return sharedpolicy.PublishResult(status, errs)
}

// NewInternetFirewallPolicyResource manages cato_if_policy.
func NewInternetFirewallPolicyResource() resource.Resource {
	return &toggleResource{f: family{
		typeName: "if_policy", id: "internet_firewall", label: "Internet Firewall",
		read: func(ctx context.Context, c *cato.Client, acct string) (bool, error) {
			res, err := c.InternetFirewallPolicy(ctx, acct, nil)
			if err != nil {
				return false, err
			}
			return res.GetPolicy().GetInternetFirewall().GetPolicy().Enabled, nil
		},
		update: func(ctx context.Context, c *cato.Client, acct string, st cato_models.PolicyToggleState) error {
			res, err := c.PolicyInternetFirewallUpdatePolicy(ctx, nil, cato_models.InternetFirewallPolicyUpdateInput{State: &st}, acct)
			if err != nil {
				return err
			}
			up := res.GetPolicy().GetInternetFirewall().GetUpdatePolicy()
			return sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors())
		},
		publish: func(ctx context.Context, c *cato.Client, acct string) error {
			res, err := c.PolicyInternetFirewallPublishPolicyRevision(ctx, nil, nil, acct)
			if err != nil {
				return err
			}
			pub := res.GetPolicy().GetInternetFirewall().GetPublishPolicyRevision()
			return publishOK(pub.GetStatus(), pub.GetErrors())
		},
	}}
}

// NewWanFirewallPolicyResource manages cato_wf_policy.
func NewWanFirewallPolicyResource() resource.Resource {
	return &toggleResource{f: family{
		typeName: "wf_policy", id: "wan_firewall", label: "WAN Firewall",
		read: func(ctx context.Context, c *cato.Client, acct string) (bool, error) {
			res, err := c.WanFirewallPolicy(ctx, acct, nil)
			if err != nil {
				return false, err
			}
			return res.GetPolicy().GetWanFirewall().GetPolicy().Enabled, nil
		},
		update: func(ctx context.Context, c *cato.Client, acct string, st cato_models.PolicyToggleState) error {
			res, err := c.PolicyWanFirewallUpdatePolicy(ctx, cato_models.WanFirewallPolicyUpdateInput{State: &st}, acct, nil)
			if err != nil {
				return err
			}
			up := res.GetPolicy().GetWanFirewall().GetUpdatePolicy()
			return sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors())
		},
		publish: func(ctx context.Context, c *cato.Client, acct string) error {
			res, err := c.PolicyWanFirewallPublishPolicyRevision(ctx, nil, acct, nil)
			if err != nil {
				return err
			}
			pub := res.GetPolicy().GetWanFirewall().GetPublishPolicyRevision()
			return publishOK(pub.GetStatus(), pub.GetErrors())
		},
	}}
}

// NewTLSInspectPolicyResource manages cato_tls_inspect_policy.
func NewTLSInspectPolicyResource() resource.Resource {
	return &toggleResource{f: family{
		typeName: "tls_inspect_policy", id: "tls_inspect", label: "TLS Inspection",
		read: func(ctx context.Context, c *cato.Client, acct string) (bool, error) {
			res, err := c.Tlsinspectpolicy(ctx, acct, nil)
			if err != nil {
				return false, err
			}
			return res.GetPolicy().GetTLSInspect().GetPolicy().Enabled, nil
		},
		update: func(ctx context.Context, c *cato.Client, acct string, st cato_models.PolicyToggleState) error {
			res, err := c.PolicyTLSInspectUpdatePolicy(ctx, cato_models.TLSInspectPolicyUpdateInput{State: &st}, acct, nil)
			if err != nil {
				return err
			}
			up := res.GetPolicy().GetTLSInspect().GetUpdatePolicy()
			return sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors())
		},
		publish: func(ctx context.Context, c *cato.Client, acct string) error {
			res, err := c.PolicyTLSInspectPublishPolicyRevision(ctx, acct, nil, nil)
			if err != nil {
				return err
			}
			pub := res.GetPolicy().GetTLSInspect().GetPublishPolicyRevision()
			return publishOK(pub.GetStatus(), pub.GetErrors())
		},
	}}
}

// NewSocketLanPolicyResource manages cato_socket_lan_policy (LAN Firewall).
func NewSocketLanPolicyResource() resource.Resource {
	return &toggleResource{f: family{
		typeName: "socket_lan_policy", id: "socket_lan", label: "LAN Firewall",
		read: func(ctx context.Context, c *cato.Client, acct string) (bool, error) {
			res, err := c.PolicySocketLanPolicy(ctx, acct, nil)
			if err != nil {
				return false, err
			}
			return res.GetPolicy().GetSocketLan().GetPolicy().Enabled, nil
		},
		update: func(ctx context.Context, c *cato.Client, acct string, st cato_models.PolicyToggleState) error {
			res, err := c.PolicySocketLanUpdatePolicy(ctx, nil, cato_models.SocketLanPolicyUpdateInput{State: &st}, acct)
			if err != nil {
				return err
			}
			up := res.GetPolicy().GetSocketLan().GetUpdatePolicy()
			return sharedpolicy.MutationResult(up.GetStatus(), up.GetErrors())
		},
		publish: func(ctx context.Context, c *cato.Client, acct string) error {
			res, err := c.PolicySocketLanPublishPolicyRevision(ctx, nil, nil, acct)
			if err != nil {
				return err
			}
			pub := res.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
			return publishOK(pub.GetStatus(), pub.GetErrors())
		},
	}}
}
