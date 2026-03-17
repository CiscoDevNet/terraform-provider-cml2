package link

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *LinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// this is a no-op at this point as links are removed automatically
	// when nodes and their interfaces are deleted

	var (
		data cmlschema.LinkModel
		err  error
	)

	tflog.Info(ctx, "Resource Link DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = r.cfg.Client().Link.Delete(ctx, models.UUID(data.LabID.ValueString()), models.UUID(data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy link, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Link DELETE done")
}
