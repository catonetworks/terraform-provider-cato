package tests

import (
	"github.com/catonetworks/terraform-provider-cato/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cato": providerserver.NewProtocol6WithError(provider.New("test")()),
}
