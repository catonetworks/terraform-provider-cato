package fqdncontainer

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	containerhelpers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/container"
)

func inlineValues(_ context.Context, v types.Set, d *diag.Diagnostics) []string {
	result := []string{}
	if v.IsNull() ||
		v.IsUnknown() {
		d.AddError("Invalid values",
			"INLINE requires a known set of members, including [] for an empty set.")
		return result
	}
	for _, item := range v.Elements() {
		s,
			ok := item.(types.String)
		if !ok ||
			!containerhelpers.KnownString(s) ||
			!validMember(s.ValueString()) {
			d.AddError("Invalid member",
				"Use known canonical lowercase FQDN values without spaces, wildcards or trailing slashes.")
			continue
		}
		result = append(result,
			s.ValueString())
	}
	return result
}
func buildInline(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.CreateFqdnContainerFromListInput {
	validateModel(ctx,
		p,
		d)
	return m.CreateFqdnContainerFromListInput{Name: p.Name.ValueString(),
		Description: p.Description.ValueString(),
		Values: inlineValues(ctx,
			p.Values,
			d)}
}
func buildURL(_ context.Context, p resourceModel, d *diag.Diagnostics) m.CreateFqdnContainerFromURLInput {
	return m.CreateFqdnContainerFromURLInput{Name: p.Name.ValueString(),
		Description: p.Description.ValueString(),
		FileType:    m.ContainerFileType(p.FileType.ValueString()),
		SyncData: containerhelpers.BuildSync(p.FileType,
			p.SyncData,
			d)}
}

// The API download is base64 UTF-8, one canonical member per line, without a header.
func decodeMembers(encoded string, d *diag.Diagnostics) types.Set {
	members := containerhelpers.DecodeMembers(encoded, validMember, d)
	if members == nil {
		return types.SetNull(types.StringType)
	}
	values := []attr.Value{}
	for _, item := range members {
		values = append(values, types.StringValue(item))
	}
	return types.SetValueMust(types.StringType, values)
}
func parseContainer(ctx context.Context, item *ContainerRecord, download *DownloadRecord, d *diag.Diagnostics) resourceModel {
	p := resourceModel{ID: types.StringValue(item.GetID()),
		Name:        types.StringValue(item.GetName()),
		Description: types.StringValue(""),
		Mode:        types.StringValue("INLINE"),
		Size:        types.Int64Value(item.GetSize()),
		FileType:    types.StringNull(),
		SyncData:    types.ObjectNull(syncAttrTypes),
		Values:      types.SetNull(types.StringType)}
	if item.GetDescription() != nil {
		p.Description = types.StringValue(*item.GetDescription())
	}
	if item.GetID() == "" ||
		item.GetName() == "" ||
		item.GetSize() < 0 {
		d.AddError("Invalid container response",
			"Missing identity or invalid size.")
		return p
	}
	sync := item.GetSyncData()
	if sync == nil {
		if download == nil ||
			download.ID != item.GetID() ||
			download.Name != item.GetName() {
			d.AddError("Invalid member download",
				"Download did not confirm the exact container identity.")
			return p
		}
		p.Values = decodeMembers(download.EncodedFile,
			d)
		return p
	}
	p.Mode = types.StringValue("URL")
	p.FileType, p.SyncData = containerhelpers.ParseSync(ctx, sync, d)
	return p
}
func identityState(p resourceModel,
	id string) resourceModel {
	p.ID = types.StringValue(id)
	p.Size = types.Int64Value(0)
	p.SyncData = containerhelpers.NormalizeObject(p.SyncData)
	return p
}

func ensureIdentity(item *ContainerRecord,
	id string) error {
	if item == nil ||
		item.GetID() != id {
		return fmt.Errorf("API did not confirm the exact container identity")
	}
	return nil
}

func safeError(p resourceModel,
	err error) string {
	if p.Mode.ValueString() == "URL" ||
		!p.SyncData.IsNull() {
		return "Container API request failed. URL details are omitted to protect source URL parameters."
	}
	return err.Error()
}
