package cmlschema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*Config)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*Config)(nil)
)

type Config struct {
	basetypes.StringValue
	// ... potentially other fields ...
}

// Type returns a ConfigType.
func (v Config) Type(_ context.Context) attr.Type {
	return ConfigType{}
}

// Equal returns true if the given value is equivalent.
func (v Config) Equal(o attr.Value) bool {
	other, ok := o.(Config)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals compares the provided configurations independent of line endings (DOS vs Unix)
func (v Config) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newConfig, ok := newValuable.(Config)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}
	oldCfg := strings.ReplaceAll(v.ValueString(), "\r\n", "\n")
	newCfg := strings.ReplaceAll(newConfig.ValueString(), "\r\n", "\n")
	return newCfg == oldCfg, diags
}

// NewConfigNull creates a Config with a null value. Determine whether the value is null via IsNull method.
func NewConfigNull() Config {
	return Config{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewConfigUnknown creates a Config with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewConfigUnknown() Config {
	return Config{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewConfigValue creates a Config with a known value. Access the value via ValueString method.
func NewConfigValue(value string) Config {
	return Config{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewConfigPointerValue creates a Config with a null value if nil or a known value. Access the value via ValueStringPointer method.
func NewConfigPointerValue(value *string) Config {
	return Config{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}
