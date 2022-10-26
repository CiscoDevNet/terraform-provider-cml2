package schema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
	"github.com/stretchr/testify/assert"
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

func TestInterface(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	value := schema.NewInterface(ctx, iface, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newIface schema.InterfaceModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newIface)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
	assert.Len(t, newIface.IP4.Elems, 1)
	assert.Len(t, newIface.IP6.Elems, 1)
}

func TestInterfaceSchema(t *testing.T) {
	schema := tfsdk.Schema{
		Attributes: schema.Interface(),
	}

	// got, diag := schema.TypeAtPath(ctx, path.Root("id").AtName("sub_test"))
	got, diag := schema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
