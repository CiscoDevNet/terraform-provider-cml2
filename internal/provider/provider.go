package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/pkg/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &CML2Provider{}
var _ provider.ProviderWithMetadata = &CML2Provider{}

// CML2Provider defines the Cisco Modeling Labs Terraform provider implementation.
type CML2Provider struct {
	version string
}

// CML2ProviderModel describes the provider data model.
type CML2ProviderModel struct {
	Address    types.String `tfsdk:"address"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Token      types.String `tfsdk:"token"`
	CAcert     types.String `tfsdk:"cacert"`
	SkipVerify types.Bool   `tfsdk:"skip_verify"`
}

func (p *CML2Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cml2"
	resp.Version = p.version
}

func (p *CML2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CML2ProviderModel

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

	// create a new CML2 client
	client := cmlclient.NewClient(
		data.Address.Value,
		data.SkipVerify.Value,
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
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CML2Provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"address": {
				Description: "CML2 controller address",
				Required:    true,
				Type:        types.StringType,
			},
			"username": {
				Description: "CML2 username",
				Optional:    true,
				Type:        types.StringType,
			},
			"password": {
				Description: "CML2 password",
				Optional:    true,
				Type:        types.StringType,
				Sensitive:   true,
			},
			"token": {
				Description: "CML2 API token (JWT)",
				Optional:    true,
				Type:        types.StringType,
				Sensitive:   true,
			},
			"cacert": {
				Description: "CA CERT, PEM encoded",
				Optional:    true,
				Type:        types.StringType,
			},
			"skip_verify": {
				Description: "disable TLS certificate verification",
				Optional:    true,
				Type:        types.BoolType,
			},
		},
	}, nil
}

func (p *CML2Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLabResource,
	}
}

func (p *CML2Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNodeDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CML2Provider{
			version: version,
		}
	}
}
