package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	Configuration   types.String `tfsdk:"configuration"`
	Interfaces      types.List   `tfsdk:"interfaces"`
	Tags            types.List   `tfsdk:"tags"`
	X               types.Int64  `tfsdk:"x"`
	Y               types.Int64  `tfsdk:"y"`
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
	"configuration":   types.StringType,
	"interfaces": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: InterfaceAttrType,
		},
	},
	"tags":           types.ListType{ElemType: types.StringType},
	"x":              types.Int64Type,
	"y":              types.Int64Type,
	"cpus":           types.Int64Type,
	"cpu_limit":      types.Int64Type,
	"ram":            types.Int64Type,
	"boot_disk_size": types.Int64Type,
	"data_volume":    types.Int64Type,
	"vnc_key":        types.StringType,
	"serial_devices": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: serialKeyAttrType,
		},
	},
	"compute_id": types.StringType,
}

var serialKeyAttrType = map[string]attr.Type{
	"console_key":   types.StringType,
	"device_number": types.Int64Type,
}

func Node() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "node ID (UUID)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"lab_id": {
			Description: "lab ID containing the node (UUID)",
			Type:        types.StringType,
			Required:    true,
		},
		"label": {
			Description: "label",
			Type:        types.StringType,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			Description: "state",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"nodedefinition": {
			Description: "node definition / type",
			Type:        types.StringType,
			Required:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"imagedefinition": {
			Description: "image definition / type",
			Type:        types.StringType,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"interfaces": {
			Description: "list of interfaces on the node",
			Computed:    true,
			Attributes: tfsdk.ListNestedAttributes(
				Interface(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"tags": {
			Description: "tags of the node",
			Computed:    true,
			Optional:    true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"configuration": {
			Description: "node configuration",
			Type:        types.StringType,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"x": {
			Description: "x coordinate",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"y": {
			Description: "x coordinate",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ram": {
			Description: "amount of RAM, megabytes",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"cpus": {
			Description: "number of cpus",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"cpu_limit": {
			Description: "cpu limit in %, 20-100",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"boot_disk_size": {
			Description: "size of boot disk volume, in GB",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"data_volume": {
			Description: "size of data volume, in GB",
			Type:        types.Int64Type,
			Computed:    true,
			Optional:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
				resource.RequiresReplace(),
			},
		},
		"serial_devices": {
			Description: "a list of serial devices (consoles)",
			Computed:    true,
			Type: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: serialKeyAttrType,
				},
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"vnc_key": {
			Description: "VNC key of console, a UUID4",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"compute_id": {
			Description: "ID of a compute this node is on, a UUID4",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

func newSerialDevices(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) types.List {

	serialDevices := types.List{
		ElemType: types.ObjectType{
			AttrTypes: serialKeyAttrType,
		},
		Elems: make([]attr.Value, 0),
	}

	for _, serial_device := range node.SerialDevices {

		newSerialDevice := serialDeviceModel{
			ConsoleKey:   types.String{Value: serial_device.ConsoleKey},
			DeviceNumber: types.Int64{Value: int64(serial_device.DeviceNumber)},
		}

		var value attr.Value
		diags.Append(tfsdk.ValueFrom(
			ctx,
			newSerialDevice,
			types.ObjectType{AttrTypes: serialKeyAttrType},
			&value,
		)...)

		serialDevices.Elems = append(serialDevices.Elems, value)
	}
	return serialDevices
}

func NewNode(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) attr.Value {

	ifaces := types.List{ElemType: types.ObjectType{
		AttrTypes: InterfaceAttrType,
	}}
	for _, iface := range node.Interfaces {
		value := NewInterface(ctx, iface, diags)
		ifaces.Elems = append(ifaces.Elems, value)
	}

	tags := types.List{ElemType: types.StringType}
	for _, tag := range node.Tags {
		tags.Elems = append(tags.Elems, types.String{Value: tag})
	}

	newNode := NodeModel{
		ID:              types.String{Value: node.ID},
		LabID:           types.String{Value: node.LabID},
		Label:           types.String{Value: node.Label},
		State:           types.String{Value: node.State},
		NodeDefinition:  types.String{Value: node.NodeDefinition},
		ImageDefinition: types.String{Value: node.ImageDefinition},
		Configuration:   types.String{Value: node.Configuration},
		Interfaces:      ifaces,
		Tags:            tags,
		X:               types.Int64{Value: int64(node.X)},
		Y:               types.Int64{Value: int64(node.Y)},
		SerialDevices:   newSerialDevices(ctx, node, diags),
		CPUs:            types.Int64{Value: int64(node.CPUs)},
		CPUlimit:        types.Int64{Value: int64(node.CPUlimit)},
		RAM:             types.Int64{Value: int64(node.RAM)},
		VNCkey:          types.String{Value: node.VNCkey},

		// these values are null if there's no compute ID
		ComputeID:    types.String{Null: true},
		BootDiskSize: types.Int64{Null: true},
		DataVolume:   types.Int64{Null: true},
	}

	// set them, if there IS a compute ID
	if len(node.ComputeID) > 0 {
		newNode.ComputeID = types.String{Value: node.ComputeID}
		newNode.BootDiskSize = types.Int64{Value: int64(node.BootDiskSize)}
		newNode.DataVolume = types.Int64{Value: int64(node.DataVolume)}
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
