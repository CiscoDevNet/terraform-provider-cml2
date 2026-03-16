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

	if err := validateAnnotationBlocks(data); err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, err.Error())
		return
	}

	create, err := buildAnnotationCreate(ctx, data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, err.Error())
		return
	}

	labID := models.UUID(data.LabID.ValueString())

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
