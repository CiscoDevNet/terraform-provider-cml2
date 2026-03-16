package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"
)

type AnnotationModel struct {
	ID    types.String `tfsdk:"id"`
	LabID types.String `tfsdk:"lab_id"`
	Type  types.String `tfsdk:"type"`
	Text  types.Object `tfsdk:"text"`
}

type AnnotationTextModel struct {
	TextContent types.String  `tfsdk:"text_content"`
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
}

var AnnotationTextAttrType = map[string]attr.Type{
	"text_content": types.StringType,
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
}

var AnnotationAttrType = map[string]attr.Type{
	"id":     types.StringType,
	"lab_id": types.StringType,
	"type":   types.StringType,
	"text":   types.ObjectType{AttrTypes: AnnotationTextAttrType},
}

func Annotation() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Annotation ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"lab_id": schema.StringAttribute{
			Description: "Lab ID.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": schema.StringAttribute{
			Description: "Annotation type. Currently supported: text.",
			Required:    true,
		},
		"text": schema.SingleNestedAttribute{
			Description: "Text annotation attributes (required when type = \"text\").",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"text_content": schema.StringAttribute{
					Required: true,
				},
				"x1": schema.Float64Attribute{
					Required: true,
				},
				"y1": schema.Float64Attribute{
					Required: true,
				},
				"color": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"border_color": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"thickness": schema.Float64Attribute{
					Optional: true,
					Computed: true,
				},
				"z_index": schema.Float64Attribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func NewAnnotation(ctx context.Context, labID models.UUID, a models.Annotation, diags *diag.Diagnostics) attr.Value {
	model := AnnotationModel{
		ID:    types.StringNull(),
		LabID: types.StringValue(string(labID)),
		Type:  types.StringValue(string(a.Type)),
		Text:  types.ObjectNull(AnnotationTextAttrType),
	}

	switch a.Type {
	case models.AnnotationTypeText:
		if a.Text != nil {
			text := AnnotationTextModel{
				TextContent: types.StringValue(a.Text.TextContent),
				X1:          types.Float64Value(a.Text.X1),
				Y1:          types.Float64Value(a.Text.Y1),
				Color:       types.StringValue(a.Text.Color),
				BorderColor: types.StringValue(a.Text.BorderColor),
				Thickness:   types.Float64Value(a.Text.Thickness),
				ZIndex:      types.Float64Value(a.Text.ZIndex),
			}
			model.ID = types.StringValue(string(a.Text.ID))
			var v attr.Value
			diags.Append(tfsdk.ValueFrom(ctx, text, types.ObjectType{AttrTypes: AnnotationTextAttrType}, &v)...)
			model.Text = v.(types.Object)
		}
	default:
		// unsupported types remain null blocks
	}

	var out attr.Value
	diags.Append(tfsdk.ValueFrom(ctx, model, types.ObjectType{AttrTypes: AnnotationAttrType}, &out)...)
	return out
}
