// Package extconn implements the CML2 extconn datasource.
package extconn

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ConnectorDataSource{}

type ConnectorDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Label      types.String `tfsdk:"label"`
	Tag        types.String `tfsdk:"tag"`
	Connectors types.List   `tfsdk:"connectors"`
}

func NewDataSource() datasource.DataSource {
	return &ConnectorDataSource{}
}

// ConnectorDataSource defines the data source implementation.
type ConnectorDataSource struct {
	cfg *common.ProviderConfig
}

func (d *ConnectorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (d *ConnectorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *ConnectorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.",
			Computed:    true,
		},
		"label": schema.StringAttribute{
			Description: "A connector label to filter the connector list returned by the controller. Connector labels must be unique, so it's either one group or no group at all if a name filter is provided.",
			Optional:    true,
		},
		"tag": schema.StringAttribute{
			Description: "A tag name to filter the groups list returned by the controller. Connector tags can be defined on multiple connectors, a list can be returned.",
			Optional:    true,
		},
		"connectors": schema.ListNestedAttribute{
			MarkdownDescription: "A list of all permission groups available on the controller.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: cmlschema.Converter(cmlschema.ExtConn()),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = "A data source that retrieves external connectors information from the controller."
	resp.Diagnostics = nil
}

func (d *ConnectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectorDataSourceModel

	tflog.Info(ctx, "Datasource Connectors READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectors, err := d.cfg.Client().ExtConnectors(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get connectors, got error: %s", err),
		)
		return
	}

	result := make([]attr.Value, 0)
	for _, connector := range connectors {
		if !data.Label.IsNull() && connector.Label != data.Label.ValueString() {
			continue
		}
		if !data.Tag.IsNull() {
			found := slices.Contains(connector.Tags, data.Tag.ValueString())
			if !found {
				continue
			}
		}
		result = append(result, cmlschema.NewExtConn(
			ctx, connector, &resp.Diagnostics),
		)

	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			result,
			types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: cmlschema.ExtConnAttrType,
				},
			},
			&data.Connectors,
		)...,
	)
	// need an ID
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(uuid.New().String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource Connectors READ: done")
}
