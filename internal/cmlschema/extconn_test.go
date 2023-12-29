package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

var conn1 *cmlclient.ExtConn = &cmlclient.ExtConn{
	Label:        "NAT",
	DeviceName: "virbr0",
	Protected: false,
	Snooped: true,
	Tags: []string{
		"NAT",
	},
	ID: "58568fbb-e1f8-4b83-a1f8-148c656eed39",
}

var conn2 *cmlclient.ExtConn = &cmlclient.ExtConn{
	Label:        "System Bridge",
	DeviceName: "bridge0",
	Protected: true,
	Snooped: true,
	Tags: []string{
		"System Bridge",
	},
	ID: "92f95da2-10fd-4a25-931e-acb31a47962c",
}

func TestConnector(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	for _, connector := range []*cmlclient.ExtConn{conn1, conn2} {
		value := cmlschema.NewExtConn(ctx, connector, diag)
		t.Logf("value: %+v", value)
		t.Logf("errors: %+v", diag.Errors())
		assert.False(t, diag.HasError())
		var newExtConn cmlschema.ExtConnModel
		diag.Append(tfsdk.ValueAs(ctx, value, &newExtConn)...)
	}
	assert.False(t, diag.HasError())
}

func TestExtConnSchema(t *testing.T) {
	extconnSchema := schema.Schema{
		Attributes: cmlschema.Converter(cmlschema.ExtConn()),
	}
	got, diag := extconnSchema.TypeAtPath(context.TODO(), path.Root("id"))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
