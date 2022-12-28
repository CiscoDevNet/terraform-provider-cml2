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

var link *cmlclient.Link = &cmlclient.Link{
	ID:      "0faf7c22-f466-41fd-9feb-d902220d55c8",
	State:   "DEFINED_ON_CORE",
	LabID:   "cd5fa81a-82aa-47da-98c5-7e9ac6c75a67",
	Label:   "sample label",
	PCAPkey: "",
	SrcID:   "b6a8bafb-5b64-4e2d-9e52-c492cad0f72a",
	DstID:   "6205678a-34b3-40f7-ad39-8133533b954b",
	SrcNode: "4dcd3095-349e-49e4-9ea2-1ad20207877f",
	DstNode: "fcf6ba6f-7db9-45b0-a6ca-c383f504aa2e",
	SrcSlot: -1,
	DstSlot: -1,
}

func TestNewLink(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	link.SrcSlot = 0
	link.DstSlot = 0

	value := cmlschema.NewLink(ctx, link, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newLink cmlschema.LinkModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newLink)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
}

func TestLinkAttrs(t *testing.T) {
	linkschema := schema.Schema{
		Attributes: cmlschema.Link(),
	}

	got, diag := linkschema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.Equal(t, 11, len(linkschema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
