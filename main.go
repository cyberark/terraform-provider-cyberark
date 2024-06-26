// Package main provides the entrypoint for the SecretsHub Terraform provider.
package main

import (
	"context"
	"log"

	"github.com/aharriscybr/terraform-provider-cybr-sh/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name secretshub

var (
	version = "dev"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/cyberark/secretshub",
		Debug:   false,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
