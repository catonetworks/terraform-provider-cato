package tlsrule

import (
	"context"
	"testing"

	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
)

// nullAttrs returns a value map with every attribute null, overlaid by set.
func nullAttrs(t *testing.T, types_ map[string]attr.Type, set map[string]attr.Value) map[string]attr.Value {
	t.Helper()
	out := make(map[string]attr.Value, len(types_))
	for k, ty := range types_ {
		switch v := ty.(type) {
		case types.ObjectType:
			out[k] = types.ObjectNull(v.AttrTypes)
		case types.ListType:
			out[k] = types.ListNull(v.ElemType)
		case types.SetType:
			out[k] = types.SetNull(v.ElemType)
		default:
			switch ty {
			case types.BoolType:
				out[k] = types.BoolNull()
			case types.Int64Type:
				out[k] = types.Int64Null()
			default:
				out[k] = types.StringNull()
			}
		}
	}
	for k, v := range set {
		out[k] = v
	}
	return out
}

func mustObject(t *testing.T, attrTypes map[string]attr.Type, vals map[string]attr.Value) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(attrTypes, vals)
	if diags.HasError() {
		t.Fatalf("building object: %v", diags)
	}
	return obj
}

// TestHydrateTLSRuleAPI_CustomService checks that custom_service and
// custom_service_ip reach both the add and the update request. They used to
// be dropped, so a rule such as "Bypass RDP (TCP/3389,3390)" was created with
// no service filter at all.
func TestHydrateTLSRuleAPI_CustomService(t *testing.T) {
	ctx := context.Background()

	ports, _ := types.ListValueFrom(ctx, types.StringType, []string{"3389", "3390"})
	customService := mustObject(t, customservice.CustomServiceAttrTypes, map[string]attr.Value{
		"port":       ports,
		"port_range": types.ObjectNull(customservice.CustomServiceAttrTypes["port_range"].(types.ObjectType).AttrTypes),
		"protocol":   types.StringValue("TCP"),
	})
	customServiceIP := mustObject(t, customservice.CustomServiceIPAttrTypes, map[string]attr.Value{
		"name": types.StringValue("rdp-hosts"),
		"ip":   types.StringNull(),
		"ip_range": mustObject(t, iprange.RangeObjectType.AttrTypes, map[string]attr.Value{
			"from": types.StringValue("10.0.0.1"),
			"to":   types.StringValue("10.0.0.9"),
		}),
	})

	application := mustObject(t, TLSApplicationAttrTypes, nullAttrs(t, TLSApplicationAttrTypes, map[string]attr.Value{
		"custom_service":    customService,
		"custom_service_ip": customServiceIP,
	}))
	rule := mustObject(t, TLSInspectionRuleRuleAttrTypes, nullAttrs(t, TLSInspectionRuleRuleAttrTypes, map[string]attr.Value{
		"name":        types.StringValue("Bypass RDP (TCP/3389,3390)"),
		"enabled":     types.BoolValue(true),
		"action":      types.StringValue("BYPASS"),
		"application": application,
	}))
	plan := InspectionRule{
		At:   types.ObjectNull(map[string]attr.Type{"position": types.StringType, "ref": types.StringType}),
		Rule: rule,
		ID:   types.StringNull(),
	}

	got, diags := hydrateTLSRuleAPI(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	for name, app := range map[string]struct {
		svcLen, ipLen int
		port          []cato_scalars.Port
		proto         string
		ipName        string
		from          string
	}{
		"create": {
			svcLen: len(got.create.Rule.Application.CustomService),
			ipLen:  len(got.create.Rule.Application.CustomServiceIP),
		},
		"update": {
			svcLen: len(got.update.Rule.Application.CustomService),
			ipLen:  len(got.update.Rule.Application.CustomServiceIP),
		},
	} {
		if app.svcLen != 1 || app.ipLen != 1 {
			t.Fatalf("%s: want 1 custom service and 1 custom service ip, got %d and %d", name, app.svcLen, app.ipLen)
		}
	}

	svc := got.create.Rule.Application.CustomService[0]
	if string(svc.Protocol) != "TCP" || len(svc.Port) != 2 || svc.Port[0] != "3389" || svc.Port[1] != "3390" || svc.PortRange != nil {
		t.Errorf("custom service not carried faithfully: %+v", svc)
	}
	svcIP := got.create.Rule.Application.CustomServiceIP[0]
	if svcIP.Name != "rdp-hosts" || svcIP.IP != nil || svcIP.IPRange == nil || svcIP.IPRange.From != "10.0.0.1" || svcIP.IPRange.To != "10.0.0.9" {
		t.Errorf("custom service ip not carried faithfully: %+v", svcIP)
	}
}
