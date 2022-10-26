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
	Groups: []*cmlclient.Group{
		{
			ID:         "fe9acf37-c1dd-4628-9658-9020bae6e036",
			Permission: "bla",
		},
	},
}

func TestNewLab(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	value := schema.NewLab(ctx, lab, diag)
	t.Logf("value: %+v", value)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())

	var newLab schema.LabModel
	diag.Append(tfsdk.ValueAs(ctx, value, &newLab)...)
	t.Logf("errors: %+v", diag.Errors())
	assert.False(t, diag.HasError())
}

func TestLabAttrs(t *testing.T) {
	schema := tfsdk.Schema{
		Attributes: schema.Lab(),
	}

	got, diag := schema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.Equal(t, 11, len(schema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
