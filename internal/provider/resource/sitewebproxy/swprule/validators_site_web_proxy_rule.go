package swprule

import (
	"context"
	"strconv"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
			"KERBEROS requires a known nonempty encrypted_keytab. Raw keytab upload is "+
				"outside this resource.")
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
			"Omit kerberos_config when disabled and encrypted_keytab is empty. The API "+
				"default is represented by an omitted block.")
	}
}
