package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/provider"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate tfplugindocs generate

const providerAddr = "registry.terraform.io/prismatic-io/prismatic"

var version string = "dev"

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: providerAddr,
		Debug:   debugMode,
	})
	if err != nil {
		log.Fatal(err)
	}
}
