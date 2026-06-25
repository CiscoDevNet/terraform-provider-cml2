package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

func TestProviderAttrs(t *testing.T) {
	schema := tfschema.Schema{
		Attributes: cmlschema.Provider(),
	}

	got, diag := schema.TypeAtPath(context.TODO(), path.Root("address"))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
	got, diag = schema.TypeAtPath(context.TODO(), path.Root("request_headers"))
	assert.Equal(t, types.MapType{ElemType: types.StringType}, got)
	assert.Equal(t, 12, len(schema.Attributes))
	assert.False(t, diag.HasError())
	t.Log(diag.Errors())
}

func TestApplyEnvVarsPopulatesNullAttrs(t *testing.T) {
	t.Setenv("CML2_ADDRESS", "https://cml.example.com")
	t.Setenv("CML2_USERNAME", "admin")
	t.Setenv("CML2_PASSWORD", "secret")
	t.Setenv("CML2_TOKEN", "jwt-token")
	t.Setenv("CML2_TOKEN_CACHE_FILE", "/tmp/cache.json")
	t.Setenv("CML2_CACERT", "PEM")
	t.Setenv("CML2_TOKEN_CACHE", "true")
	t.Setenv("CML2_SKIP_VERIFY", "1")
	t.Setenv("CML2_NAMED_CONFIGS", "True")
	t.Setenv("CML2_DYNAMIC_CONFIG", "0")

	m := cmlschema.ProviderModel{}
	diags := m.ApplyEnvVars()

	assert.False(t, diags.HasError())
	assert.Equal(t, "https://cml.example.com", m.Address.ValueString())
	assert.Equal(t, "admin", m.Username.ValueString())
	assert.Equal(t, "secret", m.Password.ValueString())
	assert.Equal(t, "jwt-token", m.Token.ValueString())
	assert.Equal(t, "/tmp/cache.json", m.TokenCacheFile.ValueString())
	assert.Equal(t, "PEM", m.CAcert.ValueString())
	assert.True(t, m.TokenCache.ValueBool())
	assert.True(t, m.SkipVerify.ValueBool())
	assert.True(t, m.NamedConfigs.ValueBool())
	assert.False(t, m.DynamicConfig.ValueBool())
}

func TestApplyEnvVarsExplicitConfigTakesPrecedence(t *testing.T) {
	t.Setenv("CML2_ADDRESS", "https://env.example.com")
	t.Setenv("CML2_NAMED_CONFIGS", "true")

	m := cmlschema.ProviderModel{
		Address:      types.StringValue("https://config.example.com"),
		NamedConfigs: types.BoolValue(false),
	}
	diags := m.ApplyEnvVars()

	assert.False(t, diags.HasError())
	assert.Equal(t, "https://config.example.com", m.Address.ValueString())
	assert.False(t, m.NamedConfigs.ValueBool())
}
