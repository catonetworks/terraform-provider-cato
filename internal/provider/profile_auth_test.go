package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// writeProfileAuthCredsFile writes a fixture catocli credentials file and points
// CATO_CREDENTIALS_FILE at it for the duration of the test.
func writeProfileAuthCredsFile(t *testing.T, contents string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "credentials")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	t.Setenv(credentialsFileEnvVar, path)
}

// withoutProfileAuthEnv clears every environment variable the provider's auth
// resolution consults, so each test starts from a blank slate regardless of the
// host shell's environment.
func withoutProfileAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISABLE_POLICY_RULE_CLEANUP", "true")
	t.Setenv("CATO_BASEURL", "")
	t.Setenv("CATO_TOKEN", "")
	t.Setenv("CATO_ACCOUNT_ID", "")
	t.Setenv("CATO_PROFILE", "")
	t.Setenv(credentialsFileEnvVar, "")
	t.Setenv("CATO_RETRY_MAX", "")
	t.Setenv("CATO_RETRY_WAIT_MIN_SECONDS", "")
	t.Setenv("CATO_RETRY_WAIT_MAX_SECONDS", "")
	t.Setenv("CATO_VERSION_CHECK_DISABLED", "")
	t.Setenv("CATO_VERSION_CHECK_TIMEOUT_SECONDS", "")
}

// profileAuthConfig builds a tfsdk.Config for the provider from a partial model.
// Version checking is always disabled and unset numeric fields fall through to
// their defaults, so tests only need to specify the auth-relevant fields.
func profileAuthConfig(ctx context.Context, t *testing.T, model catoProviderModel) tfsdk.Config {
	t.Helper()

	schemaResp := &frameworkprovider.SchemaResponse{}
	(&catoProvider{}).Schema(ctx, frameworkprovider.SchemaRequest{}, schemaResp)

	model.VersionCheckDisabled = types.BoolValue(true)

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags := plan.Set(ctx, model)
	require.False(t, diags.HasError(), "unexpected config diagnostics: %v", diags)

	return tfsdk.Config(plan)
}

func configureCatoProvider(ctx context.Context, t *testing.T, model catoProviderModel) *frameworkprovider.ConfigureResponse {
	t.Helper()
	p := &catoProvider{version: "1.5.0-test"}
	resp := &frameworkprovider.ConfigureResponse{}
	config := profileAuthConfig(ctx, t, model)
	p.Configure(ctx, frameworkprovider.ConfigureRequest{Config: config}, resp)
	return resp
}

func TestCatoProviderSchemaExposesProfile(t *testing.T) {
	t.Parallel()

	resp := &frameworkprovider.SchemaResponse{}
	(&catoProvider{}).Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	attribute, ok := resp.Schema.Attributes["profile"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, attribute.Optional)
	require.False(t, attribute.Sensitive)
	require.Contains(t, attribute.Description, "CATO_PROFILE")
	require.Contains(t, attribute.Description, "CATO_CREDENTIALS_FILE")
}

func TestCatoProviderSchemaAccountIDIsNowOptional(t *testing.T) {
	t.Parallel()

	resp := &frameworkprovider.SchemaResponse{}
	(&catoProvider{}).Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	attribute, ok := resp.Schema.Attributes["account_id"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, attribute.Optional, "account_id must be optional so profile-based auth can supply it")
	require.False(t, attribute.Required)
}

// Explicit credentials AND a profile are mutually exclusive -- supplying both is a hard
// error (not a precedence fallback). This is the core of the mutually-exclusive auth model.
func TestCatoProviderConfigureExplicitAndProfileIsConflictError(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = profile-token
account_id = profile-account
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		BaseURL:   types.StringValue("https://explicit.example.com/graphql2"),
		Token:     types.StringValue("explicit-token"),
		AccountID: types.StringValue("explicit-account"),
		Profile:   types.StringValue("refarch-dev"),
	})

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, conflictSummary(resp), "mutually exclusive")
}

// Env token (explicit credentials via CATO_TOKEN) AND a profile attribute is likewise a
// conflict -- the env-aware exclusivity check catches a stray env var colliding with profile.
func TestCatoProviderConfigureEnvTokenAndProfileIsConflictError(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()
	t.Setenv("CATO_TOKEN", "env-token")

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = profile-token
account_id = profile-account
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		BaseURL:   types.StringValue("https://explicit.example.com/graphql2"),
		AccountID: types.StringValue("explicit-account"),
		Profile:   types.StringValue("refarch-dev"),
	})

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, conflictSummary(resp), "mutually exclusive")
}

// Explicit credentials fully from environment variables (CATO_TOKEN / CATO_ACCOUNT_ID /
// CATO_BASEURL) with no provider-block attributes and no profile -- the CI path (Azure
// DevOps / Terraform Cloud). Exercises the new CATO_ACCOUNT_ID env support.
func TestCatoProviderConfigureExplicitCredsFromEnvVars(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()
	t.Setenv("CATO_TOKEN", "ci-token")
	t.Setenv("CATO_ACCOUNT_ID", "ci-account")
	t.Setenv("CATO_BASEURL", "https://ci.example.com/graphql2")

	resp := configureCatoProvider(ctx, t, catoProviderModel{})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "ci-token", svc.Token)
	require.Equal(t, "ci-account", svc.AccountId)
	require.Equal(t, "https://ci.example.com/graphql2", svc.BaseURL)
}

// A configured `profile` attribute is authoritative: ambient explicit env
// credentials (CATO_TOKEN / CATO_ACCOUNT_ID -- e.g. a leftover shell export for
// another tenant) must NOT conflict with or override it. The profile wins and the
// env is ignored.
func TestCatoProviderConfigureProfileAttrWinsOverAmbientEnv(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()
	t.Setenv("CATO_TOKEN", "ambient-wrong-token")
	t.Setenv("CATO_ACCOUNT_ID", "15412") // ambient refarch-like leftover

	writeProfileAuthCredsFile(t, `[badev4]
cato_token = profile-token
account_id = 24425
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		Profile: types.StringValue("badev4"),
	})

	require.False(t, resp.Diagnostics.HasError(), "profile must win over ambient env: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "profile-token", svc.Token)
	require.Equal(t, "24425", svc.AccountId)
	require.Equal(t, "https://profile.example.com/graphql2", svc.BaseURL)
}

// conflictSummary returns the detail of the "Conflicting Cato API Credentials" diagnostic
// (empty if absent).
func conflictSummary(resp *frameworkprovider.ConfigureResponse) string {
	for _, d := range resp.Diagnostics {
		if strings.Contains(d.Summary(), "Conflicting Cato API Credentials") {
			return d.Detail()
		}
	}
	return ""
}

func TestCatoProviderConfigureResolvesFromProfileAttributeWhenFieldsMissing(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = profile-token
account_id = profile-account
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		Profile: types.StringValue("refarch-dev"),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "profile-token", svc.Token)
	require.Equal(t, "profile-account", svc.AccountId)
	require.Equal(t, "https://profile.example.com/graphql2", svc.BaseURL)
}

func TestCatoProviderConfigureResolvesFromCatoProfileEnvVar(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()
	t.Setenv("CATO_PROFILE", "refarch-dev")

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = profile-token
account_id = profile-account
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "profile-token", svc.Token)
	require.Equal(t, "profile-account", svc.AccountId)
}

func TestCatoProviderConfigureProfileAttributeOverridesCatoProfileEnvVar(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()
	t.Setenv("CATO_PROFILE", "env-profile")

	writeProfileAuthCredsFile(t, `[env-profile]
cato_token = env-profile-token
account_id = env-profile-account
endpoint = https://env-profile.example.com/graphql2

[attr-profile]
cato_token = attr-profile-token
account_id = attr-profile-account
endpoint = https://attr-profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		Profile: types.StringValue("attr-profile"),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "attr-profile-token", svc.Token)
	require.Equal(t, "attr-profile-account", svc.AccountId)
}

func TestCatoProviderConfigureExplicitBaseURLWinsOverProfileEndpoint(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = profile-token
account_id = profile-account
endpoint = https://profile.example.com/graphql2
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		BaseURL: types.StringValue("https://explicit.example.com/graphql2"),
		Profile: types.StringValue("refarch-dev"),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	svc, ok := resp.ResourceData.(*catoclient.Service)
	require.True(t, ok)
	require.Equal(t, "https://explicit.example.com/graphql2", svc.BaseURL)
	require.Equal(t, "profile-token", svc.Token)
}

func TestCatoProviderConfigureProfileNotFoundProducesAttributeErrorWithoutLeakingSecrets(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	writeProfileAuthCredsFile(t, `[refarch-dev]
cato_token = super-secret-token
account_id = profile-account
`)

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		Profile: types.StringValue("does-not-exist"),
	})

	require.True(t, resp.Diagnostics.HasError())
	found := false
	for _, d := range resp.Diagnostics {
		detail := d.Detail()
		if strings.Contains(detail, "refarch-dev") {
			found = true
		}
		require.NotContains(t, detail, "super-secret-token", "diagnostics must never include credential values")
	}
	require.True(t, found, "expected an error listing available profiles, got: %v", resp.Diagnostics)
}

func TestCatoProviderConfigureNoCredentialsAnywhereProducesCombinedAuthError(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	resp := configureCatoProvider(ctx, t, catoProviderModel{})

	require.True(t, resp.Diagnostics.HasError())
	var summary string
	for _, d := range resp.Diagnostics {
		if strings.Contains(d.Summary(), "Unable to Authenticate") {
			summary = d.Detail()
		}
	}
	require.NotEmpty(t, summary, "expected a combined authentication error, got: %v", resp.Diagnostics)
	require.Contains(t, summary, "token")
	require.Contains(t, summary, "CATO_TOKEN")
	require.Contains(t, summary, "profile")
	require.Contains(t, summary, "CATO_PROFILE")
}

func TestCatoProviderConfigurePartialExplicitConfigErrorsOnMissingField(t *testing.T) {
	withoutProfileAuthEnv(t)
	ctx := context.Background()

	resp := configureCatoProvider(ctx, t, catoProviderModel{
		BaseURL: types.StringValue("https://explicit.example.com/graphql2"),
		Token:   types.StringValue("explicit-token"),
		// AccountID intentionally omitted, and no profile configured.
	})

	require.True(t, resp.Diagnostics.HasError())
	foundAccountIDError := false
	for _, d := range resp.Diagnostics {
		if strings.Contains(d.Summary(), "account_id") {
			foundAccountIDError = true
			require.Contains(t, d.Detail(), "profile", "the error should mention profile-based auth as an alternative")
		}
	}
	require.True(t, foundAccountIDError, "expected a clear account_id-specific error, got: %v", resp.Diagnostics)
}

func TestCatoProviderConfigureEnvProfileCannotOverrideExplicitAttributes(t *testing.T) {
	withoutProfileAuthEnv(t)
	t.Setenv("CATO_PROFILE", "another-account")
	resp := configureCatoProvider(context.Background(), t, catoProviderModel{
		BaseURL:   types.StringValue("https://explicit.example.com/graphql2"),
		Token:     types.StringValue("explicit-token"),
		AccountID: types.StringValue("12345"),
	})
	require.True(t, resp.Diagnostics.HasError())
	require.NotEmpty(t, conflictSummary(resp))
	require.Nil(t, resp.ResourceData)
}
