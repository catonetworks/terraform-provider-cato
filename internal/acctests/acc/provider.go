//go:build acctest

package acc

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/catonetworks/terraform-provider-cato/internal/provider"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cato": providerserver.NewProtocol6WithError(provider.New("test")()),
}
