package cmlschema

import (
	"fmt"

	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	r_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Converter(rSchema map[string]r_schema.Attribute) map[string]ds_schema.Attribute {
	dSchema := make(map[string]ds_schema.Attribute)
	for name, fromAttr := range rSchema {

		// required := fromAttr.IsRequired()
		// computed := fromAttr.IsComputed()
		// optional := fromAttr.IsOptional()

		// for a datasource, all attributes are computed and the required attrs
		// are on the container / outside.
		required := false
		computed := true
		optional := false

		if !(required && computed && optional) {
			computed = true
		}

		switch fromAttrType := fromAttr.(type) {
		case r_schema.StringAttribute:
			dSchema[name] = ds_schema.StringAttribute{
				Validators:          fromAttrType.Validators,
				Description:         fromAttrType.Description,
				MarkdownDescription: fromAttrType.MarkdownDescription,
				CustomType:          fromAttrType.CustomType,
				Sensitive:           fromAttrType.Sensitive,
				Optional:            optional,
				Computed:            computed,
				Required:            required,
			}
		case r_schema.BoolAttribute:
			dSchema[name] = ds_schema.BoolAttribute{
				Validators:          fromAttrType.Validators,
				Description:         fromAttrType.Description,
				MarkdownDescription: fromAttrType.MarkdownDescription,
				CustomType:          fromAttrType.CustomType,
				Sensitive:           fromAttrType.Sensitive,
				Optional:            optional,
				Computed:            computed,
				Required:            required,
			}
		case r_schema.Int64Attribute:
			dSchema[name] = ds_schema.Int64Attribute{
				Validators:          fromAttrType.Validators,
				Description:         fromAttrType.Description,
				MarkdownDescription: fromAttrType.MarkdownDescription,
				CustomType:          fromAttrType.CustomType,
				Sensitive:           fromAttrType.Sensitive,
				Optional:            optional,
				Computed:            computed,
				Required:            required,
			}
		case r_schema.ListAttribute:
			dSchema[name] = ds_schema.ListAttribute{
				Validators:          fromAttrType.Validators,
				Description:         fromAttrType.Description,
				MarkdownDescription: fromAttrType.MarkdownDescription,
				CustomType:          fromAttrType.CustomType,
				Sensitive:           fromAttrType.Sensitive,
				ElementType:         fromAttrType.ElementType,
				Optional:            optional,
				Computed:            computed,
				Required:            required,
			}
		case r_schema.ListNestedAttribute:
			dSchema[name] = ds_schema.ListNestedAttribute{
				Validators:          fromAttrType.Validators,
				Description:         fromAttrType.Description,
				MarkdownDescription: fromAttrType.MarkdownDescription,
				Sensitive:           fromAttrType.Sensitive,
				NestedObject: ds_schema.NestedAttributeObject{
					Attributes: Converter(fromAttrType.NestedObject.Attributes),
				},
				Optional: optional,
				Computed: computed,
				Required: required,
			}
		default:
			msg := fmt.Sprintf("unknown attribute type: %v", fromAttr.GetType().String())
			panic(msg)
		}
	}

	return dSchema
}
