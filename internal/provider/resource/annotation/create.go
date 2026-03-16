package annotation

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func (r *AnnotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data cmlschema.AnnotationModel

	tflog.Info(ctx, "Resource Annotation CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() != string(models.AnnotationTypeText) {
		resp.Diagnostics.AddError(common.ErrorLabel, "unsupported annotation type (currently only \"text\" is supported)")
		return
	}
	if data.Text.IsNull() {
		resp.Diagnostics.AddError(common.ErrorLabel, "text block must be set when type = \"text\"")
		return
	}

	var text cmlschema.AnnotationTextModel
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, data.Text, &text)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labID := models.UUID(data.LabID.ValueString())
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
		thickness = text.Thickness.ValueFloat64()
		if thickness < 1 {
			thickness = 1
		}
	}
	z := 0.0
	if !text.ZIndex.IsNull() {
		z = text.ZIndex.ValueFloat64()
	}

	create := models.AnnotationCreate{
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
	}

	created, err := r.cfg.Client().Annotation.Create(ctx, labID, create)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("unable to create annotation: %s", err))
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(ctx, cmlschema.NewAnnotation(ctx, labID, created, &resp.Diagnostics), types.ObjectType{AttrTypes: cmlschema.AnnotationAttrType}, &data)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Annotation CREATE done")
}
