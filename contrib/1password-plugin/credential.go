package main

import (
	"github.com/1Password/shell-plugins/sdk"
	"github.com/1Password/shell-plugins/sdk/importer"
	"github.com/1Password/shell-plugins/sdk/provision"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/1Password/shell-plugins/sdk/schema/credname"
)

const (
	fieldToken          = sdk.FieldName("Token")
	fieldGraph          = sdk.FieldName("Graph")
	fieldAPIURL         = sdk.FieldName("API URL")
	fieldTimeoutSeconds = sdk.FieldName("Timeout Seconds")
)

var envVarMapping = map[string]sdk.FieldName{
	"ROAM_API_TOKEN":       fieldToken,
	"ROAM_API_GRAPH":       fieldGraph,
	"ROAM_API_BASE_URL":    fieldAPIURL,
	"ROAM_TIMEOUT_SECONDS": fieldTimeoutSeconds,
}

// APIToken returns the credential type for Roam Research API Token.
func APIToken() schema.CredentialType {
	return schema.CredentialType{
		Name:          credname.APIToken,
		DocsURL:       sdk.URL("https://github.com/Leechael/roam-cli"),
		ManagementURL: sdk.URL("https://roamresearch.com/#/app/roam-cli-settings"),

		Fields: []schema.CredentialField{
			{
				Name:                fieldToken,
				MarkdownDescription: "Roam Research API token used to authenticate requests.",
				Secret:              true,
			},
			{
				Name:                fieldGraph,
				MarkdownDescription: "Roam Research graph name.",
				Secret:              false,
			},
			{
				Name:                fieldAPIURL,
				MarkdownDescription: "Optional custom API base URL.",
				Secret:              false,
				Optional:            true,
			},
			{
				Name:                fieldTimeoutSeconds,
				MarkdownDescription: "Optional request timeout in seconds.",
				Secret:              false,
				Optional:            true,
			},
		},

		DefaultProvisioner: provision.EnvVars(envVarMapping),

		Importer: importer.TryEnvVarPair(envVarMapping),
	}
}
