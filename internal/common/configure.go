// Package common provides functions and types used by several other packages
// of the CML2 Terraform provider.
package common

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient/pkg/client"
	"github.com/rschmied/gocmlclient/pkg/models"
)

type ProviderConfig struct {
	client *cmlclient.Client
	data   *cmlschema.ProviderModel
	mu     *sync.Mutex

	// nodeDefs caches node definitions for plan-time heuristics.
	// It is loaded lazily on first use.
	nodeDefs       models.NodeDefinitionMap
	nodeDefsLoaded bool
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
		client:         nil,
		mu:             new(sync.Mutex),
		data:           data,
		nodeDefs:       nil,
		nodeDefsLoaded: false,
	}
}

// NodeDefinitions returns the controller's node definition map.
// The result is cached for the lifetime of the provider instance.
func (r *ProviderConfig) NodeDefinitions(ctx context.Context) (models.NodeDefinitionMap, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodeDefsLoaded {
		return r.nodeDefs, nil
	}
	if r.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	defs, err := r.client.NodeDefinition.NodeDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	r.nodeDefs = defs
	r.nodeDefsLoaded = true
	return r.nodeDefs, nil
}

func (r *ProviderConfig) Initialize(ctx context.Context, diags *diag.Diagnostics) *ProviderConfig {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client != nil {
		return r
	}

	// check if provided auth configuration makes sense
	if r.data.Token.IsNull() &&
		(r.data.Username.IsNull() || r.data.Password.IsNull()) {
		diags.AddError(
			"Required configuration missing",
			"null check: either username and password or a token must be provided",
		)
	}

	if len(r.data.Token.ValueString()) == 0 &&
		(len(r.data.Username.ValueString()) == 0 || len(r.data.Password.ValueString()) == 0) {
		diags.AddError(
			"Required configuration missing",
			"value check: either username and password or a token must be provided",
		)
	}

	if len(r.data.Token.ValueString()) > 0 && len(r.data.Username.ValueString()) > 0 {
		diags.AddWarning(
			"Conflicting configuration",
			"both token and username / password were provided")
	}

	// an address must be specified
	if len(r.data.Address.ValueString()) == 0 {
		diags.AddError(
			"Required configuration missing",
			"A server address must be configured to use the CML2 provider",
		)
	}

	// address must be https
	parsedURL, err := url.Parse(r.data.Address.ValueString())
	if err != nil {
		diags.AddError(
			"Can't parse server address / URL",
			err.Error(),
		)
	}

	// Check if the scheme is HTTPS and we have something like a hostname
	if parsedURL.Scheme != "https" || len(parsedURL.Host) == 0 {
		diags.AddError(
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
		diags.AddWarning(
			"Feature",
			"\"named_configs\" is enabled",
		)
	}

	if r.data.UseCache.IsNull() {
		r.data.UseCache = types.BoolValue(false)
	} else if r.data.UseCache.ValueBool() {
		diags.AddError(
			"Experimental feature deprecated",
			"\"use_cache\" has been deprecated",
		)
	}

	// build client options
	opts := make([]cmlclient.Option, 0)

	// Policy: do not readiness-check at init (see spec/02-readiness-behavior.md)
	opts = append(opts, cmlclient.SkipReadyCheck())

	// Policy: named configs default OFF unless explicitly enabled
	if !r.data.NamedConfigs.ValueBool() {
		opts = append(opts, cmlclient.WithoutNamedConfigs())
	}

	// Policy: always request node configurations explicitly to avoid server-default
	// drift (CML 2.9 returns string when unset; CML 2.10 returns named-config list).
	// This is independent from the user-facing named_configs setting.
	opts = append(opts, cmlclient.WithNodeExcludeConfigurations(false))

	// Auth
	if len(r.data.Token.ValueString()) > 0 {
		opts = append(opts, cmlclient.WithToken(r.data.Token.ValueString()))
	}
	if len(r.data.Username.ValueString()) > 0 {
		opts = append(opts, cmlclient.WithUsernamePassword(
			r.data.Username.ValueString(),
			r.data.Password.ValueString(),
		))
	}

	// Optional token caching (username/password only). This is intentionally
	// ignored when a token is explicitly configured.
	if r.data.TokenCache.IsNull() {
		r.data.TokenCache = types.BoolValue(false)
	}
	if r.data.TokenCacheFile.IsNull() {
		r.data.TokenCacheFile = types.StringNull()
	}
	if r.data.TokenCache.ValueBool() && len(r.data.Token.ValueString()) == 0 && len(r.data.Username.ValueString()) > 0 {
		cacheFile := r.data.TokenCacheFile.ValueString()
		if len(cacheFile) == 0 {
			hostKey := parsedURL.Host
			hostKey = strings.Map(func(r rune) rune {
				switch {
				case r >= 'a' && r <= 'z':
					return r
				case r >= 'A' && r <= 'Z':
					return r
				case r >= '0' && r <= '9':
					return r
				default:
					return '_'
				}
			}, hostKey)
			cacheFile = fmt.Sprintf("/tmp/terraform-provider-cml2-token-%s.json", hostKey)
		}
		opts = append(opts, cmlclient.WithTokenStorageFile(cacheFile))
	}

	// HTTP/TLS
	if r.data.SkipVerify.ValueBool() {
		opts = append(opts, cmlclient.WithInsecureTLS())
	}
	if len(r.data.CAcert.ValueString()) > 0 {
		opts = append(opts, cmlclient.WithCACertPEM([]byte(r.data.CAcert.ValueString())))
	}

	client, err := cmlclient.New(r.data.Address.ValueString(), opts...)
	if err != nil {
		diags.AddError(
			"CML client initialization failed",
			err.Error(),
		)
		return r
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
	return config.Initialize(ctx, &resp.Diagnostics)
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
	return config.Initialize(ctx, &resp.Diagnostics)
}
