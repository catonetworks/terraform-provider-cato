package swprule

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type knownValue struct {
	name     string
	value    attr.Value
	required bool
}

func validateValues(diags *diag.Diagnostics, values ...knownValue) {
	for _, value := range values {
		if elements, ok := value.value.(types.Set); ok {
			for _, element := range elements.Elements() {
				if element.IsNull() || element.IsUnknown() {
					diags.AddError("Invalid configuration", value.name+" must not contain null or unknown elements.")
					break
				}
			}
		}
		if value.value.IsUnknown() || (value.required && value.value.IsNull()) {
			diags.AddError("Invalid configuration", value.name+" must contain a known value.")
		}
	}
}

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		diags.AddError("Invalid resource ID", "A known, non-empty resource ID is required in state.")
		return false
	}
	return true
}

func validatePosition(at *cato_models.PolicyRulePositionInput, diags *diag.Diagnostics) {
	if at == nil {
		return
	}
	if at.Position == nil {
		diags.AddError("Missing position", "A position is required when at is configured.")
		return
	}
	position := *at.Position
	if !slices.Contains(cato_models.AllPolicyRulePositionEnum, position) {
		diags.AddError("Invalid position", fmt.Sprintf("Unsupported position %q.", position))
		return
	}
	if position != cato_models.PolicyRulePositionEnumLastInPolicy && position != cato_models.PolicyRulePositionEnumFirstInPolicy {
		if at.Ref == nil || *at.Ref == "" {
			diags.AddError("Missing position reference", "Relative placement requires a known, non-empty ref.")
		}
	}
}

func validateProxyRule(ctx context.Context, model RuleModel, d *diag.Diagnostics) {
	const maxPort = 65535
	port, err := strconv.Atoi(model.Port.ValueString())
	if err != nil || port < 1 || port > maxPort || strconv.Itoa(port) != model.Port.ValueString() {
		d.AddError("Invalid proxy port", "Use a canonical decimal port from 1 through 65535.")
	}
	if strings.TrimSpace(model.Fqdn.ValueString()) == "" || strings.TrimSpace(model.Name.ValueString()) == "" {
		d.AddError("Invalid proxy rule", "Name and FQDN must not be empty.")
	}
	if !model.ShouldAssociateWithAllSites.ValueBool() && len(model.AssociatedSite.Elements()) == 0 {
		d.AddError("Missing associated sites", "Provide at least one associated_site when should_associate_with_all_sites is false.")
	}
	validateKerberos(ctx, model.AuthenticationConfig, d)
}
func validateKerberos(ctx context.Context, value types.Object, d *diag.Diagnostics) {
	var auth RuleAuthenticationConfigModel
	d.Append(value.As(ctx, &auth, basetypes.ObjectAsOptions{})...)
	if d.HasError() {
		return
	}
	kerberosMethod := auth.Method.ValueString() == string(cato_models.SiteWebProxyAuthMethodKerberos)
	if auth.KerberosConfig.IsNull() ||
		auth.KerberosConfig.IsUnknown() {
		if kerberosMethod {
			d.AddError("Missing Kerberos configuration",
				"KERBEROS requires a known configuration and encrypted keytab.")
		}
		return
	}
	var k RuleAuthenticationConfigKerberosConfigModel
	d.Append(auth.KerberosConfig.As(ctx,
		&k,
		basetypes.ObjectAsOptions{})...)
	if d.HasError() {
		return
	}
	validateEmptyKerberos(k, d)
	if kerberosMethod &&
		(k.EncryptedKeytab.IsNull() ||
			k.EncryptedKeytab.IsUnknown() ||
			strings.TrimSpace(k.EncryptedKeytab.ValueString()) == "") {
		d.AddError("Missing encrypted keytab",
			"KERBEROS requires a known nonempty encrypted_keytab. Raw keytab upload is outside this resource.")
	}
}
func validateEmptyKerberos(k RuleAuthenticationConfigKerberosConfigModel, d *diag.Diagnostics) {
	if k.IsEnabled.IsNull() || k.IsEnabled.IsUnknown() || k.IsEnabled.ValueBool() {
		return
	}
	if k.EncryptedKeytab.IsNull() ||
		(!k.EncryptedKeytab.IsUnknown() &&
			k.EncryptedKeytab.ValueString() == "") {
		d.AddError("Empty Kerberos configuration",
			"Omit kerberos_config when disabled and encrypted_keytab is empty. The API default is represented by an omitted block.")
	}
}
