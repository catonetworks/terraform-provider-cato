package fqdncontainer

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*containerResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerResource)(nil)
	_ resource.ResourceWithImportState = (*containerResource)(nil)
)

type containerResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the FQDN container resource.
func NewResource() resource.Resource { return &containerResource{} }
func (r *containerResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *containerResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *containerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fqdn_container"
}
func (r *containerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}
func (r *containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ resource.ResourceWithValidateConfig = (*containerResource)(nil)

func (r *containerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx,
		&p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateModel(ctx,
		p,
		&resp.Diagnostics)
	if resp.Diagnostics.HasError() ||
		!r.configured(&resp.Diagnostics) {
		return
	}
	var item *ContainerRecord
	var err error
	if p.Mode.ValueString() == "INLINE" {
		in := buildInline(ctx,
			p,
			&resp.Diagnostics)
		item,
			err = r.getClient().CreateInline(ctx,
			r.client.AccountId,
			in)
	} else {
		in := buildURL(ctx,
			p,
			&resp.Diagnostics)
		item,
			err = r.getClient().CreateURL(ctx,
			r.client.AccountId,
			in)
	}
	if item != nil && item.ID != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(p, item.ID))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create container",
			safeError(p,
				err))
		return
	}
	if item == nil ||
		!validID(types.StringValue(item.ID),
			&resp.Diagnostics) {
		resp.Diagnostics.AddError("Missing container identity",
			"API did not return a container ID.")
		return
	}
	current := r.hydrate(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx,
		&p)...)
	if resp.Diagnostics.HasError() ||
		!validID(p.ID,
			&resp.Diagnostics) ||
		!r.configured(&resp.Diagnostics) {
		return
	}
	item,
		err := r.getClient().Find(ctx,
		r.client.AccountId,
		p.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read container",
			safeError(p,
				err))
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err = ensureIdentity(item,
		p.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid container response",
			err.Error())
		return
	}
	current := r.hydrate(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p,
		state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx,
		&p)...)
	resp.Diagnostics.Append(req.State.Get(ctx,
		&state)...)
	if resp.Diagnostics.HasError() ||
		!validID(state.ID,
			&resp.Diagnostics) {
		return
	}
	validateModel(ctx,
		p,
		&resp.Diagnostics)
	if resp.Diagnostics.HasError() ||
		!r.configured(&resp.Diagnostics) {
		return
	}
	old,
		err := r.getClient().Find(ctx,
		r.client.AccountId,
		state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read container before update",
			safeError(p,
				err))
		return
	}
	if err = ensureIdentity(old,
		state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Container no longer exists",
			err.Error())
		return
	}
	mode := "INLINE"
	if old.SyncData != nil {
		mode = "URL"
	}
	if old.Name != p.Name.ValueString() ||
		mode != p.Mode.ValueString() {
		resp.Diagnostics.AddError("Container replacement required",
			"Name and mode cannot be changed in place. Replace the container.")
		return
	}
	ref := &m.ContainerRefInput{By: m.ObjectRefByID, Input: state.ID.ValueString()}
	var item *ContainerRecord
	if mode == "INLINE" {
		in := buildInline(ctx,
			p,
			&resp.Diagnostics)
		item,
			err = r.getClient().UpdateInline(ctx,
			r.client.AccountId,
			m.UpdateFqdnContainerFromListInput{Ref: ref,
				Description: in.Description,
				Values:      in.Values})
	} else {
		in := buildURL(ctx,
			p,
			&resp.Diagnostics)
		s := in.SyncData
		n := s.Notifications
		item,
			err = r.getClient().UpdateURL(ctx,
			r.client.AccountId,
			m.UpdateFqdnContainerFromURLInput{Ref: ref,
				Description: &in.Description,
				FileType:    &in.FileType,
				SyncData: &m.UpdateContainerSyncDataInput{URL: &s.URL,
					TimeInterval: &s.TimeInterval,
					TimeUnit:     &s.TimeUnit,
					Notifications: &m.UpdateContainerSyncDataNotificationInput{MailingList: n.MailingList,
						SubscriptionGroup: n.SubscriptionGroup,
						Webhook:           n.Webhook}}})
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to update container",
			safeError(p,
				err))
		return
	}
	if err = ensureIdentity(item,
		state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid update response",
			err.Error())
		return
	}
	current := r.hydrate(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *containerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx,
		&p)...)
	if resp.Diagnostics.HasError() ||
		!validID(p.ID,
			&resp.Diagnostics) ||
		!r.configured(&resp.Diagnostics) {
		return
	}
	item,
		err := r.getClient().Find(ctx,
		r.client.AccountId,
		p.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to check container before deletion",
			safeError(p,
				err))
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err = ensureIdentity(item,
		p.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid container response",
			err.Error())
		return
	}
	item,
		err = r.getClient().Delete(ctx,
		r.client.AccountId,
		p.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete container",
			safeError(p,
				err))
		return
	}
	if err = ensureIdentity(item,
		p.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid delete response",
			err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *containerResource) hydrate(ctx context.Context, item *ContainerRecord, d *diag.Diagnostics) resourceModel {
	var download *DownloadRecord
	if item.SyncData == nil {
		var err error
		download,
			err = r.getClient().Download(ctx,
			r.client.AccountId,
			item.ID)
		if err != nil {
			d.AddError("Failed to read container members",
				err.Error())
			return resourceModel{}
		}
	}
	return parseContainer(ctx, item, download, d)
}
