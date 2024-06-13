package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

func TestLifecycleAttrs(t *testing.T) {
	lifecycleschema := schema.Schema{
		Attributes: cmlschema.Lifecycle(),
	}

	got, diag := lifecycleschema.TypeAtPath(context.TODO(), path.Root("id"))
	t.Log(diag.Errors())
	assert.Equal(t, 12, len(lifecycleschema.Attributes))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}
