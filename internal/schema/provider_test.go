package schema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestProviderAttrs(t *testing.T) {
	schema := tfsdk.Schema{
		Attributes: schema.Provider(),
	}

	got, diag := schema.TypeAtPath(context.TODO(), path.Root("address"))
	assert.Equal(t, types.StringType, got)
	assert.Equal(t, 7, len(schema.Attributes))
	assert.False(t, diag.HasError())
	t.Log(diag.Errors())
}
