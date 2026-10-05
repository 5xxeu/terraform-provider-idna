package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/i386dev/terraform-provider-idna/internal/provider"
)

// Set by goreleaser
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/i386dev/idna",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
