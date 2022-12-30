package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

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

func (d *LabDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Lab ID that identifies the lab",
			Optional:    true,
		},
		"title": schema.StringAttribute{
			Description: "Lab title. If not unique, it will return the first one that matches. Use ID for labs with non-unique titles.",
			Optional:    true,
		},
		"lab": schema.SingleNestedAttribute{
			Description: "lab data",
			Attributes:  cmlschema.Converter(cmlschema.Lab()),
			Computed:    true,
		},
	}
	resp.Schema.MarkdownDescription = "A lab data source. Either the lab `id` or the lab `title` must be provided to retrieve the `lab` data from the controller."
	resp.Diagnostics = nil
}

func (d *LabDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data LabDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.ID.IsNull() && data.Title.IsNull() {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
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
	if data.ID.IsNull() {
		lab, err = d.cfg.Client().LabGetByTitle(ctx, data.Title.ValueString(), false)
	} else {
		lab, err = d.cfg.Client().LabGet(ctx, data.ID.ValueString(), false)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewLab(ctx, lab, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.LabAttrType},
			&data.Lab,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Datasource Lab READ: done")
}
