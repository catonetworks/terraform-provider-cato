package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider"
)

var version = "0.0.0-poc"

func main() {
	debug := flag.Bool("debug", false, "enable provider debugging")
	flag.Parse()
	err := providerserver.Serve(context.Background(), provider.NewProvider(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/catonetworks/cato-refactor",
		Debug:   *debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
