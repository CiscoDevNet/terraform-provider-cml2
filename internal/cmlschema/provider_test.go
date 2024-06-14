package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/stretchr/testify/assert"
)

func TestProviderAttrs(t *testing.T) {
	schema := tfschema.Schema{
		Attributes: cmlschema.Provider(),
	}

	got, diag := schema.TypeAtPath(context.TODO(), path.Root("address"))
	assert.Equal(t, types.StringType, got)
	assert.Equal(t, 9, len(schema.Attributes))
	assert.False(t, diag.HasError())
	t.Log(diag.Errors())
}
