package fqdncontainer

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
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
			!knownString(s) ||
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
	in := m.CreateFqdnContainerFromURLInput{Name: p.Name.ValueString(),
		Description: p.Description.ValueString(),
		FileType:    m.ContainerFileType(p.FileType.ValueString())}
	if !knownString(p.FileType) ||
		!in.FileType.IsValid() ||
		p.SyncData.IsNull() ||
		p.SyncData.IsUnknown() {
		d.AddError("Invalid sync configuration",
			"URL mode requires known file_type and sync_data.")
		return in
	}
	a := p.SyncData.Attributes()
	n,
		ok := a["notifications"].(types.Object)
	interval,
		okI := a["time_interval"].(types.Int64)
	unit,
		okU := a["time_unit"].(types.String)
	u,
		okURL := a["url"].(types.String)
	if !ok || !okI || !okU || !okURL || !validSyncScalars(n, interval, unit, u) {
		d.AddError("Invalid sync settings",
			"Use a positive Int32 interval, HOUR/DAY, notifications and an HTTPS URL without user credentials or fragments.")
		return in
	}

	refs := n.Attributes()
	ml, sg, wh := refSet(refs["mailing_list"], d), refSet(refs["subscription_group"], d), refSet(refs["webhook"], d)
	validateDestinations([]int{len(ml), len(sg), len(wh)}, d)

	notify := &m.CreateContainerSyncDataNotificationInput{MailingList: []*m.SubscriptionMailingListRefInput{},
		SubscriptionGroup: []*m.SubscriptionGroupRefInput{},
		Webhook:           []*m.SubscriptionWebhookRefInput{}}
	for _, ref := range ml {
		notify.MailingList = append(notify.MailingList,
			&m.SubscriptionMailingListRefInput{By: ref.By,
				Input: ref.Input})
	}
	for _, ref := range sg {
		notify.SubscriptionGroup = append(notify.SubscriptionGroup,
			&m.SubscriptionGroupRefInput{By: ref.By,
				Input: ref.Input})
	}
	for _, ref := range wh {
		notify.Webhook = append(notify.Webhook,
			&m.SubscriptionWebhookRefInput{By: ref.By,
				Input: ref.Input})
	}
	in.SyncData = &m.CreateContainerSyncDataInput{Notifications: notify,
		TimeInterval: interval.ValueInt64(),
		TimeUnit:     m.ContainerSyncDataTimeUnit(unit.ValueString()),
		URL:          u.ValueString()}
	return in
}
func refSet(value attr.Value, d *diag.Diagnostics) []*m.ContainerRefInput {
	result := []*m.ContainerRefInput{}
	set,
		ok := value.(types.Set)
	if !ok ||
		set.IsNull() ||
		set.IsUnknown() {
		d.AddError("Invalid notification references",
			"All three notification sets must be known, including [] for unused types.")
		return result
	}
	for _, item := range set.Elements() {
		o,
			ok := item.(types.Object)
		if !ok ||
			o.IsNull() ||
			o.IsUnknown() {
			d.AddError("Invalid notification reference",
				"References must be known objects.")
			continue
		}
		a := o.Attributes()
		id,
			idOK := a["id"].(types.String)
		name,
			nOK := a["name"].(types.String)
		ref := &m.ContainerRefInput{}
		switch {
		case idOK &&
			knownString(id) &&
			id.ValueString() != "":
			ref.By = m.ObjectRefByID
			ref.Input = id.ValueString()
		case nOK &&
			knownString(name) &&
			name.ValueString() != "":
			ref.By = m.ObjectRefByName
			ref.Input = name.ValueString()
		default:
			d.AddError("Invalid notification reference",
				"Each reference needs a known nonempty id or name.")
			continue
		}
		result = append(result,
			ref)
	}
	return result
}

// The API download is base64 UTF-8, one canonical member per line, without a header.
func decodeMembers(encoded string, d *diag.Diagnostics) types.Set {
	const maxBytes = 256 * 1024 * 1024
	const maxItems = 1000000
	empty := types.SetNull(types.StringType)
	if len(encoded) > base64.StdEncoding.EncodedLen(maxBytes) {
		d.AddError("Invalid member download",
			"Member download exceeds the supported size.")
		return empty
	}
	data, err := io.ReadAll(io.LimitReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded)), maxBytes+1))
	if err != nil ||
		len(data) > maxBytes ||
		!utf8.Valid(data) {
		d.AddError("Invalid member download",
			"Expected bounded base64-encoded UTF-8 data.")
		return empty
	}
	values := []attr.Value{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte,
		4096),
		4096)
	seen := map[string]bool{}
	for scanner.Scan() {
		s := scanner.Text()
		if !validMember(s) ||
			seen[s] ||
			len(values) >= maxItems {
			d.AddError("Invalid member download",
				"API returned invalid, duplicate or excessive members.")
			return empty
		}
		seen[s] = true
		values = append(values,
			types.StringValue(s))
	}
	if scanner.Err() != nil {
		d.AddError("Invalid member download",
			"API member line exceeds the supported length.")
		return empty
	}
	return types.SetValueMust(types.StringType,
		values)
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
	if sync.FileType == nil ||
		!sync.FileType.IsValid() ||

		!sync.TimeUnit.IsValid() ||

		!validInterval(sync.TimeInterval) ||
		!validURL(sync.URL) {
		d.AddError("Invalid sync response",
			"API returned incomplete or invalid URL synchronization settings.")
		return p
	}
	p.FileType = types.StringValue(string(*sync.FileType))
	notification := sync.Notifications
	attrs := map[string]attr.Value{
		"mailing_list":       parseRefs(ctx, notification.GetMailingList(), d),
		"subscription_group": parseRefs(ctx, notification.GetSubscriptionGroup(), d),
		"webhook":            parseRefs(ctx, notification.GetWebhook(), d),
	}

	notifications, diags := types.ObjectValue(notificationsAttrTypes, attrs)
	d.Append(diags...)
	p.SyncData,
		diags = types.ObjectValue(syncAttrTypes,
		map[string]attr.Value{"notifications": notifications,
			"time_interval": types.Int64Value(sync.TimeInterval),
			"time_unit":     types.StringValue(string(sync.TimeUnit)),
			"url":           types.StringValue(sync.URL)})
	d.Append(diags...)
	return p
}
func identityState(p resourceModel,
	id string) resourceModel {
	p.ID = types.StringValue(id)
	p.Size = types.Int64Value(0)
	p.SyncData = normalizeObject(p.SyncData)
	return p
}
func normalizeObject(o types.Object) types.Object {
	if o.IsNull() ||
		o.IsUnknown() {
		return o
	}
	a := o.Attributes()
	for k, v := range a {
		switch value := v.(type) {
		case types.Object:
			a[k] = normalizeObject(value)
		case types.Set:
			elems := []attr.Value{}
			for _, e := range value.Elements() {
				if obj,
					ok := e.(types.Object); ok {
					e = normalizeObject(obj)
				}
				elems = append(elems,
					e)
			}
			if !value.IsNull() &&
				!value.IsUnknown() {
				a[k] = types.SetValueMust(value.ElementType(context.Background()),
					elems)
			}
		case types.String:
			if value.IsUnknown() {
				a[k] = types.StringValue("")
			}
		}
	}
	return types.ObjectValueMust(o.AttributeTypes(context.Background()),
		a)
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

func validSyncScalars(n types.Object, interval types.Int64, unit, u types.String) bool {
	if n.IsNull() || n.IsUnknown() || interval.IsNull() || interval.IsUnknown() {
		return false
	}
	return validInterval(interval.ValueInt64()) &&
		knownString(unit) &&
		m.ContainerSyncDataTimeUnit(unit.ValueString()).IsValid() &&
		knownString(u) &&
		validURL(u.ValueString())
}

func parseRefs[T interface {
	GetID() string
	GetName() string
}](ctx context.Context, refs []T, d *diag.Diagnostics) types.Set {
	values := []idname.Ref{}
	for _, ref := range refs {
		if ref.GetID() == "" {
			d.AddError("Invalid notification reference",
				"API returned a reference without an ID.")
			continue
		}
		values = append(values,
			idname.Ref{ID: types.StringValue(ref.GetID()),
				Name: types.StringValue(ref.GetName())})
	}
	result, diags := types.SetValueFrom(ctx, refType, values)
	d.Append(diags...)
	return result
}

func validateDestinations(sizes []int, d *diag.Diagnostics) {
	n := 0
	for _, size := range sizes {
		if size > 0 {
			n++
		}
	}
	if n != 1 {
		d.AddError("Invalid notifications", "Exactly one notification destination type must contain references.")
	}
}
