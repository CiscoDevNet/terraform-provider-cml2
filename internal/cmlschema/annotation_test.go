package cmlschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

func TestNewAnnotation_Text(t *testing.T) {
	ctx := context.Background()
	d := &diag.Diagnostics{}

	labID := models.UUID("lab-1")
	id := models.UUID("a-1")
	a := models.Annotation{Type: models.AnnotationTypeText, Text: &models.TextAnnotationResponse{ID: id, TextAnnotation: models.TextAnnotation{Type: models.AnnotationTypeText, TextContent: "hi", X1: 1, Y1: 2, Color: "#fff", BorderColor: "#000", Thickness: 1, ZIndex: 0}}}

	v := cmlschema.NewAnnotation(ctx, labID, a, d)
	assert.False(t, d.HasError())

	var out cmlschema.AnnotationModel
	d.Append(tfsdk.ValueAs(ctx, v, &out)...)
	assert.False(t, d.HasError())
	assert.Equal(t, "lab-1", out.LabID.ValueString())
	assert.Equal(t, "a-1", out.ID.ValueString())
	assert.Equal(t, string(models.AnnotationTypeText), out.Type.ValueString())
	assert.False(t, out.Text.IsNull())
}

func TestNewAnnotation_Rectangle(t *testing.T) {
	ctx := context.Background()
	d := &diag.Diagnostics{}

	labID := models.UUID("lab-1")
	id := models.UUID("a-2")
	a := models.Annotation{Type: models.AnnotationTypeRectangle, Rectangle: &models.RectangleAnnotationResponse{ID: id, RectangleAnnotation: models.RectangleAnnotation{Type: models.AnnotationTypeRectangle, X1: 1, Y1: 2, X2: 3, Y2: 4, Color: "#fff", BorderColor: "#000", Thickness: 1, ZIndex: 0}}}

	v := cmlschema.NewAnnotation(ctx, labID, a, d)
	assert.False(t, d.HasError())

	var out cmlschema.AnnotationModel
	d.Append(tfsdk.ValueAs(ctx, v, &out)...)
	assert.False(t, d.HasError())
	assert.Equal(t, string(models.AnnotationTypeRectangle), out.Type.ValueString())
	assert.False(t, out.Rectangle.IsNull())
}

func TestNewAnnotation_Ellipse(t *testing.T) {
	ctx := context.Background()
	d := &diag.Diagnostics{}

	labID := models.UUID("lab-1")
	id := models.UUID("a-3")
	a := models.Annotation{Type: models.AnnotationTypeEllipse, Ellipse: &models.EllipseAnnotationResponse{ID: id, EllipseAnnotation: models.EllipseAnnotation{Type: models.AnnotationTypeEllipse, X1: 1, Y1: 2, X2: 3, Y2: 4, Color: "#fff", BorderColor: "#000", Thickness: 1, ZIndex: 0}}}

	v := cmlschema.NewAnnotation(ctx, labID, a, d)
	assert.False(t, d.HasError())

	var out cmlschema.AnnotationModel
	d.Append(tfsdk.ValueAs(ctx, v, &out)...)
	assert.False(t, d.HasError())
	assert.Equal(t, string(models.AnnotationTypeEllipse), out.Type.ValueString())
	assert.False(t, out.Ellipse.IsNull())
}

func TestNewAnnotation_Line(t *testing.T) {
	ctx := context.Background()
	d := &diag.Diagnostics{}

	labID := models.UUID("lab-1")
	id := models.UUID("a-4")
	a := models.Annotation{Type: models.AnnotationTypeLine, Line: &models.LineAnnotationResponse{ID: id, LineAnnotation: models.LineAnnotation{Type: models.AnnotationTypeLine, X1: 1, Y1: 2, X2: 3, Y2: 4, Color: "#fff", BorderColor: "#000", Thickness: 1, ZIndex: 0}}}

	v := cmlschema.NewAnnotation(ctx, labID, a, d)
	assert.False(t, d.HasError())

	var out cmlschema.AnnotationModel
	d.Append(tfsdk.ValueAs(ctx, v, &out)...)
	assert.False(t, d.HasError())
	assert.Equal(t, string(models.AnnotationTypeLine), out.Type.ValueString())
	assert.False(t, out.Line.IsNull())
}
