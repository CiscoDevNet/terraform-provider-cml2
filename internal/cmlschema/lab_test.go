package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

var lab *models.Lab = &models.Lab{
	ID:          "0faf7c22-f466-41fd-9feb-d902220d55c8",
	State:       models.LabStateDefined,
	Created:     "2022-10-24T10:33:30+00:00",
	Modified:    "2022-10-24T10:33:35+00:00",
	Title:       "",
	Description: "",
	Notes:       "",
	OwnerID:     "00000000-0000-4000-a000-000000000000",
	NodeCount:   0,
	LinkCount:   0,
	Nodes:       make(models.NodeMap),
	Links:       models.LinkList{},
}

func TestNewLab(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	value := cmlschema.NewLab(ctx, lab, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newLab cmlschema.LabModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newLab)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
}

func TestLabAttrs(t *testing.T) {
	labschema := schema.Schema{Attributes: cmlschema.Lab()}
	got, diag := labschema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	groups, diag := labschema.TypeAtPath(context.TODO(), path.Root("groups"))
	t.Log(diag.Errors())
	assert.Equal(t, 12, len(labschema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
	assert.Equal(t,
		types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: cmlschema.LabGroupAttrType,
			},
		},
		groups,
	)
}
