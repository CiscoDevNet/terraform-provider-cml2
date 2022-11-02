package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &LabDataSource{}
var _ datasource.DataSourceWithValidateConfig = &LabDataSource{}

type LabDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Title types.String `tfsdk:"title"`
	Lab   types.Object `tfsdk:"lab"`
}

func NewDataSource() datasource.DataSource {
	return &LabDataSource{}
}

// ExampleDataSource defines the data source implementation.
type LabDataSource struct {
	cfg *common.ProviderConfig
}

func (d *LabDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

func (d *LabDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *LabDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the
		// language server.
		MarkdownDescription: "A CML2 lab data source. Either the lab `id` or the lab `title` must be provided to retrieve the `lab` data from the controller.  Note that **all** of the attributes of the lab element are read-only even though the auto-generated schema documentation lists some of them as \"optional\".",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "lab id",
				Optional:    true,
				Type:        types.StringType,
			},
			"title": {
				Description: "lab title",
				Optional:    true,
				Type:        types.StringType,
			},
			"lab": {
				Description: "lab data",
				Attributes: tfsdk.SingleNestedAttributes(
					schema.Lab(),
				),
				Computed: true,
			},
		},
	}, nil
}

func (d *LabDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {

	var data LabDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.ID.Null && data.Title.Null {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			"need to provide either title to search for or a lab ID",
		)
		return
	}
}

func (d *LabDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LabDataSourceModel

	tflog.Info(ctx, "Datasource Lab READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		lab *cmlclient.Lab
		err error
	)
	if data.ID.Null {
		lab, err = d.cfg.Client().LabGetByTitle(ctx, data.Title.Value, false)
	} else {
		lab, err = d.cfg.Client().LabGet(ctx, data.ID.Value, false)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			schema.NewLab(ctx, lab, &resp.Diagnostics),
			types.ObjectType{AttrTypes: schema.LabAttrType},
			&data.Lab,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Datasource Lab READ: done")
}
