package common

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
)

type ProviderConfig struct {
	client *cmlclient.Client
	data   *cmlschema.ProviderModel
	mu     *sync.Mutex
}

func (r *ProviderConfig) Client() *cmlclient.Client {
	return r.client
}

func (r *ProviderConfig) UseNamedConfigs() bool {
	return r.data.NamedConfigs.ValueBool()
}

func (r *ProviderConfig) Lock() {
	r.mu.Lock()
}

func (r *ProviderConfig) Unlock() {
	r.mu.Unlock()
}

func NewProviderConfig(data *cmlschema.ProviderModel) *ProviderConfig {
	return &ProviderConfig{
		client: nil,
		mu:     new(sync.Mutex),
		data:   data,
	}
}

func (r *ProviderConfig) Initialize(ctx context.Context, diag diag.Diagnostics) *ProviderConfig {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client != nil {
		return r
	}

	// check if provided auth configuration makes sense
	if r.data.Token.IsNull() &&
		(r.data.Username.IsNull() || r.data.Password.IsNull()) {
		diag.AddError(
			"Required configuration missing",
			"null check: either username and password or a token must be provided",
		)
	}

	if len(r.data.Token.ValueString()) == 0 &&
		(len(r.data.Username.ValueString()) == 0 || len(r.data.Password.ValueString()) == 0) {
		diag.AddError(
			"Required configuration missing",
			"value check: either username and password or a token must be provided",
		)
	}

	if len(r.data.Token.ValueString()) > 0 && len(r.data.Username.ValueString()) > 0 {
		diag.AddWarning(
			"Conflicting configuration",
			"both token and username / password were provided")
	}

	// an address must be specified
	if len(r.data.Address.ValueString()) == 0 {
		diag.AddError(
			"Required configuration missing",
			"A server address must be configured to use the CML2 provider",
		)
	}

	// address must be https
	parsedURL, err := url.Parse(r.data.Address.ValueString())
	if err != nil {
		diag.AddError(
			"Can't parse server address / URL",
			err.Error(),
		)
	}

	// Check if the scheme is HTTPS and we have something like a hostname
	if parsedURL.Scheme != "https" || len(parsedURL.Host) == 0 {
		diag.AddError(
			"Invalid server address / URL, ensure it uses HTTPS",
			"A valid CML server URL using HTTPS must be provided.",
		)
	}

	if r.data.SkipVerify.IsNull() {
		tflog.Warn(ctx, "Unspecified certificate verification, will verify")
		r.data.SkipVerify = types.BoolValue(false)
	}

	if r.data.NamedConfigs.IsNull() {
		r.data.NamedConfigs = types.BoolValue(false)
	} else if r.data.NamedConfigs.ValueBool() {
		diag.AddWarning(
			"Feature",
			"\"named_configs\" is enabled",
		)
	}

	if r.data.UseCache.IsNull() {
		r.data.UseCache = types.BoolValue(false)
	} else if r.data.UseCache.ValueBool() {
		diag.AddError(
			"Experimental feature deprecated",
			"\"use_cache\" has been deprecated",
		)
	}

	// create a new CML2 client
	client := cmlclient.New(
		r.data.Address.ValueString(),
		r.data.SkipVerify.ValueBool(),
	)
	if len(r.data.Username.ValueString()) > 0 {
		client.SetUsernamePassword(
			r.data.Username.ValueString(),
			r.data.Password.ValueString(),
		)
	}
	if len(r.data.Token.ValueString()) > 0 {
		client.SetToken(r.data.Token.ValueString())
	}
	if r.data.NamedConfigs.ValueBool() {
		tflog.Warn(ctx, "Want to use named configurations")
		client.UseNamedConfigs()
	}
	if len(r.data.CAcert.ValueString()) > 0 {
		err := client.SetCACert([]byte(r.data.CAcert.ValueString()))
		if err != nil {
			diag.AddError(
				"Configuration issue",
				fmt.Sprintf("Provided certificate could not be used: %s", err),
			)
		}
	}
	r.client = client
	return r
}

func DatasourceConfigure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *ProviderConfig {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return nil
	}
	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *provider.ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return config.Initialize(ctx, resp.Diagnostics)
}

func ResourceConfigure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) *ProviderConfig {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return nil
	}

	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return config.Initialize(ctx, resp.Diagnostics)
}
