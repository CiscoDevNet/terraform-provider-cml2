package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

type NodeModel struct {
	ID              types.String `tfsdk:"id"`
	LabID           types.String `tfsdk:"lab_id"`
	Label           types.String `tfsdk:"label"`
	State           types.String `tfsdk:"state"`
	NodeDefinition  types.String `tfsdk:"nodedefinition"`
	ImageDefinition types.String `tfsdk:"imagedefinition"`
	Configuration   Config       `tfsdk:"configuration"`
	Interfaces      types.List   `tfsdk:"interfaces"`
	Tags            types.Set    `tfsdk:"tags"`
	X               types.Int64  `tfsdk:"x"`
	Y               types.Int64  `tfsdk:"y"`
	HideLinks       types.Bool   `tfsdk:"hide_links"`
	CPUs            types.Int64  `tfsdk:"cpus"`
	CPUlimit        types.Int64  `tfsdk:"cpu_limit"`
	RAM             types.Int64  `tfsdk:"ram"`
	BootDiskSize    types.Int64  `tfsdk:"boot_disk_size"`
	DataVolume      types.Int64  `tfsdk:"data_volume"`
	VNCkey          types.String `tfsdk:"vnc_key"`
	SerialDevices   types.List   `tfsdk:"serial_devices"`
	ComputeID       types.String `tfsdk:"compute_id"`
}

type serialDeviceModel struct {
	ConsoleKey   types.String `tfsdk:"console_key"`
	DeviceNumber types.Int64  `tfsdk:"device_number"`
}

// with simplified=true
// {
// 	"id": "a3f93420-69d5-4af8-b358-3ef93a97c763",
// 	"label": "server-0",
// 	"x": 431,
// 	"y": 308,
// 	"node_definition": "server",
// 	"image_definition": null,
// 	"state": "BOOTED",
// 	"cpus": null,
// 	"cpu_limit": null,
// 	"ram": null,
// 	"data_volume": null,
// 	"boot_disk_size": null
// }

// with simplified=false
// {
// 	"boot_disk_size": 16,
// 	"compute_id": "9c2519bf-dda6-4d31-942e-8068a6349b5e",
// 	"configuration": "# this is a shell script which will be sourced at boot\nhostname inserthostname_here\n# configurable user account\nUSERNAME=cisco\nPASSWORD=cisco\n# no password for tc user by default\nTC_PASSWORD=",
// 	"cpu_limit": 100,
// 	"cpus": 1,
// 	"data_volume": 0,
// 	"hide_links": false,
// 	"id": "a3f93420-69d5-4af8-b358-3ef93a97c763",
// 	"image_definition": null,
// 	"lab_id": "1248b67f-5fe0-4913-9c46-fbe044abc297",
// 	"label": "server-0",
// 	"node_definition": "server",
// 	"ram": 128,
// 	"tags": [],
// 	"vnc_key": "24c5a70c-1809-4360-9bf4-41e57f6a5e20",
// 	"x": 431,
// 	"y": 308,
// 	"config_filename": "cfg",
// 	"config_mediatype": "ISO",
// 	"config_image_path": "/var/local/virl2/images/1248b67f-5fe0-4913-9c46-fbe044abc297/a3f93420-69d5-4af8-b358-3ef93a97c763/config.img",
// 	"cpu_model": null,
// 	"data_image_path": "/var/local/virl2/images/1248b67f-5fe0-4913-9c46-fbe044abc297/a3f93420-69d5-4af8-b358-3ef93a97c763/data.img",
// 	"disk_image": "server-tcl-11-1/tcl-11-1.qcow2",
// 	"disk_image_2": null,
// 	"disk_image_3": null,
// 	"disk_image_path": "/var/local/virl2/images/1248b67f-5fe0-4913-9c46-fbe044abc297/a3f93420-69d5-4af8-b358-3ef93a97c763/nodedisk_0",
// 	"disk_image_path_2": null,
// 	"disk_image_path_3": null,
// 	"disk_driver": "virtio",
// 	"driver_id": "server",
// 	"efi_boot": false,
// 	"image_dir": "/var/local/virl2/images/1248b67f-5fe0-4913-9c46-fbe044abc297/a3f93420-69d5-4af8-b358-3ef93a97c763",
// 	"libvirt_image_dir": "/var/lib/libvirt/images/virl-base-images",
// 	"nic_driver": "virtio",
// 	"number_of_serial_devices": 1,
// 	"serial_devices": [
// 	  {
// 		"console_key": "f62f10aa-ca23-4500-bfe2-17fd567c7e12",
// 		"device_number": 0
// 	  }
// 	],
// 	"video_memory": 16,
// 	"video_model": null,
// 	"state": "BOOTED",
// 	"boot_progress": "Booted"
// }

var NodeAttrType = map[string]attr.Type{
	"id":              types.StringType,
	"lab_id":          types.StringType,
	"label":           types.StringType,
	"state":           types.StringType,
	"nodedefinition":  types.StringType,
	"imagedefinition": types.StringType,
	"configuration":   ConfigType{},
	"interfaces": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: InterfaceAttrType,
		},
	},
	"tags":           types.SetType{ElemType: types.StringType},
	"x":              types.Int64Type,
	"y":              types.Int64Type,
	"hide_links":     types.BoolType,
	"cpus":           types.Int64Type,
	"cpu_limit":      types.Int64Type,
	"ram":            types.Int64Type,
	"boot_disk_size": types.Int64Type,
	"data_volume":    types.Int64Type,
	"vnc_key":        types.StringType,
	"serial_devices": types.ListType{ElemType: SerialDevicesAttrType},
	"compute_id":     types.StringType,
}

var SerialDevicesAttrType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"console_key":   types.StringType,
		"device_number": types.Int64Type,
	},
}

func Node() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Node ID (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"lab_id": schema.StringAttribute{
			Description: "Lab ID containing the node (UUID).",
			Required:    true,
		},
		"label": schema.StringAttribute{
			Description: "Node label.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "Node state (`DEFINED_ON_CORE`, `STOPPED`, `STARTED`, `BOOTED`).",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"nodedefinition": schema.StringAttribute{
			Description: "Node definition / type. This can only be set at create time.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"imagedefinition": schema.StringAttribute{
			Description: "Image definition, must match the node type. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"interfaces": schema.ListNestedAttribute{
			Description: "List of interfaces on the node.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: Interface(),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"tags": schema.SetAttribute{
			Description: "Set of tags of the node.",
			Computed:    true,
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"configuration": schema.StringAttribute{
			Description: "Node configuration. Can be changed until the node is started once. Will require a replace in that case.",
			CustomType:  ConfigType{},
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"x": schema.Int64Attribute{
			Description: "X coordinate on the topology canvas.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"y": schema.Int64Attribute{
			Description: "Y coordinate on the topology canvas.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"hide_links": schema.BoolAttribute{
			Description: "If true, links are not shown in the topology. This is a visual cue and does not influence any simulation function.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"ram": schema.Int64Attribute{
			Description: "Amount of RAM, megabytes. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"cpus": schema.Int64Attribute{
			Description: "Number of CPUs. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"cpu_limit": schema.Int64Attribute{
			Description: "CPU limit in %, 20-100. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"boot_disk_size": schema.Int64Attribute{
			Description: "Size of boot disk volume, in GB. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"data_volume": schema.Int64Attribute{
			Description: "Size of data volume, in GB. Can be changed until the node is started once. Will require a replace in that case.",
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				// int64planmodifier.RequiresReplace(),
				// replace is controlled in modify_plan()
			},
		},
		"serial_devices": schema.ListAttribute{
			Description: "List of serial devices (consoles).",
			Computed:    true,
			ElementType: SerialDevicesAttrType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"vnc_key": schema.StringAttribute{
			Description: "VNC key of console, a UUID4.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"compute_id": schema.StringAttribute{
			Description: "ID of a compute this node is on, a UUID4.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func newSerialDevice(ctx context.Context, sd cmlclient.SerialDevice, diags *diag.Diagnostics) attr.Value {
	newSerialDevice := serialDeviceModel{
		ConsoleKey:   types.StringValue(sd.ConsoleKey),
		DeviceNumber: types.Int64Value(int64(sd.DeviceNumber)),
	}

	var value attr.Value
	diags.Append(tfsdk.ValueFrom(
		ctx,
		newSerialDevice,
		SerialDevicesAttrType,
		&value,
	)...)
	return value
}

func newTags(_ context.Context, node *cmlclient.Node, diags *diag.Diagnostics) types.Set {
	// Node tags can't be null, there's always a set of tags, even if it's empty
	valueSet := make([]attr.Value, 0)
	for _, tag := range node.Tags {
		valueSet = append(valueSet, types.StringValue(tag))
	}
	tags, dia := types.SetValue(types.StringType, valueSet)
	diags.Append(dia...)
	return tags
}

func newSerialDevices(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) types.List {
	if len(node.SerialDevices) == 0 {
		return types.ListNull(SerialDevicesAttrType)
	}
	valueList := make([]attr.Value, 0)
	for _, serial_device := range node.SerialDevices {
		valueList = append(valueList, newSerialDevice(ctx, serial_device, diags))
	}
	serialDevices, dia := types.ListValue(
		SerialDevicesAttrType,
		valueList,
	)
	diags.Append(dia...)
	return serialDevices
}

func newInterfaces(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) types.List {
	if len(node.Interfaces) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: InterfaceAttrType})
	}
	valueList := make([]attr.Value, 0)
	for _, iface := range node.Interfaces {
		valueList = append(valueList, NewInterface(ctx, iface, diags))
	}
	ifaces, dia := types.ListValue(
		types.ObjectType{AttrTypes: InterfaceAttrType},
		valueList,
	)
	diags.Append(dia...)
	return ifaces
}

func NewNode(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) attr.Value {
	newNode := NodeModel{
		ID:             types.StringValue(node.ID),
		LabID:          types.StringValue(node.LabID),
		Label:          types.StringValue(node.Label),
		State:          types.StringValue(node.State),
		NodeDefinition: types.StringValue(node.NodeDefinition),
		Configuration: Config{
			StringValue: types.StringPointerValue(node.Configuration),
		},
		Interfaces:    newInterfaces(ctx, node, diags),
		Tags:          newTags(ctx, node, diags),
		X:             types.Int64Value(int64(node.X)),
		Y:             types.Int64Value(int64(node.Y)),
		HideLinks:     types.BoolValue(bool(node.HideLinks)),
		SerialDevices: newSerialDevices(ctx, node, diags),

		// these values are null if unset
		VNCkey:          types.StringNull(),
		RAM:             types.Int64Null(),
		CPUs:            types.Int64Null(),
		CPUlimit:        types.Int64Null(),
		ImageDefinition: types.StringNull(),
		ComputeID:       types.StringNull(),
		BootDiskSize:    types.Int64Null(),
		DataVolume:      types.Int64Null(),
	}

	if len(node.VNCkey) > 0 {
		newNode.VNCkey = types.StringValue(node.VNCkey)
	}
	if node.CPUlimit > 0 {
		newNode.CPUlimit = types.Int64Value(int64((node.CPUlimit)))
	}
	if node.RAM > 0 {
		newNode.RAM = types.Int64Value(int64(node.RAM))
	}
	if node.CPUs > 0 {
		newNode.CPUs = types.Int64Value(int64(node.CPUs))
	}
	if node.BootDiskSize > 0 {
		newNode.BootDiskSize = types.Int64Value(int64(node.BootDiskSize))
	}
	if node.DataVolume > 0 {
		newNode.DataVolume = types.Int64Value(int64(node.DataVolume))
	}
	if len(node.ComputeID) > 0 {
		newNode.ComputeID = types.StringValue(node.ComputeID)
	}
	if len(node.ImageDefinition) > 0 {
		newNode.ImageDefinition = types.StringValue(node.ImageDefinition)
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newNode,
			types.ObjectType{AttrTypes: NodeAttrType},
			&value,
		)...,
	)
	return value
}
