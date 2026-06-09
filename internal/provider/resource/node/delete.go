package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlerrors "github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Delete deletes an existing node from a CML lab.
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

	labID := models.UUID(data.LabID.ValueString())
	nodeID := models.UUID(data.ID.ValueString())

	err = r.cfg.Client().Node.Stop(ctx, labID, nodeID)
	if err != nil {
		if errors.Is(err, cmlerrors.ErrElementNotFound) || errors.Is(err, cmlerrors.ErrAPINotFound) {
			// Node already gone (deleted externally). Treat as successful cleanup.
			return
		}
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to stop node, got error: %s", err),
		)
		return
	}

	err = r.cfg.Client().Node.Wipe(ctx, labID, nodeID)
	if err != nil {
		if errors.Is(err, cmlerrors.ErrElementNotFound) || errors.Is(err, cmlerrors.ErrAPINotFound) {
			return
		}
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to wipe node, got error: %s", err),
		)
		return
	}

	err = r.cfg.Client().Node.Delete(ctx, labID, nodeID)
	if err != nil {
		if errors.Is(err, cmlerrors.ErrElementNotFound) || errors.Is(err, cmlerrors.ErrAPINotFound) {
			return
		}
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy node, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Node DELETE done")
}
