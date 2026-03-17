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

	if !data.Groups.IsUnknown() && !data.Groups.IsNull() {
		groups := make([]models.LabGroup, 0)
		var g cmlschema.LabGroupModel
		for _, elem := range data.Groups.Elements() {
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, elem, &g)...)
			if resp.Diagnostics.HasError() {
				return
			}
			perm := models.OldPermissionReadOnly
			if g.Permission.ValueString() == string(models.OldPermissionReadWrite) {
				perm = models.OldPermissionReadWrite
			}
			groups = append(groups, models.LabGroup{ID: models.UUID(g.ID.ValueString()), Permission: perm})
		}
		updateReq.Groups = groups
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
