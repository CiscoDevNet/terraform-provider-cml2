package link

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
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

	link := &cmlclient.Link{
		ID:    data.ID.ValueString(),
		LabID: data.LabID.ValueString(),
	}

	err = r.cfg.Client().LinkDestroy(ctx, link)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy link, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Link DELETE done")
}
