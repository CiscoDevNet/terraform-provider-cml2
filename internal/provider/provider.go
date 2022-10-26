package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	d_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/lab"
	d_node "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/node"
	r_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lab"
	r_lifecycle "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lifecycle"
	r_link "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/link"
	r_node "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/node"

	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &CML2Provider{}
var _ provider.ProviderWithMetadata = &CML2Provider{}

const CML2ErrorLabel = "CML2 Provider Error"

// CML2Provider defines the Cisco Modeling Labs Terraform provider implementation.
type CML2Provider struct {
	version string
}

func (p *CML2Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cml2"
	resp.Version = p.version
}

func (p *CML2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data schema.ProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if provided auth configuration makes sense
	if data.Token.Null &&
		(data.Username.Null || data.Password.Null) {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("null check: either username and password or a token must be provided %T", p),
		)
	}

	if len(data.Token.Value) == 0 &&
		(len(data.Username.Value) == 0 || len(data.Password.Value) == 0) {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("value check: either username and password or a token must be provided %T", p),
		)
	}

	if len(data.Token.Value) > 0 && len(data.Username.Value) > 0 {
		resp.Diagnostics.AddWarning(
			"Conflicting configuration",
			"both token and username / password were provided")
	}

	// an address must be specified
	if len(data.Address.Value) == 0 {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("a server address must be configured to use %T", p),
		)
	}
	if data.SkipVerify.Null {
		tflog.Warn(ctx, "unspecified certificate verification, will verify")
		data.SkipVerify.Value = false
	}

	if data.UseCache.Null {
		data.SkipVerify.Value = false
	}
	if data.UseCache.Value {
		resp.Diagnostics.AddWarning(
			"Experimental feature enabled",
			"\"use_cache\" is considered experimental and may not work as expected; use with care",
		)
	}

	// create a new CML2 client
	client := cmlclient.NewClient(
		data.Address.Value,
		data.SkipVerify.Value,
		data.UseCache.Value,
	)
	if len(data.Username.Value) > 0 {
		client.SetUsernamePassword(data.Username.Value, data.Password.Value)
	}
	if len(data.Token.Value) > 0 {
		client.SetToken(data.Token.Value)
	}

	if len(data.CAcert.Value) > 0 {
		err := client.SetCACert([]byte(data.CAcert.Value))
		if err != nil {
			resp.Diagnostics.AddError(
				"Configuration issue",
				fmt.Sprintf("Provided certificate could not be used: %s", err),
			)
		}
	}
	config := common.NewProviderConfig(client)
	resp.DataSourceData = config
	resp.ResourceData = config
}

func (p *CML2Provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: schema.Provider(),
	}, nil
}

func (p *CML2Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		r_lab.NewResource,
		r_lifecycle.NewResource,
		r_link.NewResource,
		r_node.NewResource,
	}
}

func (p *CML2Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		d_lab.NewDataSource,
		d_node.NewDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CML2Provider{
			version: version,
		}
	}
}
