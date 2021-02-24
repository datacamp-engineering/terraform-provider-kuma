package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/nickvdyck/terraform-provider-kuma/kuma"
)

func main() {
	opt := &plugin.ServeOpts{
		ProviderFunc: kuma.Provider,
	}
	plugin.Serve(opt)
}
