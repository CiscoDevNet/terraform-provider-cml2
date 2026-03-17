package annotation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *AnnotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cmlschema.AnnotationModel

	tflog.Info(ctx, "Resource Annotation DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labID := models.UUID(data.LabID.ValueString())
	annID := models.UUID(data.ID.ValueString())

	if err := r.cfg.Client().Annotation.Delete(ctx, labID, annID); err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("unable to delete annotation: %s", err))
		return
	}

	tflog.Info(ctx, "Resource Annotation DELETE done")
}
