package main

import (
	"github.com/1Password/shell-plugins/sdk"
	"github.com/1Password/shell-plugins/sdk/needsauth"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/1Password/shell-plugins/sdk/schema/credname"
)

// RoamCLI returns the executable schema for the roam-cli binary.
func RoamCLI() schema.Executable {
	return schema.Executable{
		Name:    "Roam Research CLI",
		Runs:    []string{"roam-cli"},
		DocsURL: sdk.URL("https://github.com/Leechael/roam-cli"),
		NeedsAuth: needsauth.IfAll(
			needsauth.NotForHelpOrVersion(),
			needsauth.NotWithoutArgs(),
			needsauth.NotWhenContainsArgs("completion"),
			needsauth.NotWhenContainsArgs("__complete"),
			needsauth.NotWhenContainsArgs("__completeNoDesc"),
		),
		Uses: []schema.CredentialUsage{
			{Name: credname.APIToken},
		},
	}
}
