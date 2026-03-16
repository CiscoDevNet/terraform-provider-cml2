package annotation

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func clampThickness(v float64) float64 {
	if v < 1 {
		return 1
	}
	return v
}

func buildAnnotationCreate(ctx context.Context, data cmlschema.AnnotationModel, diags *diag.Diagnostics) (models.AnnotationCreate, error) {
	typeStr := data.Type.ValueString()
	switch typeStr {
	case string(models.AnnotationTypeText):
		if data.Text.IsNull() {
			return models.AnnotationCreate{}, fmt.Errorf("text block must be set when type=\"text\"")
		}
		var text cmlschema.AnnotationTextModel
		diags.Append(tfsdk.ValueAs(ctx, data.Text, &text)...)
		if diags.HasError() {
			return models.AnnotationCreate{}, fmt.Errorf("invalid text block")
		}
		borderColor := "#000000"
		if !text.BorderColor.IsNull() {
			borderColor = text.BorderColor.ValueString()
		}
		color := "#ffffff"
		if !text.Color.IsNull() {
			color = text.Color.ValueString()
		}
		thickness := 1.0
		if !text.Thickness.IsNull() {
			thickness = clampThickness(text.Thickness.ValueFloat64())
		}
		z := 0.0
		if !text.ZIndex.IsNull() {
			z = text.ZIndex.ValueFloat64()
		}
		return models.AnnotationCreate{
			Type: models.AnnotationTypeText,
			Text: &models.TextAnnotation{
				Type:        models.AnnotationTypeText,
				Rotation:    0,
				BorderColor: borderColor,
				BorderStyle: "",
				Color:       color,
				Thickness:   thickness,
				X1:          text.X1.ValueFloat64(),
				Y1:          text.Y1.ValueFloat64(),
				ZIndex:      z,
				TextBold:    false,
				TextContent: text.TextContent.ValueString(),
				TextFont:    "sans",
				TextItalic:  false,
				TextSize:    12,
				TextUnit:    "px",
			},
		}, nil
	case string(models.AnnotationTypeRectangle):
		if data.Rectangle.IsNull() {
			return models.AnnotationCreate{}, fmt.Errorf("rectangle block must be set when type=\"rectangle\"")
		}
		var rec cmlschema.AnnotationRectangleModel
		diags.Append(tfsdk.ValueAs(ctx, data.Rectangle, &rec)...)
		if diags.HasError() {
			return models.AnnotationCreate{}, fmt.Errorf("invalid rectangle block")
		}
		borderColor := "#000000"
		if !rec.BorderColor.IsNull() {
			borderColor = rec.BorderColor.ValueString()
		}
		color := "#ffffff"
		if !rec.Color.IsNull() {
			color = rec.Color.ValueString()
		}
		thickness := 1.0
		if !rec.Thickness.IsNull() {
			thickness = clampThickness(rec.Thickness.ValueFloat64())
		}
		z := 0.0
		if !rec.ZIndex.IsNull() {
			z = rec.ZIndex.ValueFloat64()
		}
		return models.AnnotationCreate{
			Type: models.AnnotationTypeRectangle,
			Rectangle: &models.RectangleAnnotation{
				Type:         models.AnnotationTypeRectangle,
				Rotation:     0,
				BorderColor:  borderColor,
				BorderStyle:  "",
				Color:        color,
				Thickness:    thickness,
				X1:           rec.X1.ValueFloat64(),
				Y1:           rec.Y1.ValueFloat64(),
				X2:           rec.X2.ValueFloat64(),
				Y2:           rec.Y2.ValueFloat64(),
				ZIndex:       z,
				BorderRadius: 0,
			},
		}, nil
	case string(models.AnnotationTypeEllipse):
		if data.Ellipse.IsNull() {
			return models.AnnotationCreate{}, fmt.Errorf("ellipse block must be set when type=\"ellipse\"")
		}
		var el cmlschema.AnnotationEllipseModel
		diags.Append(tfsdk.ValueAs(ctx, data.Ellipse, &el)...)
		if diags.HasError() {
			return models.AnnotationCreate{}, fmt.Errorf("invalid ellipse block")
		}
		borderColor := "#000000"
		if !el.BorderColor.IsNull() {
			borderColor = el.BorderColor.ValueString()
		}
		color := "#ffffff"
		if !el.Color.IsNull() {
			color = el.Color.ValueString()
		}
		thickness := 1.0
		if !el.Thickness.IsNull() {
			thickness = clampThickness(el.Thickness.ValueFloat64())
		}
		z := 0.0
		if !el.ZIndex.IsNull() {
			z = el.ZIndex.ValueFloat64()
		}
		return models.AnnotationCreate{
			Type: models.AnnotationTypeEllipse,
			Ellipse: &models.EllipseAnnotation{
				Type:        models.AnnotationTypeEllipse,
				Rotation:    0,
				BorderColor: borderColor,
				BorderStyle: "",
				Color:       color,
				Thickness:   thickness,
				X1:          el.X1.ValueFloat64(),
				Y1:          el.Y1.ValueFloat64(),
				X2:          el.X2.ValueFloat64(),
				Y2:          el.Y2.ValueFloat64(),
				ZIndex:      z,
			},
		}, nil
	case string(models.AnnotationTypeLine):
		if data.Line.IsNull() {
			return models.AnnotationCreate{}, fmt.Errorf("line block must be set when type=\"line\"")
		}
		var ln cmlschema.AnnotationLineModel
		diags.Append(tfsdk.ValueAs(ctx, data.Line, &ln)...)
		if diags.HasError() {
			return models.AnnotationCreate{}, fmt.Errorf("invalid line block")
		}
		color := "#ffffff"
		if !ln.Color.IsNull() {
			color = ln.Color.ValueString()
		}
		thickness := 1.0
		if !ln.Thickness.IsNull() {
			thickness = clampThickness(ln.Thickness.ValueFloat64())
		}
		z := 0.0
		if !ln.ZIndex.IsNull() {
			z = ln.ZIndex.ValueFloat64()
		}
		if ln.LineStart.IsNull() || ln.LineEnd.IsNull() {
			return models.AnnotationCreate{}, fmt.Errorf("line_start and line_end must be set when type=\"line\" (allowed: arrow, square, circle)")
		}
		start := models.LineEnd(ln.LineStart.ValueString())
		end := models.LineEnd(ln.LineEnd.ValueString())
		borderColor := "#000000"
		return models.AnnotationCreate{
			Type: models.AnnotationTypeLine,
			Line: &models.LineAnnotation{
				Type:        models.AnnotationTypeLine,
				BorderColor: borderColor,
				BorderStyle: "",
				Color:       color,
				Thickness:   thickness,
				X1:          ln.X1.ValueFloat64(),
				Y1:          ln.Y1.ValueFloat64(),
				X2:          ln.X2.ValueFloat64(),
				Y2:          ln.Y2.ValueFloat64(),
				ZIndex:      z,
				LineStart:   start,
				LineEnd:     end,
			},
		}, nil
	default:
		return models.AnnotationCreate{}, fmt.Errorf("unsupported annotation type %q", typeStr)
	}
}

func buildAnnotationUpdate(ctx context.Context, data cmlschema.AnnotationModel, diags *diag.Diagnostics) (models.AnnotationUpdate, error) {
	typeStr := data.Type.ValueString()
	switch typeStr {
	case string(models.AnnotationTypeText):
		if data.Text.IsNull() {
			return models.AnnotationUpdate{}, fmt.Errorf("text block must be set when type=\"text\"")
		}
		var text cmlschema.AnnotationTextModel
		diags.Append(tfsdk.ValueAs(ctx, data.Text, &text)...)
		if diags.HasError() {
			return models.AnnotationUpdate{}, fmt.Errorf("invalid text block")
		}
		content := text.TextContent.ValueString()
		x1 := text.X1.ValueFloat64()
		y1 := text.Y1.ValueFloat64()
		upd := models.AnnotationUpdate{Type: models.AnnotationTypeText, Text: &models.TextAnnotationPartial{Type: models.AnnotationTypeText, TextContent: &content, X1: &x1, Y1: &y1}}
		return upd, nil
	case string(models.AnnotationTypeRectangle):
		if data.Rectangle.IsNull() {
			return models.AnnotationUpdate{}, fmt.Errorf("rectangle block must be set when type=\"rectangle\"")
		}
		var rec cmlschema.AnnotationRectangleModel
		diags.Append(tfsdk.ValueAs(ctx, data.Rectangle, &rec)...)
		if diags.HasError() {
			return models.AnnotationUpdate{}, fmt.Errorf("invalid rectangle block")
		}
		x1 := rec.X1.ValueFloat64()
		y1 := rec.Y1.ValueFloat64()
		x2 := rec.X2.ValueFloat64()
		y2 := rec.Y2.ValueFloat64()
		upd := models.AnnotationUpdate{Type: models.AnnotationTypeRectangle, Rectangle: &models.RectangleAnnotationPartial{Type: models.AnnotationTypeRectangle, X1: &x1, Y1: &y1, X2: &x2, Y2: &y2}}
		return upd, nil
	case string(models.AnnotationTypeEllipse):
		if data.Ellipse.IsNull() {
			return models.AnnotationUpdate{}, fmt.Errorf("ellipse block must be set when type=\"ellipse\"")
		}
		var el cmlschema.AnnotationEllipseModel
		diags.Append(tfsdk.ValueAs(ctx, data.Ellipse, &el)...)
		if diags.HasError() {
			return models.AnnotationUpdate{}, fmt.Errorf("invalid ellipse block")
		}
		x1 := el.X1.ValueFloat64()
		y1 := el.Y1.ValueFloat64()
		x2 := el.X2.ValueFloat64()
		y2 := el.Y2.ValueFloat64()
		upd := models.AnnotationUpdate{Type: models.AnnotationTypeEllipse, Ellipse: &models.EllipseAnnotationPartial{Type: models.AnnotationTypeEllipse, X1: &x1, Y1: &y1, X2: &x2, Y2: &y2}}
		return upd, nil
	case string(models.AnnotationTypeLine):
		if data.Line.IsNull() {
			return models.AnnotationUpdate{}, fmt.Errorf("line block must be set when type=\"line\"")
		}
		var ln cmlschema.AnnotationLineModel
		diags.Append(tfsdk.ValueAs(ctx, data.Line, &ln)...)
		if diags.HasError() {
			return models.AnnotationUpdate{}, fmt.Errorf("invalid line block")
		}
		x1 := ln.X1.ValueFloat64()
		y1 := ln.Y1.ValueFloat64()
		x2 := ln.X2.ValueFloat64()
		y2 := ln.Y2.ValueFloat64()
		upd := models.AnnotationUpdate{Type: models.AnnotationTypeLine, Line: &models.LineAnnotationPartial{Type: models.AnnotationTypeLine, X1: &x1, Y1: &y1, X2: &x2, Y2: &y2}}
		return upd, nil
	default:
		return models.AnnotationUpdate{}, fmt.Errorf("unsupported annotation type %q", typeStr)
	}
}

func validateAnnotationBlocks(data cmlschema.AnnotationModel) error {
	set := 0
	if !data.Text.IsNull() {
		set++
	}
	if !data.Rectangle.IsNull() {
		set++
	}
	if !data.Ellipse.IsNull() {
		set++
	}
	if !data.Line.IsNull() {
		set++
	}
	if set != 1 {
		return fmt.Errorf("exactly one of text/rectangle/ellipse/line must be set")
	}

	typeStr := data.Type.ValueString()
	switch typeStr {
	case string(models.AnnotationTypeText):
		if data.Text.IsNull() {
			return fmt.Errorf("text block must be set when type=\"text\"")
		}
	case string(models.AnnotationTypeRectangle):
		if data.Rectangle.IsNull() {
			return fmt.Errorf("rectangle block must be set when type=\"rectangle\"")
		}
	case string(models.AnnotationTypeEllipse):
		if data.Ellipse.IsNull() {
			return fmt.Errorf("ellipse block must be set when type=\"ellipse\"")
		}
	case string(models.AnnotationTypeLine):
		if data.Line.IsNull() {
			return fmt.Errorf("line block must be set when type=\"line\"")
		}
	default:
		return fmt.Errorf("unsupported annotation type %q", typeStr)
	}

	return nil
}
