package users

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

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &UsersDataSource{}

type UsersDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Users    types.List   `tfsdk:"users"`
}

func NewDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	cfg *common.ProviderConfig
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.",
			Computed:    true,
		},
		"username": schema.StringAttribute{
			Description: "A user name to filter the users list returned by the controller. User names must be unique, so it's either one user or no user at all if a name filter is provided.",
			Optional:    true,
		},
		"users": schema.ListNestedAttribute{
			MarkdownDescription: "A list of all users available on the controller.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: cmlschema.Converter(cmlschema.User()),
			},
			Computed: true,
		},
	}
	resp.Schema.MarkdownDescription = "A data source that retrieves a list of users from the controller."
	resp.Diagnostics = nil
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	tflog.Info(ctx, "Datasource Users READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.cfg.Client().Users(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get users, got error: %s", err),
		)
		return
	}

	userList := make([]attr.Value, 0)
	for _, user := range users {
		// filter by username
		if !data.Username.IsNull() && user.Username != data.Username.ValueString() {
			continue
		}
		userList = append(userList, cmlschema.NewUser(
			ctx, user, &resp.Diagnostics),
		)
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			userList,
			types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: cmlschema.UserAttrType,
				},
			},
			&data.Users,
		)...,
	)
	// need an ID
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(uuid.New().String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource Users READ: done")
}
