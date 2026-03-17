package annotation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Update updates an existing classic annotation.
func (r *AnnotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cmlschema.AnnotationModel

	tflog.Info(ctx, "Resource Annotation UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateAnnotationBlocks(plan); err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, err.Error())
		return
	}

	upd, err := buildAnnotationUpdate(ctx, plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, err.Error())
		return
	}

	labID := models.UUID(plan.LabID.ValueString())
	annID := models.UUID(plan.ID.ValueString())

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
