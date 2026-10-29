package ipaddressrangecontainer

import (
	"context"
	"fmt"
	"sort"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	containerhelpers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/container"
)

func inlineValues(_ context.Context, v types.Set, d *diag.Diagnostics) []*m.IPAddressRangeInput {
	result := []*m.IPAddressRangeInput{}
	if v.IsNull() ||
		v.IsUnknown() {
		d.AddError("Invalid values",
			"INLINE requires a known set of ranges, including [] for an empty set.")
		return result
	}
	for _, item := range v.Elements() {
		text,
			ok := memberText(item)
		if !ok {
			d.AddError("Invalid range",
				"Use known canonical IPv4 from/to addresses with from <= to.")
			continue
		}
		from,
			to,
			_ := parseMember(text)
		result = append(result,
			&m.IPAddressRangeInput{From: from.String(),
				To: to.String()})
	}
	sort.Slice(result, func(i, j int) bool {
		a, _, _ := parseMember(result[i].From)
		b, _, _ := parseMember(result[j].From)
		return a.Compare(b) < 0
	})
	for i := 1; i < len(result); i++ {
		_,
			previous,
			_ := parseMember(result[i-1].To)
		current,
			_,
			_ := parseMember(result[i].From)
		if previous.Compare(current) >= 0 {
			d.AddError("Overlapping ranges",
				"Use disjoint ranges; the API merges overlaps and would change configured set elements.")
		}
	}
	return result
}
func buildInline(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.CreateIPAddressRangeContainerFromListInput {
	validateModel(ctx,
		p,
		d)
	return m.CreateIPAddressRangeContainerFromListInput{Name: p.Name.ValueString(),
		Description: p.Description.ValueString(),
		Values: inlineValues(ctx,
			p.Values,
			d)}
}
func buildURL(_ context.Context, p resourceModel, d *diag.Diagnostics) m.CreateIPAddressRangeContainerFromURLInput {
	return m.CreateIPAddressRangeContainerFromURLInput{Name: p.Name.ValueString(),
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
		return types.SetNull(memberType)
	}
	values := []attr.Value{}
	for _, item := range members {
		values = append(values, memberValue(item))
	}
	return types.SetValueMust(memberType, values)
}
func parseContainer(ctx context.Context, item *ContainerRecord, download *DownloadRecord, d *diag.Diagnostics) resourceModel {
	p := resourceModel{ID: types.StringValue(item.GetID()),
		Name:        types.StringValue(item.GetName()),
		Description: types.StringValue(""),
		Mode:        types.StringValue("INLINE"),
		Size:        types.Int64Value(item.GetSize()),
		FileType:    types.StringNull(),
		SyncData:    types.ObjectNull(syncAttrTypes),
		Values:      types.SetNull(memberType)}
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

func memberValue(s string) types.Object {
	from,
		to,
		ok := parseMember(s)
	if !ok {
		return types.ObjectNull(rangeAttrTypes)
	}
	return types.ObjectValueMust(rangeAttrTypes,
		map[string]attr.Value{"from": types.StringValue(from.String()),
			"to": types.StringValue(to.String())})
}
