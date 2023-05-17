package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	d_groups "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/groups"
	d_images "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/images"
	d_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/lab"
	d_node "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/node"
	d_system "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/system"
	d_users "github.com/rschmied/terraform-provider-cml2/internal/provider/datasource/users"
	r_group "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/group"
	r_lab "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lab"
	r_lifecycle "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/lifecycle"
	r_link "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/link"
	r_node "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/node"
	r_user "github.com/rschmied/terraform-provider-cml2/internal/provider/resource/user"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &CML2Provider{}

// CML2Provider defines the Cisco Modeling Labs Terraform provider implementation.
type CML2Provider struct {
	version string
	name    string
}

func (p *CML2Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = p.name
	resp.Version = p.version
}

func (p *CML2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data cmlschema.ProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// https://dev.to/camptocamp-ops/how-to-allow-dynamic-terraform-provider-configuration-20ik
	dynamic_config := false
	if data.DynamicConfig.IsNull() {
		data.DynamicConfig = types.BoolValue(false)
	} else if data.DynamicConfig.ValueBool() {
		dynamic_config = true
		// resp.Diagnostics.AddWarning(
		// 	"Dynamic configuration",
		// 	With "\"dynamic_config\", late binding of the provider configuration is enabled",
		// )
	}

	config := common.NewProviderConfig(&data)
	if !dynamic_config {
		config.Initialize(ctx, resp.Diagnostics)
	}
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
		r_group.NewResource,
		r_user.NewResource,
	}
}

func (p *CML2Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		d_lab.NewDataSource,
		d_node.NewDataSource,
		d_images.NewDataSource,
		d_system.NewDataSource,
		d_groups.NewDataSource,
		d_users.NewDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CML2Provider{
			version: version,
			name:    "cml2",
		}
	}
}
