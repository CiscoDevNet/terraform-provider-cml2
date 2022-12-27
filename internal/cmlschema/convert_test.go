package cmlschema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
)

func unknownAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"float": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"bla": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func TestConverter(t *testing.T) {
	nodeAttrs := Node()
	attrs := Converter(nodeAttrs)
	assert.Equal(t, len(attrs), len(nodeAttrs))
}

func TestUnknownAttrType(t *testing.T) {
	unknownAttrs := unknownAttrs()
	assert.Panics(t, func() {
		_ = Converter(unknownAttrs)
	})
}
