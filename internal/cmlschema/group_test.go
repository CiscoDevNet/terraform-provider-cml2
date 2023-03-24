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

var group1 *cmlclient.Group = &cmlclient.Group{
	Name:        "CCNA Study Group Class of 21",
	Description: "string",
	Members: []string{
		"90f84e38-a71c-4d57-8d90-00fa8a197385",
		"60f84e39-ffff-4d99-8a78-00fa8aaf5666",
	},
	Labs: []struct {
		ID         string `json:"id"`
		Permission string `json:"permission"`
	}{
		{
			"90f84e38-a71c-4d57-8d90-00fa8a197385",
			"read_only",
		},
	},
	ID: "85401911-851f-4e6a-b5c3-4aa1d91fa21d",
}

var group2 *cmlclient.Group = &cmlclient.Group{
	Name:        "CCNA Study Group Class of 01",
	Description: "string",
	Members:     []string{},
	Labs: []struct {
		ID         string `json:"id"`
		Permission string `json:"permission"`
	}{},
	ID: "85401911-851f-4e6a-b5c3-4aa1d91fa21d",
}

func TestGroup(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	for _, group := range []*cmlclient.Group{group1, group2} {
		value := cmlschema.NewGroup(ctx, group, diag)
		t.Logf("value: %+v", value)
		t.Logf("errors: %+v", diag.Errors())
		assert.False(t, diag.HasError())
		var newGroup cmlschema.GroupModel
		diag.Append(tfsdk.ValueAs(ctx, value, &newGroup)...)
	}
	assert.False(t, diag.HasError())
}

func TestGroupSchema(t *testing.T) {
	groupSchema := schema.Schema{
		Attributes: cmlschema.Converter(cmlschema.Group()),
	}
	got, diag := groupSchema.TypeAtPath(context.TODO(), path.Root("id"))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
