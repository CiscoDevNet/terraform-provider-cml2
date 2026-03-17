package lab

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func (r *LabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.LabModel

	tflog.Info(ctx, "Resource Lab READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	managedNodeStaging := data.NodeStaging
	// Heuristic: during import, state typically only has ID and most computed attrs are unknown.
	// In that case we must not suppress node_staging, otherwise ImportStateVerify will fail.
	isImportRead := data.Created.IsUnknown() && data.Modified.IsUnknown()

	lab, err := r.cfg.Client().Lab.GetByID(ctx, models.UUID(data.ID.ValueString()), false)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	// Save data into Terraform state
	value := cmlschema.NewLab(ctx, &lab, &resp.Diagnostics)
	var newData cmlschema.LabModel
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, value, &newData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !isImportRead {
		keepNodeStagingNullWhenUnmanaged(managedNodeStaging, &newData)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)

	tflog.Info(ctx, "Resource Lab READ done")
}
