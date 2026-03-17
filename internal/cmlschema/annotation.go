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

// AnnotationModel is the Terraform representation of a CML annotation.
type AnnotationModel struct {
	ID        types.String `tfsdk:"id"`
	LabID     types.String `tfsdk:"lab_id"`
	Type      types.String `tfsdk:"type"`
	Text      types.Object `tfsdk:"text"`
	Rectangle types.Object `tfsdk:"rectangle"`
	Ellipse   types.Object `tfsdk:"ellipse"`
	Line      types.Object `tfsdk:"line"`
}

// AnnotationTextModel is the Terraform representation of a text annotation.
type AnnotationTextModel struct {
	TextContent types.String  `tfsdk:"text_content"`
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
}

// AnnotationRectangleModel is the Terraform representation of a rectangle annotation.
type AnnotationRectangleModel struct {
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	X2          types.Float64 `tfsdk:"x2"`
	Y2          types.Float64 `tfsdk:"y2"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
}

// AnnotationEllipseModel is the Terraform representation of an ellipse annotation.
type AnnotationEllipseModel struct {
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	X2          types.Float64 `tfsdk:"x2"`
	Y2          types.Float64 `tfsdk:"y2"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
}

// AnnotationLineModel is the Terraform representation of a line annotation.
type AnnotationLineModel struct {
	X1        types.Float64 `tfsdk:"x1"`
	Y1        types.Float64 `tfsdk:"y1"`
	X2        types.Float64 `tfsdk:"x2"`
	Y2        types.Float64 `tfsdk:"y2"`
	Color     types.String  `tfsdk:"color"`
	Thickness types.Float64 `tfsdk:"thickness"`
	ZIndex    types.Float64 `tfsdk:"z_index"`
	LineStart types.String  `tfsdk:"line_start"`
	LineEnd   types.String  `tfsdk:"line_end"`
}

// AnnotationTextAttrType is the attribute type map for AnnotationTextModel.
var AnnotationTextAttrType = map[string]attr.Type{
	"text_content": types.StringType,
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
}

// AnnotationRectangleAttrType is the attribute type map for AnnotationRectangleModel.
var AnnotationRectangleAttrType = map[string]attr.Type{
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"x2":           types.Float64Type,
	"y2":           types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
}

// AnnotationEllipseAttrType is the attribute type map for AnnotationEllipseModel.
var AnnotationEllipseAttrType = map[string]attr.Type{
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"x2":           types.Float64Type,
	"y2":           types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
}

// AnnotationLineAttrType is the attribute type map for AnnotationLineModel.
var AnnotationLineAttrType = map[string]attr.Type{
	"x1":         types.Float64Type,
	"y1":         types.Float64Type,
	"x2":         types.Float64Type,
	"y2":         types.Float64Type,
	"color":      types.StringType,
	"thickness":  types.Float64Type,
	"z_index":    types.Float64Type,
	"line_start": types.StringType,
	"line_end":   types.StringType,
}

// AnnotationAttrType is the attribute type map for AnnotationModel.
// AnnotationAttrType is the attribute type map for AnnotationModel.
var AnnotationAttrType = map[string]attr.Type{
	"id":        types.StringType,
	"lab_id":    types.StringType,
	"type":      types.StringType,
	"text":      types.ObjectType{AttrTypes: AnnotationTextAttrType},
	"rectangle": types.ObjectType{AttrTypes: AnnotationRectangleAttrType},
	"ellipse":   types.ObjectType{AttrTypes: AnnotationEllipseAttrType},
	"line":      types.ObjectType{AttrTypes: AnnotationLineAttrType},
}

// Annotation returns the schema for the annotation resource.
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
			Description: "Annotation type. Supported: text, rectangle, ellipse, line.",
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
		"rectangle": schema.SingleNestedAttribute{
			Description: "Rectangle annotation attributes (required when type = \"rectangle\").",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1":           schema.Float64Attribute{Required: true},
				"y1":           schema.Float64Attribute{Required: true},
				"x2":           schema.Float64Attribute{Required: true},
				"y2":           schema.Float64Attribute{Required: true},
				"color":        schema.StringAttribute{Optional: true, Computed: true},
				"border_color": schema.StringAttribute{Optional: true, Computed: true},
				"thickness":    schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":      schema.Float64Attribute{Optional: true, Computed: true},
			},
		},
		"ellipse": schema.SingleNestedAttribute{
			Description: "Ellipse annotation attributes (required when type = \"ellipse\").",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1":           schema.Float64Attribute{Required: true},
				"y1":           schema.Float64Attribute{Required: true},
				"x2":           schema.Float64Attribute{Required: true},
				"y2":           schema.Float64Attribute{Required: true},
				"color":        schema.StringAttribute{Optional: true, Computed: true},
				"border_color": schema.StringAttribute{Optional: true, Computed: true},
				"thickness":    schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":      schema.Float64Attribute{Optional: true, Computed: true},
			},
		},
		"line": schema.SingleNestedAttribute{
			Description: "Line annotation attributes (required when type = \"line\").",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1":         schema.Float64Attribute{Required: true},
				"y1":         schema.Float64Attribute{Required: true},
				"x2":         schema.Float64Attribute{Required: true},
				"y2":         schema.Float64Attribute{Required: true},
				"color":      schema.StringAttribute{Optional: true, Computed: true},
				"thickness":  schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":    schema.Float64Attribute{Optional: true, Computed: true},
				"line_start": schema.StringAttribute{Optional: true, Computed: true},
				"line_end":   schema.StringAttribute{Optional: true, Computed: true},
			},
		},
	}
}

// NewAnnotation converts a CML annotation into a Terraform value.
func NewAnnotation(ctx context.Context, labID models.UUID, a models.Annotation, diags *diag.Diagnostics) attr.Value {
	model := AnnotationModel{
		ID:        types.StringNull(),
		LabID:     types.StringValue(string(labID)),
		Type:      types.StringValue(string(a.Type)),
		Text:      types.ObjectNull(AnnotationTextAttrType),
		Rectangle: types.ObjectNull(AnnotationRectangleAttrType),
		Ellipse:   types.ObjectNull(AnnotationEllipseAttrType),
		Line:      types.ObjectNull(AnnotationLineAttrType),
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
	case models.AnnotationTypeRectangle:
		if a.Rectangle != nil {
			rec := AnnotationRectangleModel{
				X1:          types.Float64Value(a.Rectangle.X1),
				Y1:          types.Float64Value(a.Rectangle.Y1),
				X2:          types.Float64Value(a.Rectangle.X2),
				Y2:          types.Float64Value(a.Rectangle.Y2),
				Color:       types.StringValue(a.Rectangle.Color),
				BorderColor: types.StringValue(a.Rectangle.BorderColor),
				Thickness:   types.Float64Value(a.Rectangle.Thickness),
				ZIndex:      types.Float64Value(a.Rectangle.ZIndex),
			}
			model.ID = types.StringValue(string(a.Rectangle.ID))
			var v attr.Value
			diags.Append(tfsdk.ValueFrom(ctx, rec, types.ObjectType{AttrTypes: AnnotationRectangleAttrType}, &v)...)
			model.Rectangle = v.(types.Object)
		}
	case models.AnnotationTypeEllipse:
		if a.Ellipse != nil {
			el := AnnotationEllipseModel{
				X1:          types.Float64Value(a.Ellipse.X1),
				Y1:          types.Float64Value(a.Ellipse.Y1),
				X2:          types.Float64Value(a.Ellipse.X2),
				Y2:          types.Float64Value(a.Ellipse.Y2),
				Color:       types.StringValue(a.Ellipse.Color),
				BorderColor: types.StringValue(a.Ellipse.BorderColor),
				Thickness:   types.Float64Value(a.Ellipse.Thickness),
				ZIndex:      types.Float64Value(a.Ellipse.ZIndex),
			}
			model.ID = types.StringValue(string(a.Ellipse.ID))
			var v attr.Value
			diags.Append(tfsdk.ValueFrom(ctx, el, types.ObjectType{AttrTypes: AnnotationEllipseAttrType}, &v)...)
			model.Ellipse = v.(types.Object)
		}
	case models.AnnotationTypeLine:
		if a.Line != nil {
			lineStart := ""
			if a.Line.LineStart != nil {
				lineStart = string(*a.Line.LineStart)
			}
			lineEnd := ""
			if a.Line.LineEnd != nil {
				lineEnd = string(*a.Line.LineEnd)
			}
			ln := AnnotationLineModel{
				X1:        types.Float64Value(a.Line.X1),
				Y1:        types.Float64Value(a.Line.Y1),
				X2:        types.Float64Value(a.Line.X2),
				Y2:        types.Float64Value(a.Line.Y2),
				Color:     types.StringValue(a.Line.Color),
				Thickness: types.Float64Value(a.Line.Thickness),
				ZIndex:    types.Float64Value(a.Line.ZIndex),
				LineStart: types.StringValue(lineStart),
				LineEnd:   types.StringValue(lineEnd),
			}
			model.ID = types.StringValue(string(a.Line.ID))
			var v attr.Value
			diags.Append(tfsdk.ValueFrom(ctx, ln, types.ObjectType{AttrTypes: AnnotationLineAttrType}, &v)...)
			model.Line = v.(types.Object)
		}
	default:
		// unsupported types remain null blocks
	}

	var out attr.Value
	diags.Append(tfsdk.ValueFrom(ctx, model, types.ObjectType{AttrTypes: AnnotationAttrType}, &out)...)
	return out
}
