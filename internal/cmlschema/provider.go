package cmlschema

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProviderModel describes the provider configuration data model.
type ProviderModel struct {
	Address        types.String `tfsdk:"address"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	Token          types.String `tfsdk:"token"`
	RequestHeaders types.Map    `tfsdk:"request_headers"`
	TokenCache     types.Bool   `tfsdk:"token_cache"`
	TokenCacheFile types.String `tfsdk:"token_cache_file"`
	CAcert         types.String `tfsdk:"cacert"`
	SkipVerify     types.Bool   `tfsdk:"skip_verify"`
	UseCache       types.Bool   `tfsdk:"use_cache"`
	NamedConfigs   types.Bool   `tfsdk:"named_configs"`
	DynamicConfig  types.Bool   `tfsdk:"dynamic_config"`
}

// ApplyEnvVars fills unset (null) provider attributes from their corresponding environment variables.
func (m *ProviderModel) ApplyEnvVars() diag.Diagnostics {
	var diags diag.Diagnostics

	applyString := func(target *types.String, env string) {
		if !target.IsNull() {
			return
		}
		if v, ok := os.LookupEnv(env); ok {
			*target = types.StringValue(v)
		}
	}

	applyBool := func(target *types.Bool, env string) {
		if !target.IsNull() {
			return
		}
		v, ok := os.LookupEnv(env)
		if !ok {
			return
		}
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			diags.AddError(
				fmt.Sprintf("Invalid boolean environment variable for %q=%q. Use 1/t/T/TRUE/true/True/0/f/F/FALSE/false/False instead.", env, v),
				err.Error(),
			)
			return
		}
		*target = types.BoolValue(parsed)
	}

	applyString(&m.Address, "CML2_ADDRESS")
	applyString(&m.Username, "CML2_USERNAME")
	applyString(&m.Password, "CML2_PASSWORD")
	applyString(&m.Token, "CML2_TOKEN")
	applyString(&m.TokenCacheFile, "CML2_TOKEN_CACHE_FILE")
	applyString(&m.CAcert, "CML2_CACERT")

	applyBool(&m.TokenCache, "CML2_TOKEN_CACHE")
	applyBool(&m.SkipVerify, "CML2_SKIP_VERIFY")
	applyBool(&m.NamedConfigs, "CML2_NAMED_CONFIGS")
	applyBool(&m.DynamicConfig, "CML2_DYNAMIC_CONFIG")

	return diags
}

// Provider returns the schema for provider configuration.
func Provider() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"address": schema.StringAttribute{
			MarkdownDescription: "CML2 controller address, must start with `https://`. Can also be set via the `CML2_ADDRESS` environment variable.",
			Optional:            true,
		},
		"username": schema.StringAttribute{
			Description: "CML2 username. Can also be set via the CML2_USERNAME environment variable.",
			Optional:    true,
		},
		"password": schema.StringAttribute{
			Description: "CML2 password. Can also be set via the CML2_PASSWORD environment variable.",
			Optional:    true,
			Sensitive:   true,
		},
		"token": schema.StringAttribute{
			Description: "CML2 API token (JWT). Can also be set via the CML2_TOKEN environment variable.",
			Optional:    true,
			Sensitive:   true,
		},
		"request_headers": schema.MapAttribute{
			Description: "Static HTTP headers to inject into every outbound CML client request, including authentication bootstrap requests.",
			Optional:    true,
			Sensitive:   true,
			ElementType: types.StringType,
		},
		"token_cache": schema.BoolAttribute{
			Description: "Enables caching of an auth token in a local file when using username/password. Ignored when `token` is set. Can also be set via the CML2_TOKEN_CACHE environment variable.",
			Optional:    true,
		},
		"token_cache_file": schema.StringAttribute{
			Description: "Path to the token cache file. Used only when `token_cache=true` and username/password auth is used. Can also be set via the CML2_TOKEN_CACHE_FILE environment variable.",
			Optional:    true,
		},
		"cacert": schema.StringAttribute{
			Description: "A CA CERT, PEM encoded. When provided, the controller cert will be checked against it.  Otherwise, the system trust anchors will be used. Can also be set via the CML2_CACERT environment variable.",
			Optional:    true,
		},
		"skip_verify": schema.BoolAttribute{
			Description: "Disables TLS certificate verification (default is false -- will not skip / it will verify the certificate!). Can also be set via the CML2_SKIP_VERIFY environment variable.",
			Optional:    true,
		},
		"use_cache": schema.BoolAttribute{
			Description:        "Enables the client cache, **Deprecated**",
			DeprecationMessage: "This has been deprecated, wasn't really useful and potentially buggy",
			Optional:           true,
		},
		"named_configs": schema.BoolAttribute{
			Description: "Enables the use of named configs (CML version >2.7.0 required!). Can also be set via the CML2_NAMED_CONFIGS environment variable.",
			Optional:    true,
		},
		"dynamic_config": schema.BoolAttribute{
			MarkdownDescription: "Does late binding of the provider configuration. If set to `true` then provider configuration errors will only be caught when resources and data sources are actually created/read. Defaults to `false`. Can also be set via the CML2_DYNAMIC_CONFIG environment variable.",
			Optional:            true,
		},
	}
}
