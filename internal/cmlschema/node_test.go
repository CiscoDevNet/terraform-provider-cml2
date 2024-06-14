package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

var (
	config                 = "hostname router1"
	node   *cmlclient.Node = &cmlclient.Node{
		ID:              "8bf321c3-3312-44f2-9098-fa89e2e05d7e",
		Label:           "router 1",
		X:               10,
		Y:               20,
		NodeDefinition:  "IOSv",
		ImageDefinition: "iosv-15.9",
		Configuration:   &config,
		CPUs:            1,
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
		DataVolume:   64,
		VNCkey:       "b1bea19f-e2b9-4a72-98f7-7663339cf317",
	}
)

// func newPtr[V int | string](value V) *V {
// 	ptr := new(V)
// 	*ptr = value
// 	return ptr
// }

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
	assert.Len(t, newNode.SerialDevices.Elements(), 1)
}

func TestNewNodeVariant(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	// modifies node!!
	node.Interfaces = nil
	node.SerialDevices = nil

	value := cmlschema.NewNode(ctx, node, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newNode cmlschema.NodeModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newNode)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
	assert.Len(t, newNode.Interfaces.Elements(), 0)
	assert.Len(t, newNode.SerialDevices.Elements(), 0)
}

func TestNodeAttrs(t *testing.T) {
	nodeschema := schema.Schema{
		Attributes: cmlschema.Node(),
	}

	got, diag := nodeschema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.Equal(t, 21, len(nodeschema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}

func TestNewNamedConfigs(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	tests := []struct {
		name string
		node *cmlclient.Node
		want int
	}{
		{
			"empty",
			&cmlclient.Node{},
			0,
		},
		{
			"one-element",
			&cmlclient.Node{
				Configurations: []cmlclient.NodeConfig{
					{Name: "bla", Content: "hostname bla"},
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmlschema.NewNamedConfigs(ctx, tt.node, diag); len(got.Elements()) != tt.want {
				t.Errorf("NewNamedConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
