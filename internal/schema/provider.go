package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProviderModel describes the provider configuration data model.
type ProviderModel struct {
	Address    types.String `tfsdk:"address"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Token      types.String `tfsdk:"token"`
	CAcert     types.String `tfsdk:"cacert"`
	SkipVerify types.Bool   `tfsdk:"skip_verify"`
	UseCache   types.Bool   `tfsdk:"use_cache"`
}

func Provider() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{

		"address": {
			MarkdownDescription: "CML2 controller address, must start with `https://`.",
			Required:            true,
			Type:                types.StringType,
		},
		"username": {
			Description: "CML2 username.",
			Optional:    true,
			Type:        types.StringType,
		},
		"password": {
			Description: "CML2 password.",
			Optional:    true,
			Type:        types.StringType,
			Sensitive:   true,
		},
		"token": {
			Description: "CML2 API token (JWT).",
			Optional:    true,
			Type:        types.StringType,
			Sensitive:   true,
		},
		"cacert": {
			Description: "A CA CERT, PEM encoded. When provided, the controller cert will be checked against it.  Otherwise, the system trust anchors will be used.",
			Optional:    true,
			Type:        types.StringType,
		},
		"skip_verify": {
			Description: "Disables TLS certificate verification.",
			Optional:    true,
			Type:        types.BoolType,
		},
		"use_cache": {
			Description: "Enables the client cache, this is considered experimental.",
			Optional:    true,
			Type:        types.BoolType,
		},
	}
}
