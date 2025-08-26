package cmlschema_test

import (
	"context"
	"testing"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/stretchr/testify/assert"
)

var image *cmlclient.ImageDefinition = &cmlclient.ImageDefinition{
	ID:            "22e72ed9-238f-4f8f-966f-4a5dc6b8de1a",
	SchemaVersion: "0.0.1",
	NodeDefID:     "testnodedef",
	Description:   "testdescription",
	Label:         "Test Node Label",
	DiskImage1:    "bla.qcpw2",
	DiskImage2:    "",
	DiskImage3:    "",
	ReadOnly:      true,
	DiskSubfolder: "bla",
	RAM:           nil,
	CPUs:          nil,
	CPUlimit:      nil,
	DataVolume:    nil,
	BootDiskSize:  nil,
}

func TestImageDef(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	value := cmlschema.NewImageDefinition(ctx, image, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newImage cmlschema.ImageDefinitionModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newImage)...)
	assert.False(t, diag.HasError())
}

func TestImageSchema(t *testing.T) {
	imageSchema := schema.Schema{
		Attributes: cmlschema.ImageDef(),
	}
	got, diag := imageSchema.TypeAtPath(context.TODO(), path.Root("id"))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
