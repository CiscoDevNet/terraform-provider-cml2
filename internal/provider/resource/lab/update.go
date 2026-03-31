package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Update updates an existing CML lab.
func (r LabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		data cmlschema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	managedNodeStaging := data.NodeStaging

	updateReq := models.LabUpdateRequest{
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
		Notes:       data.Notes.ValueString(),
	}

	if ns := expandNodeStaging(ctx, data.NodeStaging, &resp.Diagnostics); ns != nil {
		updateReq.NodeStaging = ns
	}

	updateReq.Associations = expandGroupAssociations(ctx, data.Groups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newLab, err := r.cfg.Client().Lab.Update(ctx, models.UUID(data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update lab, got error: %s", err),
		)
		return
	}

	// Refresh to get populated groups from API.
	fullLab, err := r.cfg.Client().Lab.GetByID(ctx, newLab.ID, false)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get lab, got error: %s", err))
		return
	}
	if err := r.hydrateGroups(ctx, &fullLab); err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get lab groups, got error: %s", err))
		return
	}

	value := cmlschema.NewLab(ctx, &fullLab, &resp.Diagnostics)
	var newData cmlschema.LabModel
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, value, &newData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	keepNodeStagingNullWhenUnmanaged(managedNodeStaging, &newData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)

	tflog.Info(ctx, "Resource Lab UPDATE done")
}
