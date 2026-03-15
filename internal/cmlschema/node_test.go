package cmlschema_test

import (
	"context"
	"testing"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

var (
	config              = "hostname router1"
	node   *models.Node = func() *models.Node {
		img := "iosv-15.9"
		cpuLimit := 100
		ram := 512
		computeID := models.UUID("f3678fb5-985d-45c2-b0f5-e54174798912")
		bootDisk := 16
		dataVol := 64
		vnc := models.UUID("b1bea19f-e2b9-4a72-98f7-7663339cf317")
		cfg := config
		return &models.Node{
			ID:              "8bf321c3-3312-44f2-9098-fa89e2e05d7e",
			Label:           "router 1",
			X:               10,
			Y:               20,
			NodeDefinition:  "IOSv",
			ImageDefinition: &img,
			Configuration:   cfg,
			CPUs:            1,
			CPUlimit:        &cpuLimit,
			RAM:             &ram,
			State:           models.NodeStateBooted,
			Interfaces: models.InterfaceList{
				iface,
				iface,
			},
			SerialDevices: []models.SerialDevice{{ConsoleKey: "1eab9ba0-c92e-4568-a742-6b4b2244c5b2", DeviceNumber: 0}},
			Tags:          []string{"red", "blue"},
			ComputeID:     &computeID,
			BootDiskSize:  &bootDisk,
			DataVolume:    &dataVol,
			VNCkey:        &vnc,
		}
	}()
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
		node *models.Node
		want int
	}{
		{
			"empty",
			&models.Node{},
			0,
		},
		{
			"one-element",
			&models.Node{
				Configurations: []models.NodeConfig{
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
