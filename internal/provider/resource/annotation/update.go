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

func (r *AnnotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cmlschema.AnnotationModel

	tflog.Info(ctx, "Resource Annotation UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Type.ValueString() != string(models.AnnotationTypeText) {
		resp.Diagnostics.AddError(common.ErrorLabel, "unsupported annotation type (currently only \"text\" is supported)")
		return
	}
	if plan.Text.IsNull() {
		resp.Diagnostics.AddError(common.ErrorLabel, "text block must be set when type = \"text\"")
		return
	}

	var text cmlschema.AnnotationTextModel
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, plan.Text, &text)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labID := models.UUID(plan.LabID.ValueString())
	annID := models.UUID(plan.ID.ValueString())

	content := text.TextContent.ValueString()
	x1 := text.X1.ValueFloat64()
	y1 := text.Y1.ValueFloat64()
	upd := models.AnnotationUpdate{
		Type: models.AnnotationTypeText,
		Text: &models.TextAnnotationPartial{
			Type:        models.AnnotationTypeText,
			TextContent: &content,
			X1:          &x1,
			Y1:          &y1,
		},
	}

	updated, err := r.cfg.Client().Annotation.Update(ctx, labID, annID, upd)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("unable to update annotation: %s", err))
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(ctx, cmlschema.NewAnnotation(ctx, labID, updated, &resp.Diagnostics), types.ObjectType{AttrTypes: cmlschema.AnnotationAttrType}, &plan)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Info(ctx, "Resource Annotation UPDATE done")
}
