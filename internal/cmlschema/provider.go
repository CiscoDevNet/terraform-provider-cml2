package cmlschema

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
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

func Provider() map[string]schema.Attribute {
	return map[string]schema.Attribute{

		"address": schema.StringAttribute{
			MarkdownDescription: "CML2 controller address, must start with `https://`.",
			Required:            true,
		},
		"username": schema.StringAttribute{
			Description: "CML2 username.",
			Optional:    true,
		},
		"password": schema.StringAttribute{
			Description: "CML2 password.",
			Optional:    true,
			Sensitive:   true,
		},
		"token": schema.StringAttribute{
			Description: "CML2 API token (JWT).",
			Optional:    true,
			Sensitive:   true,
		},
		"cacert": schema.StringAttribute{
			Description: "A CA CERT, PEM encoded. When provided, the controller cert will be checked against it.  Otherwise, the system trust anchors will be used.",
			Optional:    true,
		},
		"skip_verify": schema.BoolAttribute{
			Description: "Disables TLS certificate verification (default is false -- will not skip / it will verify the certificate!)",
			Optional:    true,
		},
		"use_cache": schema.BoolAttribute{
			Description: "Enables the client cache, this is considered experimental (default is false -- will not use the cache!)",
			Optional:    true,
		},
	}
}
