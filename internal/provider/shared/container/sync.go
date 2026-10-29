package container

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"math"
	"net/url"
	"strings"
	"unicode/utf8"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// ReferenceType is the common id/name reference element for sync notifications.
var ReferenceType = types.ObjectType{AttrTypes: idname.RefAttrTypes}

// NotificationAttributeTypes describes the three mutually exclusive destinations.
var NotificationAttributeTypes = map[string]attr.Type{"mailing_list": types.SetType{ElemType: ReferenceType},
	"subscription_group": types.SetType{ElemType: ReferenceType},
	"webhook":            types.SetType{ElemType: ReferenceType}}

// SyncAttributeTypes describes URL-owned synchronization settings.
var SyncAttributeTypes = map[string]attr.Type{"notifications": types.ObjectType{AttrTypes: NotificationAttributeTypes},
	"time_interval": types.Int64Type,
	"time_unit":     types.StringType,
	"url":           types.StringType}

// KnownString retains common container synchronization semantics.
func KnownString(s types.String) bool { return !s.IsNull() && !s.IsUnknown() }

// ValidURL retains common container synchronization semantics.
func ValidURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Fragment == ""
}

// ValidInterval retains common container synchronization semantics.
func ValidInterval(v int64) bool { return v > 0 && v <= math.MaxInt32 }

// ReferenceSet retains common container synchronization semantics.
func ReferenceSet(value attr.Value, d *diag.Diagnostics) []*m.ContainerRefInput {
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
			KnownString(id) &&
			id.ValueString() != "":
			ref.By = m.ObjectRefByID
			ref.Input = id.ValueString()
		case nOK &&
			KnownString(name) &&
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

// validSyncScalars retains common container synchronization semantics.
func validSyncScalars(n types.Object, interval types.Int64, unit, u types.String) bool {
	if n.IsNull() || n.IsUnknown() || interval.IsNull() || interval.IsUnknown() {
		return false
	}
	return ValidInterval(interval.ValueInt64()) &&
		KnownString(unit) &&
		m.ContainerSyncDataTimeUnit(unit.ValueString()).IsValid() &&
		KnownString(u) &&
		ValidURL(u.ValueString())
}

// ValidateDestinations retains common container synchronization semantics.
func ValidateDestinations(sizes []int, d *diag.Diagnostics) {
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

// ParseRefs retains common container synchronization semantics.
func ParseRefs[T interface {
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
	result, diags := types.SetValueFrom(ctx, ReferenceType, values)
	d.Append(diags...)
	return result
}

// NormalizeObject retains common container synchronization semantics.
func NormalizeObject(o types.Object) types.Object {
	if o.IsNull() ||
		o.IsUnknown() {
		return o
	}
	a := o.Attributes()
	for k, v := range a {
		switch value := v.(type) {
		case types.Object:
			a[k] = NormalizeObject(value)
		case types.Set:
			elems := []attr.Value{}
			for _, e := range value.Elements() {
				if obj,
					ok := e.(types.Object); ok {
					e = NormalizeObject(obj)
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

// BuildSync builds the URL configuration without taking ownership of fetched members.
func BuildSync(fileType types.String, syncData types.Object, d *diag.Diagnostics) *m.CreateContainerSyncDataInput {
	kind := m.ContainerFileType(fileType.ValueString())
	if !KnownString(fileType) ||
		!kind.IsValid() ||
		syncData.IsNull() ||
		syncData.IsUnknown() {
		d.AddError("Invalid sync configuration",
			"URL mode requires known file_type and sync_data.")
		return nil
	}
	a := syncData.Attributes()
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
		return nil
	}

	refs := n.Attributes()
	ml, sg, wh := ReferenceSet(refs["mailing_list"], d), ReferenceSet(refs["subscription_group"], d), ReferenceSet(refs["webhook"], d)
	ValidateDestinations([]int{len(ml), len(sg), len(wh)}, d)

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
	return &m.CreateContainerSyncDataInput{Notifications: notify,
		TimeInterval: interval.ValueInt64(),
		TimeUnit:     m.ContainerSyncDataTimeUnit(unit.ValueString()),
		URL:          u.ValueString()}
}

// DecodeMembers decodes the bounded API file, validating each complete member before hydration.
func DecodeMembers(encoded string, valid func(string) bool, d *diag.Diagnostics) []string {
	const maxBytes = 256 * 1024 * 1024
	const maxItems = 1000000
	var empty []string
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
	values := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte,
		4096),
		4096)
	seen := map[string]bool{}
	for scanner.Scan() {
		s := scanner.Text()
		if !valid(s) ||
			seen[s] ||
			len(values) >= maxItems {
			d.AddError("Invalid member download",
				"API returned invalid, duplicate or excessive members.")
			return empty
		}
		seen[s] = true
		values = append(values,
			s)
	}
	if scanner.Err() != nil {
		d.AddError("Invalid member download",
			"API member line exceeds the supported length.")
		return empty
	}
	return values
}

// ParseSync hydrates URL-owned metadata and notification references.
func ParseSync(ctx context.Context,
	sync *cato.ContainerList_Container_List_Containers_SyncData,
	d *diag.Diagnostics) (fileType types.String,
	object types.Object) {
	fileType, object = types.StringNull(), types.ObjectNull(SyncAttributeTypes)
	if sync == nil {
		d.AddError("Invalid sync response", "API returned no URL synchronization settings.")
		return fileType, object
	}

	if sync.FileType == nil ||
		!sync.FileType.IsValid() ||

		!sync.TimeUnit.IsValid() ||

		!ValidInterval(sync.TimeInterval) ||
		!ValidURL(sync.URL) {
		d.AddError("Invalid sync response",
			"API returned incomplete or invalid URL synchronization settings.")
		return fileType, object
	}
	fileType = types.StringValue(string(*sync.FileType))
	notification := sync.Notifications
	attrs := map[string]attr.Value{
		"mailing_list":       ParseRefs(ctx, notification.GetMailingList(), d),
		"subscription_group": ParseRefs(ctx, notification.GetSubscriptionGroup(), d),
		"webhook":            ParseRefs(ctx, notification.GetWebhook(), d),
	}

	notifications, diags := types.ObjectValue(NotificationAttributeTypes, attrs)
	d.Append(diags...)
	object,
		diags = types.ObjectValue(SyncAttributeTypes,
		map[string]attr.Value{"notifications": notifications,
			"time_interval": types.Int64Value(sync.TimeInterval),
			"time_unit":     types.StringValue(string(sync.TimeUnit)),
			"url":           types.StringValue(sync.URL)})
	d.Append(diags...)
	return fileType, object
}
