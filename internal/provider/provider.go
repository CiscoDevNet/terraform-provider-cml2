package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	d_images "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/images"
	d_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/lab"
	d_node "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/node"
	r_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lab"
	r_lifecycle "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lifecycle"
	r_link "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/link"
	r_node "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/node"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &CML2Provider{}

// CML2Provider defines the Cisco Modeling Labs Terraform provider implementation.
type CML2Provider struct {
	version string
}

func (p *CML2Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cml2"
	resp.Version = p.version
}

func (p *CML2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data cmlschema.ProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if provided auth configuration makes sense
	if data.Token.IsNull() &&
		(data.Username.IsNull() || data.Password.IsNull()) {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("null check: either username and password or a token must be provided %T", p),
		)
	}

	if len(data.Token.ValueString()) == 0 &&
		(len(data.Username.ValueString()) == 0 || len(data.Password.ValueString()) == 0) {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("value check: either username and password or a token must be provided %T", p),
		)
	}

	if len(data.Token.ValueString()) > 0 && len(data.Username.ValueString()) > 0 {
		resp.Diagnostics.AddWarning(
			"Conflicting configuration",
			"both token and username / password were provided")
	}

	// an address must be specified
	if len(data.Address.ValueString()) == 0 {
		resp.Diagnostics.AddError(
			"Required configuration missing",
			fmt.Sprintf("a server address must be configured to use %T", p),
		)
	}
	if data.SkipVerify.IsNull() {
		tflog.Warn(ctx, "unspecified certificate verification, will verify")
		data.SkipVerify = types.BoolValue(false)
	}

	if data.UseCache.IsNull() {
		data.UseCache = types.BoolValue(false)
	} else if data.UseCache.ValueBool() {
		resp.Diagnostics.AddWarning(
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

func (p *CML2Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema.MarkdownDescription = `The CML2 Terraform provider helps to
deploy and run entire "virtual networks as code" into the Cisco Modeling Labs network
simulation platform. Available deployment methods allow to create networks (e.g.,
routers, switches and endpoints and their connectivity) as well as import existing CML2
topologies. It also includes fine-grained lifecycle control (staged start up),
configuration injection, IP address retrieval from network devices, and more.`
	resp.Schema.Attributes = cmlschema.Provider()
	resp.Diagnostics = nil
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
		d_images.NewDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CML2Provider{
			version: version,
		}
	}
}
