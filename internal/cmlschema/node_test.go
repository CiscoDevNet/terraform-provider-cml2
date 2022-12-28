package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

var node *cmlclient.Node = &cmlclient.Node{
	ID:              "8bf321c3-3312-44f2-9098-fa89e2e05d7e",
	Label:           "router 1",
	X:               10,
	Y:               20,
	NodeDefinition:  "IOSv",
	ImageDefinition: "",
	Configuration:   "hostname router1",
	CPUs:            0,
	CPUlimit:        100,
	RAM:             512,
	State:           "BOOTED",
	Interfaces: cmlclient.InterfaceList{
		iface,
		iface,
	},
	SerialDevices: []cmlclient.SerialDevice{
		{
			ConsoleKey:   "1eab9ba0-c92e-4568-a742-6b4b2244c5b2",
			DeviceNumber: 0,
		},
	},
	Tags:         []string{"red", "blue"},
	ComputeID:    "f3678fb5-985d-45c2-b0f5-e54174798912",
	BootDiskSize: 16,
	DataVolume:   0,
}

func TestNewNode(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	value := cmlschema.NewNode(ctx, node, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newNode cmlschema.NodeModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newNode)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
	assert.Len(t, newNode.Interfaces.Elements(), 2)
}

func TestNodeAttrs(t *testing.T) {
	nodeschema := schema.Schema{
		Attributes: cmlschema.Node(),
	}

	got, diag := nodeschema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.Equal(t, 19, len(nodeschema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
