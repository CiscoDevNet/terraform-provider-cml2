package node

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &NodeDataSource{}

func NewDataSource() datasource.DataSource {
	return &NodeDataSource{}
}

// ExampleDataSource defines the data source implementation.
type NodeDataSource struct {
	cfg *common.ProviderConfig
}

// ExampleDataSourceModel describes the data source data model.
type NodeDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	LabID types.String `tfsdk:"lab_id"`
	Node  types.Object `tfsdk:"node"`
}

func (d *NodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (d *NodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *NodeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A CML2 node data source.  Both, the node `id` and the `lab_id` must be provided to retrieve the `node` data from the controller.  Note that **all** of the attributes of the node element are read-only even though the auto-generated schema documentation lists some of them as \"optional\".",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "node id",
				Required:    true,
				Type:        types.StringType,
			},
			"lab_id": {
				Description: "lab id",
				Required:    true,
				Type:        types.StringType,
			},
			"node": {
				Description: "node data",
				Attributes: tfsdk.SingleNestedAttributes(
					schema.Node(),
				),
				Computed: true,
			},
		},
	}, nil
}

func (d *NodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDataSourceModel

	tflog.Info(ctx, "Datasource Node READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := d.cfg.Client().LabGet(ctx, data.LabID.Value, true) // deep!
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	node, found := lab.Nodes[data.ID.Value]
	if !found {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("requested node %s not found", data.Node),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			schema.NewNode(ctx, node, &resp.Diagnostics),
			types.ObjectType{AttrTypes: schema.NodeAttrType},
			&data.Node,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource Node READ: done")
}
