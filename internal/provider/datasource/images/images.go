package images

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
var _ datasource.DataSource = &ImagesDataSource{}

type ImagesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	NodeDef   types.String `tfsdk:"nodedefinition"`
	ImageList types.List   `tfsdk:"image_list"`
}

func NewDataSource() datasource.DataSource {
	return &ImagesDataSource{}
}

// ImagesDataSource defines the data source implementation.
type ImagesDataSource struct {
	cfg *common.ProviderConfig
}

// type ImageDefinition struct {
// 	ID            string `json:"id"`
// 	SchemaVersion string `json:"schema_version"`
// 	NodeDefID     string `json:"node_definition_id"`
// 	Description   string `json:"description"`
// 	Label         string `json:"label"`
// 	DiskImage1    string `json:"disk_image"`
// 	DiskImage2    string `json:"disk_image_2"`
// 	DiskImage3    string `json:"disk_image_3"`
// 	ReadOnly      bool   `json:"read_only"`
// 	DiskSubfolder string `json:"disk_subfolder"`
// 	RAM           *int   `json:"ram"`
// 	CPUs          *int   `json:"cpus"`
// 	CPUlimit      *int   `json:"cpu_limit"`
// 	DataVolume    *int   `json:"data_volume"`
// 	BootDiskSize  *int   `json:"boot_disk_size"`
// }

// [
//   {
//     "id": "alpine-3-13-2-base",
//     "node_definition_id": "alpine",
//     "description": "Alpine Linux and network tools",
//     "label": "Alpine 3.13.2",
//     "disk_image": "alpine-3-13-2-base.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "alpine-3-13-2-base",
//     "schema_version": "0.0.1"
//   },

func (d *ImagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *ImagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *ImagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.",
			Computed:    true,
		},
		"nodedefinition": schema.StringAttribute{
			Description: "A node definition ID to filter the image list.",
			Optional:    true,
		},
		"image_list": schema.ListNestedAttribute{
			MarkdownDescription: "A list of all image definitions available on the controller, potentially filtered by the provided `nodedefinition` attribute.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: cmlschema.ImageDef(),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = "A data source that retrieves image definitions from the controller. The optional `nodedefinition` ID can be provided to filter the list of image definitions for the specified node definition. If no node definition ID is provided, the complete image definition list known to the controller is returned."
	resp.Diagnostics = nil
}

func (d *ImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImagesDataSourceModel

	tflog.Info(ctx, "Datasource Images READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	images, err := d.cfg.Client().GetImageDefs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get image definitions, got error: %s", err),
		)
		return
	}

	imageList := make([]attr.Value, 0)
	for _, img := range images {
		if !data.NodeDef.IsNull() && img.NodeDefID != data.NodeDef.ValueString() {
			continue
		}
		imageList = append(imageList, cmlschema.NewImageDefinition(
			ctx, &img, &resp.Diagnostics),
		)

	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			imageList,
			types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: cmlschema.ImageDefAttrType,
				},
			},
			&data.ImageList,
		)...,
	)
	// need an ID
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(uuid.New().String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource Images READ: done")
}
