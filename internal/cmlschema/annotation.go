package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Rotation    types.Float64 `tfsdk:"rotation"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	BorderStyle types.String  `tfsdk:"border_style"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
	TextBold    types.Bool    `tfsdk:"text_bold"`
	TextFont    types.String  `tfsdk:"text_font"`
	TextItalic  types.Bool    `tfsdk:"text_italic"`
	TextSize    types.Float64 `tfsdk:"text_size"`
	TextUnit    types.String  `tfsdk:"text_unit"`
}

// AnnotationRectangleModel is the Terraform representation of a rectangle annotation.
type AnnotationRectangleModel struct {
	X1           types.Float64 `tfsdk:"x1"`
	Y1           types.Float64 `tfsdk:"y1"`
	X2           types.Float64 `tfsdk:"x2"`
	Y2           types.Float64 `tfsdk:"y2"`
	Rotation     types.Float64 `tfsdk:"rotation"`
	Color        types.String  `tfsdk:"color"`
	BorderColor  types.String  `tfsdk:"border_color"`
	BorderStyle  types.String  `tfsdk:"border_style"`
	Thickness    types.Float64 `tfsdk:"thickness"`
	ZIndex       types.Float64 `tfsdk:"z_index"`
	BorderRadius types.Float64 `tfsdk:"border_radius"`
}

// AnnotationEllipseModel is the Terraform representation of an ellipse annotation.
type AnnotationEllipseModel struct {
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	X2          types.Float64 `tfsdk:"x2"`
	Y2          types.Float64 `tfsdk:"y2"`
	Rotation    types.Float64 `tfsdk:"rotation"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	BorderStyle types.String  `tfsdk:"border_style"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
}

// AnnotationLineModel is the Terraform representation of a line annotation.
type AnnotationLineModel struct {
	X1          types.Float64 `tfsdk:"x1"`
	Y1          types.Float64 `tfsdk:"y1"`
	X2          types.Float64 `tfsdk:"x2"`
	Y2          types.Float64 `tfsdk:"y2"`
	Color       types.String  `tfsdk:"color"`
	BorderColor types.String  `tfsdk:"border_color"`
	BorderStyle types.String  `tfsdk:"border_style"`
	Thickness   types.Float64 `tfsdk:"thickness"`
	ZIndex      types.Float64 `tfsdk:"z_index"`
	LineStart   types.String  `tfsdk:"line_start"`
	LineEnd     types.String  `tfsdk:"line_end"`
}

// AnnotationTextAttrType is the attribute type map for AnnotationTextModel.
var AnnotationTextAttrType = map[string]attr.Type{
	"text_content": types.StringType,
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"rotation":     types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"border_style": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
	"text_bold":    types.BoolType,
	"text_font":    types.StringType,
	"text_italic":  types.BoolType,
	"text_size":    types.Float64Type,
	"text_unit":    types.StringType,
}

// AnnotationRectangleAttrType is the attribute type map for AnnotationRectangleModel.
var AnnotationRectangleAttrType = map[string]attr.Type{
	"x1":            types.Float64Type,
	"y1":            types.Float64Type,
	"x2":            types.Float64Type,
	"y2":            types.Float64Type,
	"rotation":      types.Float64Type,
	"color":         types.StringType,
	"border_color":  types.StringType,
	"border_style":  types.StringType,
	"thickness":     types.Float64Type,
	"z_index":       types.Float64Type,
	"border_radius": types.Float64Type,
}

// AnnotationEllipseAttrType is the attribute type map for AnnotationEllipseModel.
var AnnotationEllipseAttrType = map[string]attr.Type{
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"x2":           types.Float64Type,
	"y2":           types.Float64Type,
	"rotation":     types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"border_style": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
}

// AnnotationLineAttrType is the attribute type map for AnnotationLineModel.
var AnnotationLineAttrType = map[string]attr.Type{
	"x1":           types.Float64Type,
	"y1":           types.Float64Type,
	"x2":           types.Float64Type,
	"y2":           types.Float64Type,
	"color":        types.StringType,
	"border_color": types.StringType,
	"border_style": types.StringType,
	"thickness":    types.Float64Type,
	"z_index":      types.Float64Type,
	"line_start":   types.StringType,
	"line_end":     types.StringType,
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
			MarkdownDescription: "Annotation type. Supported: `text`, `rectangle`, `ellipse`, `line`.\n\n" +
				"Coordinate semantics differ by type:\n\n" +
				"- `rectangle`/`ellipse`: `x1`/`y1` define the origin, `x2`/`y2` define width/height.\n" +
				"- `line`: `x1`/`y1` and `x2`/`y2` are two endpoints.\n\n" +
				"Colors are CSS-style hex strings, typically `#RRGGBB` or `#RRGGBBAA` (alpha).",
			Required: true,
		},
		"text": schema.SingleNestedAttribute{
			Description: "Text annotation attributes (required when type = \"text\").",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"text_content": schema.StringAttribute{
					Required: true,
				},
				"x1": schema.Float64Attribute{
					Description: "Text anchor X in canvas coordinates.",
					Required:    true,
				},
				"y1": schema.Float64Attribute{
					Description: "Text anchor Y in canvas coordinates.",
					Required:    true,
				},
				"rotation": schema.Float64Attribute{
					Description: "Text rotation in degrees.",
					Optional:    true,
					Computed:    true,
					Default:     float64default.StaticFloat64(0),
				},
				"color": schema.StringAttribute{
					Description: "Text color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`.",
					Optional:    true,
					Computed:    true,
				},
				"border_color": schema.StringAttribute{
					Description: "Border color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`.",
					Optional:    true,
					Computed:    true,
				},
				"border_style": schema.StringAttribute{
					Description: "Border style as dash pattern. Solid: `\"\"`. Allowed values observed: `\"\"`, `\"2,2\"`, `\"4,2\"`.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Validators: []validator.String{
						stringvalidator.OneOf("", "2,2", "4,2"),
					},
				},
				"thickness": schema.Float64Attribute{
					Optional: true,
					Computed: true,
				},
				"z_index": schema.Float64Attribute{
					Optional: true,
					Computed: true,
				},
				"text_bold": schema.BoolAttribute{
					Description: "Whether text is bold.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
				},
				"text_font": schema.StringAttribute{
					Description: "Font family name.",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("sans"),
				},
				"text_italic": schema.BoolAttribute{
					Description: "Whether text is italic.",
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
				},
				"text_size": schema.Float64Attribute{
					Description: "Text size.",
					Optional:    true,
					Computed:    true,
					Default:     float64default.StaticFloat64(12),
				},
				"text_unit": schema.StringAttribute{
					Description: "Text size unit (e.g. `px`, `pt`).",
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("px"),
				},
			},
		},
		"rectangle": schema.SingleNestedAttribute{
			Description: "Rectangle annotation attributes (required when type = \"rectangle\"). Note: x2/y2 are width/height (not a second point).",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1": schema.Float64Attribute{
					Description: "Rectangle origin X in canvas coordinates.",
					Required:    true,
				},
				"y1": schema.Float64Attribute{
					Description: "Rectangle origin Y in canvas coordinates.",
					Required:    true,
				},
				"x2": schema.Float64Attribute{
					Description: "Rectangle width (not an X coordinate of a second point).",
					Required:    true,
				},
				"y2": schema.Float64Attribute{
					Description: "Rectangle height (not a Y coordinate of a second point).",
					Required:    true,
				},
				"rotation": schema.Float64Attribute{
					Description: "Rectangle rotation in degrees.",
					Optional:    true,
					Computed:    true,
					Default:     float64default.StaticFloat64(0),
				},
				"color":         schema.StringAttribute{Optional: true, Computed: true, Description: "Fill color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_color":  schema.StringAttribute{Optional: true, Computed: true, Description: "Border color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_style":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Border style as dash pattern. Solid: `\"\"`. Allowed values observed: `\"\"`, `\"2,2\"`, `\"4,2\"`.", Validators: []validator.String{stringvalidator.OneOf("", "2,2", "4,2")}},
				"thickness":     schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":       schema.Float64Attribute{Optional: true, Computed: true},
				"border_radius": schema.Float64Attribute{Optional: true, Computed: true, Default: float64default.StaticFloat64(0), Description: "Border radius."},
			},
		},
		"ellipse": schema.SingleNestedAttribute{
			Description: "Ellipse annotation attributes (required when type = \"ellipse\"). Note: x2/y2 are width/height (not a second point).",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1": schema.Float64Attribute{
					Description: "Ellipse origin X in canvas coordinates.",
					Required:    true,
				},
				"y1": schema.Float64Attribute{
					Description: "Ellipse origin Y in canvas coordinates.",
					Required:    true,
				},
				"x2": schema.Float64Attribute{
					Description: "Ellipse width (not an X coordinate of a second point).",
					Required:    true,
				},
				"y2": schema.Float64Attribute{
					Description: "Ellipse height (not a Y coordinate of a second point).",
					Required:    true,
				},
				"rotation":     schema.Float64Attribute{Optional: true, Computed: true, Default: float64default.StaticFloat64(0), Description: "Ellipse rotation in degrees."},
				"color":        schema.StringAttribute{Optional: true, Computed: true, Description: "Fill color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_color": schema.StringAttribute{Optional: true, Computed: true, Description: "Border color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_style": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Border style as dash pattern. Solid: `\"\"`. Allowed values observed: `\"\"`, `\"2,2\"`, `\"4,2\"`.", Validators: []validator.String{stringvalidator.OneOf("", "2,2", "4,2")}},
				"thickness":    schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":      schema.Float64Attribute{Optional: true, Computed: true},
			},
		},
		"line": schema.SingleNestedAttribute{
			Description: "Line annotation attributes (required when type = \"line\"). Note: x1/y1 and x2/y2 are two endpoints.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"x1": schema.Float64Attribute{
					Description: "Line start point X in canvas coordinates.",
					Required:    true,
				},
				"y1": schema.Float64Attribute{
					Description: "Line start point Y in canvas coordinates.",
					Required:    true,
				},
				"x2": schema.Float64Attribute{
					Description: "Line end point X in canvas coordinates.",
					Required:    true,
				},
				"y2": schema.Float64Attribute{
					Description: "Line end point Y in canvas coordinates.",
					Required:    true,
				},
				"color":        schema.StringAttribute{Optional: true, Computed: true, Description: "Line color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_color": schema.StringAttribute{Optional: true, Computed: true, Description: "Border color as CSS hex, e.g. `#RRGGBB` or `#RRGGBBAA`."},
				"border_style": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Border style as dash pattern. Solid: `\"\"`. Allowed values observed: `\"\"`, `\"2,2\"`, `\"4,2\"`.", Validators: []validator.String{stringvalidator.OneOf("", "2,2", "4,2")}},
				"thickness":    schema.Float64Attribute{Optional: true, Computed: true},
				"z_index":      schema.Float64Attribute{Optional: true, Computed: true},
				"line_start":   schema.StringAttribute{Optional: true, Description: "Marker/style at the start point (x1,y1). Allowed: `arrow`, `square`, `circle`, or `null`.", Validators: []validator.String{stringvalidator.OneOf("arrow", "square", "circle")}},
				"line_end":     schema.StringAttribute{Optional: true, Description: "Marker/style at the end point (x2,y2). Allowed: `arrow`, `square`, `circle`, or `null`.", Validators: []validator.String{stringvalidator.OneOf("arrow", "square", "circle")}},
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
				Rotation:    types.Float64Value(a.Text.Rotation),
				Color:       types.StringValue(a.Text.Color),
				BorderColor: types.StringValue(a.Text.BorderColor),
				BorderStyle: types.StringValue(string(a.Text.BorderStyle)),
				Thickness:   types.Float64Value(a.Text.Thickness),
				ZIndex:      types.Float64Value(a.Text.ZIndex),
				TextBold:    types.BoolValue(a.Text.TextBold),
				TextFont:    types.StringValue(a.Text.TextFont),
				TextItalic:  types.BoolValue(a.Text.TextItalic),
				TextSize:    types.Float64Value(a.Text.TextSize),
				TextUnit:    types.StringValue(a.Text.TextUnit),
			}
			model.ID = types.StringValue(string(a.Text.ID))
			var v attr.Value
			diags.Append(tfsdk.ValueFrom(ctx, text, types.ObjectType{AttrTypes: AnnotationTextAttrType}, &v)...)
			model.Text = v.(types.Object)
		}
	case models.AnnotationTypeRectangle:
		if a.Rectangle != nil {
			rec := AnnotationRectangleModel{
				X1:           types.Float64Value(a.Rectangle.X1),
				Y1:           types.Float64Value(a.Rectangle.Y1),
				X2:           types.Float64Value(a.Rectangle.X2),
				Y2:           types.Float64Value(a.Rectangle.Y2),
				Rotation:     types.Float64Value(a.Rectangle.Rotation),
				Color:        types.StringValue(a.Rectangle.Color),
				BorderColor:  types.StringValue(a.Rectangle.BorderColor),
				BorderStyle:  types.StringValue(string(a.Rectangle.BorderStyle)),
				Thickness:    types.Float64Value(a.Rectangle.Thickness),
				ZIndex:       types.Float64Value(a.Rectangle.ZIndex),
				BorderRadius: types.Float64Value(a.Rectangle.BorderRadius),
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
				Rotation:    types.Float64Value(a.Ellipse.Rotation),
				Color:       types.StringValue(a.Ellipse.Color),
				BorderColor: types.StringValue(a.Ellipse.BorderColor),
				BorderStyle: types.StringValue(string(a.Ellipse.BorderStyle)),
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
			lineStart := types.StringNull()
			if a.Line.LineStart != nil {
				lineStart = types.StringValue(string(*a.Line.LineStart))
			}
			lineEnd := types.StringNull()
			if a.Line.LineEnd != nil {
				lineEnd = types.StringValue(string(*a.Line.LineEnd))
			}
			ln := AnnotationLineModel{
				X1:          types.Float64Value(a.Line.X1),
				Y1:          types.Float64Value(a.Line.Y1),
				X2:          types.Float64Value(a.Line.X2),
				Y2:          types.Float64Value(a.Line.Y2),
				Color:       types.StringValue(a.Line.Color),
				BorderColor: types.StringValue(a.Line.BorderColor),
				BorderStyle: types.StringValue(string(a.Line.BorderStyle)),
				Thickness:   types.Float64Value(a.Line.Thickness),
				ZIndex:      types.Float64Value(a.Line.ZIndex),
				LineStart:   lineStart,
				LineEnd:     lineEnd,
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
