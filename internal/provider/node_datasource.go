package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &NodeDataSource{}

func NewNodeDataSource() datasource.DataSource {
	return &NodeDataSource{}
}

// ExampleDataSource defines the data source implementation.
type NodeDataSource struct {
	client *cmlclient.Client
}

// ExampleDataSourceModel describes the data source data model.
type NodeDataSourceModel struct {
	LabID  types.String `tfsdk:"lab_id"`
	NodeID types.String `tfsdk:"node_id"`
	Node   types.Object `tfsdk:"node"`
}

func (d *NodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (d *NodeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "node data source",

		Attributes: map[string]tfsdk.Attribute{
			"lab_id": {
				MarkdownDescription: "lab id",
				Required:            true,
				Type:                types.StringType,
			},
			"node_id": {
				MarkdownDescription: "node id",
				Required:            true,
				Type:                types.StringType,
			},
			"node": {
				MarkdownDescription: "node data",
				Attributes: tfsdk.SingleNestedAttributes(
					nodeSchema(),
				),
				Computed: true,
			},
		},
	}, nil
}

func (d *NodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := d.client.GetLab(ctx, data.LabID.Value, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	node, found := lab.Nodes[data.NodeID.Value]
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
			newNode(ctx, node, resp.Diagnostics),
			types.ObjectType{AttrTypes: nodeAttrType},
			&data.Node,
		)...,
	)

	tflog.Info(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
