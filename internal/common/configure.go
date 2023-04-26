package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

type ProviderConfig struct {
	client *cmlclient.Client
	data   *cmlschema.ProviderModel
	mu     *sync.Mutex
}

func (r *ProviderConfig) Client() *cmlclient.Client {
	return r.client
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

func (r *ProviderConfig) Initialize(ctx context.Context, data *cmlschema.ProviderModel, diag diag.Diagnostics) *ProviderConfig {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client != nil {
		return r
	}

	// check if provided auth configuration makes sense
	if data.Token.IsNull() &&
		(data.Username.IsNull() || data.Password.IsNull()) {
		diag.AddError(
			"Required configuration missing",
			"null check: either username and password or a token must be provided",
		)
	}

	if len(data.Token.ValueString()) == 0 &&
		(len(data.Username.ValueString()) == 0 || len(data.Password.ValueString()) == 0) {
		diag.AddError(
			"Required configuration missing",
			"value check: either username and password or a token must be provided",
		)
	}

	if len(data.Token.ValueString()) > 0 && len(data.Username.ValueString()) > 0 {
		diag.AddWarning(
			"Conflicting configuration",
			"both token and username / password were provided")
	}

	// an address must be specified
	if len(data.Address.ValueString()) == 0 {
		diag.AddError(
			"Required configuration missing",
			"A server address must be configured to use th CML2 provider",
		)
	}
	if data.SkipVerify.IsNull() {
		tflog.Warn(ctx, "unspecified certificate verification, will verify")
		data.SkipVerify = types.BoolValue(false)
	}

	if data.UseCache.IsNull() {
		data.UseCache = types.BoolValue(false)
	} else if data.UseCache.ValueBool() {
		diag.AddWarning(
			"Experimental feature enabled",
			"\"use_cache\" is considered experimental and may not work as expected; use with care",
		)
	}

	// create a new CML2 client
	client := cmlclient.New(
		data.Address.ValueString(),
		data.SkipVerify.ValueBool(),
		data.UseCache.ValueBool(),
	)
	if len(data.Username.ValueString()) > 0 {
		client.SetUsernamePassword(
			data.Username.ValueString(),
			data.Password.ValueString(),
		)
	}
	if len(data.Token.ValueString()) > 0 {
		client.SetToken(data.Token.ValueString())
	}

	if len(data.CAcert.ValueString()) > 0 {
		err := client.SetCACert([]byte(data.CAcert.ValueString()))
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
	return config.Initialize(ctx, config.data, resp.Diagnostics)
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
	return config.Initialize(ctx, config.data, resp.Diagnostics)
}
