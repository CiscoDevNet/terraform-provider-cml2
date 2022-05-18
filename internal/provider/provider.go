package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.Provider = &cml2{}

// cml2 satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type cml2 struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	//
	client *cmlclient.Client

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Address    types.String `tfsdk:"address"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Token      types.String `tfsdk:"token"`
	CAcert     types.String `tfsdk:"cacert"`
	SkipVerify types.Bool   `tfsdk:"skip_verify"`
}

func (p *cml2) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

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
			"Both auth options provided",
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
	p.client = cmlclient.NewClient(
		data.Address.Value,
		data.SkipVerify.Value,
	)
	if len(data.Username.Value) > 0 {
		p.client.SetUsernamePassword(data.Username.Value, data.Password.Value)
	}
	if len(data.Token.Value) > 0 {
		p.client.SetToken(data.Token.Value)
	}

	if len(data.CAcert.Value) > 0 {
		err := p.client.SetCACert([]byte(data.CAcert.Value))
		if err != nil {
			resp.Diagnostics.AddError(
				"Configuration issue",
				fmt.Sprintf("Provided certificate could not be used: %s", err),
			)
		}
	}
	p.configured = true
}

func (p *cml2) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"cml2_lab": cml2LabResourceType{},
		// "cml2_node": cmlNodeResourceType{},
		// "cml2_link":      cmlLinkResourceType{},
		// "cml2_interface": cmlInterfaceResourceType{},
	}, nil
}

func (p *cml2) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		// "cml2_lab_details": cmlLabDetailDataSourceType{},
	}, nil
}

func (p *cml2) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"address": {
				MarkdownDescription: "CML2 controller address",
				Required:            true,
				Type:                types.StringType,
			},
			"username": {
				MarkdownDescription: "CML2 username",
				Optional:            true,
				Type:                types.StringType,
			},
			"password": {
				MarkdownDescription: "CML2 password",
				Optional:            true,
				Type:                types.StringType,
				Sensitive:           true,
			},
			"token": {
				MarkdownDescription: "CML2 API token (JWT)",
				Optional:            true,
				Type:                types.StringType,
				Sensitive:           true,
			},
			"cacert": {
				MarkdownDescription: "CA CERT, PEM encoded",
				Optional:            true,
				Type:                types.StringType,
			},
			"skip_verify": {
				MarkdownDescription: "Disable TLS certificate verification",
				Optional:            true,
				Type:                types.BoolType,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &cml2{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (cml2, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*cml2)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return cml2{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return cml2{}, diags
	}

	return *p, diags
}
