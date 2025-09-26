package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/robertdebock/terraform-provider-orbstack/internal/provider"
)

var (
	// Set by goreleaser or -ldflags at build time
	version = "dev"
)

func main() {
	// Allow an override for Terraform context cancellation testing
	debug := flag.Bool("debug", false, "run provider with debug support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/robertdebock/orbstack",
		Debug:   *debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
