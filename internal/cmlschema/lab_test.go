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
	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

var lab *cmlclient.Lab = &cmlclient.Lab{
	ID:          "0faf7c22-f466-41fd-9feb-d902220d55c8",
	State:       "DEFINED_ON_CORE",
	Created:     "2022-10-24T10:33:30+00:00",
	Modified:    "2022-10-24T10:33:35+00:00",
	Title:       "",
	Description: "",
	Notes:       "",
	Owner:       &cmlclient.User{},
	NodeCount:   0,
	LinkCount:   0,
	Nodes:       make(cmlclient.NodeMap),
	Links:       []*cmlclient.Link{},
	Groups: []*cmlclient.LabGroup{
		{
			ID:         "fe9acf37-c1dd-4628-9658-9020bae6e036",
			Name:       "students",
			Permission: "bla",
		},
	},
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
	assert.Equal(t, 11, len(labschema.Attributes))
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
