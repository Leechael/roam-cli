package main

import (
	"github.com/1Password/shell-plugins/sdk"
	"github.com/1Password/shell-plugins/sdk/schema"
)

// New returns the Roam Research 1Password shell plugin definition.
func New() schema.Plugin {
	return schema.Plugin{
		Name: "roamresearch",
		Platform: schema.PlatformInfo{
			Name:     "Roam Research",
			Homepage: sdk.URL("https://roamresearch.com"),
		},
		Credentials: []schema.CredentialType{APIToken()},
		Executables: []schema.Executable{RoamCLI()},
	}
}
