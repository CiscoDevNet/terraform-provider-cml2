package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

// ImageDefModel is the TF representation of a CML2 image definition
// it does not contain all attributes as defined by the API endpoint
// the ones ommitted are irrelevant for TF operations (e.g. disk paths)
type ImageDefinitionModel struct {
	ID            types.String `tfsdk:"id"`
	NodeDefID     types.String `tfsdk:"node_definition_id"`
	Description   types.String `tfsdk:"description"`
	Label         types.String `tfsdk:"label"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	RAM           types.Int64  `tfsdk:"ram"`
	CPUs          types.Int64  `tfsdk:"cpus"`
	CPUlimit      types.Int64  `tfsdk:"cpu_limit"`
	DataVolume    types.Int64  `tfsdk:"data_volume"`
	BootDiskSize  types.Int64  `tfsdk:"boot_disk_size"`
	SchemaVersion types.String `tfsdk:"schema_version"`
}

var ImageDefAttrType = map[string]attr.Type{
	"id":                 types.StringType,
	"node_definition_id": types.StringType,
	"description":        types.StringType,
	"label":              types.StringType,
	"read_only":          types.BoolType,
	"ram":                types.Int64Type,
	"cpus":               types.Int64Type,
	"cpu_limit":          types.Int64Type,
	"data_volume":        types.Int64Type,
	"boot_disk_size":     types.Int64Type,
	"schema_version":     types.StringType,
}

// ImageDef returns the schema for the Image definition model
func ImageDef() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID to identifying the image",
			Computed:    true,
		},
		"node_definition_id": schema.StringAttribute{
			Description: "ID of the node definition this image belongs to",
			Optional:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description of this image definition",
			Computed:    true,
		},
		"label": schema.StringAttribute{
			Description: "Text label of this image definition",
			Computed:    true,
		},
		"read_only": schema.BoolAttribute{
			Description: "Is this image definition read only?",
			Computed:    true,
		},
		"ram": schema.Int64Attribute{
			Description: "Image specific RAM value, can be null",
			Computed:    true,
		},
		"cpus": schema.Int64Attribute{
			Description: "Image specific amount of CPUs, can be null",
			Computed:    true,
		},
		"cpu_limit": schema.Int64Attribute{
			Description: "Image specific CPU limit, can be null",
			Computed:    true,
		},
		"boot_disk_size": schema.Int64Attribute{
			Description: "Image specific boot disk size, can be null",
			Computed:    true,
		},
		"data_volume": schema.Int64Attribute{
			Description: "Image specific data volume size, can be null",
			Computed:    true,
		},
		"schema_version": schema.StringAttribute{
			Description: "Version of the image definition schemage",
			Computed:    true,
		},
	}
}

func NewImageDefinition(ctx context.Context, image *cmlclient.ImageDefinition, diags *diag.Diagnostics) attr.Value {

	newImage := ImageDefinitionModel{
		ID:            types.StringValue(image.ID),
		NodeDefID:     types.StringValue(image.NodeDefID),
		Label:         types.StringValue(image.Label),
		Description:   types.StringValue(image.Description),
		ReadOnly:      types.BoolValue(image.ReadOnly),
		SchemaVersion: types.StringValue(image.SchemaVersion),

		RAM:          types.Int64Null(),
		BootDiskSize: types.Int64Null(),
		DataVolume:   types.Int64Null(),
		CPUs:         types.Int64Null(),
		CPUlimit:     types.Int64Null(),
	}

	if image.RAM != nil {
		newImage.RAM = types.Int64Value(int64(*image.RAM))
	}
	if image.BootDiskSize != nil {
		newImage.BootDiskSize = types.Int64Value(int64(*image.BootDiskSize))
	}
	if image.DataVolume != nil {
		newImage.DataVolume = types.Int64Value(int64(*image.DataVolume))
	}
	if image.CPUs != nil {
		newImage.CPUs = types.Int64Value(int64(*image.CPUs))
	}
	if image.CPUlimit != nil {
		newImage.CPUlimit = types.Int64Value(int64(*image.CPUlimit))
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newImage,
			types.ObjectType{AttrTypes: ImageDefAttrType},
			&value,
		)...,
	)
	return value
}
