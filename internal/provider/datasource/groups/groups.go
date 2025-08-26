// Package groups implements the CML2 groups datasource.
package groups

import (
	"context"
	"fmt"

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
var _ datasource.DataSource = &GroupDataSource{}

type GroupDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Groups types.List   `tfsdk:"groups"`
}

func NewDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// GroupDataSource defines the data source implementation.
type GroupDataSource struct {
	cfg *common.ProviderConfig
}

func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "A group name to filter the groups list returned by the controller. Group names must be unique, so it's either one group or no group at all if a name filter is provided.",
			Optional:    true,
		},
		"groups": schema.ListNestedAttribute{
			MarkdownDescription: "A list of all permission groups available on the controller.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: cmlschema.Converter(cmlschema.Group()),
			},
			Computed: true,
		},
	}
	resp.Schema.MarkdownDescription = "A data source that retrieves a list of permission group information from the controller."
	resp.Diagnostics = nil
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	tflog.Info(ctx, "Datasource Group READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.cfg.Client().Groups(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get groups, got error: %s", err),
		)
		return
	}

	groupList := make([]attr.Value, 0)
	for _, group := range groups {
		// filter by group name
		if !data.Name.IsNull() && group.Name != data.Name.ValueString() {
			continue
		}
		groupList = append(groupList, cmlschema.NewGroup(
			ctx, group, &resp.Diagnostics),
		)
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			groupList,
			types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: cmlschema.GroupAttrType,
				},
			},
			&data.Groups,
		)...,
	)
	// need an ID
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(uuid.New().String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource System READ: done")
}
