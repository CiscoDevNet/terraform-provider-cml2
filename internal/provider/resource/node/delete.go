package node

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

func (r *NodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var (
		data cmlschema.NodeModel
		err  error
	)

	tflog.Info(ctx, "Resource Node DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := &cmlclient.Node{
		ID:    data.ID.ValueString(),
		LabID: data.LabID.ValueString(),
	}

	err = r.cfg.Client().NodeStop(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to stop node, got error: %s", err),
		)
		return
	}

	err = r.cfg.Client().NodeWipe(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to wipe node, got error: %s", err),
		)
		return
	}

	err = r.cfg.Client().NodeDestroy(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy node, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Node DELETE done")
}
