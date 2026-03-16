package cmlschema_test

import (
	"context"
	"testing"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
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
