package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	cmlclient "github.com/rschmied/gocmlclient"
)

type ProviderConfig struct {
	client *cmlclient.Client
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

func NewProviderConfig(client *cmlclient.Client) *ProviderConfig {
	return &ProviderConfig{
		client: client,
		mu:     new(sync.Mutex),
	}
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
	return config
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
	return config
}
