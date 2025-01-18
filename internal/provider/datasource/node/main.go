package node

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

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

func (d *NodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Node ID to identify the node",
			Required:    true,
		},
		"lab_id": schema.StringAttribute{
			Description: "Lab ID to identify the lab that contains the node",
			Required:    true,
		},
		"node": schema.SingleNestedAttribute{
			Description: "node data",
			Attributes:  cmlschema.Converter(cmlschema.Node()),
			Computed:    true,
		},
	}
	resp.Schema.MarkdownDescription = "A node data source.  Both, the node `id` and the `lab_id` must be provided to retrieve the `node` data from the controller."
	resp.Diagnostics = nil
}

func (d *NodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDataSourceModel

	tflog.Info(ctx, "Datasource Node READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := d.cfg.Client().LabGet(ctx, data.LabID.ValueString(), true) // deep!
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	node, found := lab.Nodes[data.ID.ValueString()]
	if !found {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("requested node %s not found", data.Node),
		)
		return
	}

	data.ID = types.StringValue(node.ID)
	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, node, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data.Node,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource Node READ: done")
}
