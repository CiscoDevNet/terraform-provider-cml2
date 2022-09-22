package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

var iface *cmlclient.Interface = &cmlclient.Interface{
	ID:          "7c7285f5-c8c0-415a-a84d-59874347884a",
	Label:       "to router 5",
	Type:        "physical",
	Slot:        0,
	State:       "STARTED",
	MACaddress:  "fe:54:00:a7:b6:ae",
	IsConnected: true,
	DeviceName:  "eth0",
	IP4:         []string{"1.2.3.4/24"},
	IP6:         []string{"fe80::1/64"},
}

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
	Interfaces: cmlclient.InterfaceMap{
		"7c7285f5-c8c0-415a-a84d-59874347884a": iface,
	},
	Tags: []string{"red", "blue"},
}

func TestNewInterface(t *testing.T) {

	// r := NewLabResource().(*LabResource)
	t.Run("simple", func(t *testing.T) {
		// t.Parallel()
		diag := &diag.Diagnostics{}
		value := newInterface(context.Background(), iface, diag)
		t.Logf("value: %+v", value)

		if diag.HasError() {
			t.Fatalf("Having errors %+v", diag.Errors())
		}
	})
}

func TestNewNode(t *testing.T) {

	// r := NewLabResource().(*LabResource)
	t.Run("simple", func(t *testing.T) {
		// t.Parallel()
		diag := &diag.Diagnostics{}
		value := newNode(context.Background(), node, diag)
		t.Logf("value: %+v", value)

		if diag.HasError() {
			t.Fatalf("Having errors %+v", diag.Errors())
		}
	})
}
