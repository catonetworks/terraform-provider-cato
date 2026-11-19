package networkconfig

import (
	cato "github.com/catonetworks/cato-go-sdk"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// partialResponseClient keeps request headers, retries and SDK tracing while allowing
// these adapters to retain acknowledged identities from GraphQL data+errors.
// The shared provider client is never toggled: the SDK CustomDo captures its own
// client pointer, so it must be constructed for this private decoder instance.
func partialResponseClient(original *catoclient.ProviderSDKClient) *catoclient.ProviderSDKClient {
	if original == nil || original.Client == nil || original.Client.Client == nil {
		return original
	}
	transport := original.Client.Client
	scoped, err := cato.New(transport.BaseURL, "", "", nil, nil)
	if err != nil {
		// The pinned SDK constructor cannot fail. Keep the configured transport if a
		// later SDK adds constructor validation; contract tests cover partial decoding.
		return original
	}
	scoped.Client.Client = transport.Client
	scoped.Client.RequestInterceptor = transport.RequestInterceptor
	scoped.Client.IsUnsafeRequestInterceptor = transport.IsUnsafeRequestInterceptor
	scoped.Client.ParseDataWhenErrors = true
	return catoclient.NewProviderSDKClient(scoped)
}
